# Sprint 4 — Step 2: Visit Domain + Vitals + Orders + ICD-10 API (Backend)

## Mục tiêu
Xây dựng module `internal/visit`: domain entities (Visit, VisitVital, VisitOrder), toàn bộ Visit API cho Doctor và Nurse, kèm ICD-10 full-text search.

## Prerequisite
- Step 1 hoàn thành: `QueueEntry` entity và `queue_entries` table sẵn sàng.
- `patients` table và Patient API từ Sprint 3 ✅.

## Files cần tạo
```
backend/
├── internal/visit/
│   ├── domain/
│   │   ├── visit.go           -- Entity Visit + status machine
│   │   ├── vital.go           -- Entity VisitVital
│   │   ├── order.go           -- Entity VisitOrder
│   │   └── repository.go      -- VisitRepository interface
│   ├── application/
│   │   ├── commands/
│   │   │   ├── create_visit.go
│   │   │   ├── update_status.go
│   │   │   ├── record_vitals.go
│   │   │   ├── create_order.go
│   │   │   └── close_visit.go
│   │   └── queries/
│   │       ├── get_worklist.go
│   │       ├── get_visit_detail.go
│   │       └── search_icd10.go
│   ├── infrastructure/
│   │   └── visit_repository_pg.go
│   ├── handlers/
│   │   └── visit_handler.go
│   └── bootstrap/
│       ├── module.go
│       └── router.go
```

## Nhiệm vụ chi tiết

### 1. Domain Layer

**`visit.go`:**
```go
type VisitStatus string
const (
    VisitRegistered VisitStatus = "REGISTERED"
    VisitWaiting    VisitStatus = "WAITING"
    VisitInProgress VisitStatus = "IN_PROGRESS"
    VisitOrdered    VisitStatus = "ORDERED"
    VisitCompleted  VisitStatus = "COMPLETED"
    VisitCancelled  VisitStatus = "CANCELLED"
)

type Visit struct {
    ID             uuid.UUID
    PatientID      uuid.UUID
    DoctorID       uuid.UUID
    QueueEntryID   *uuid.UUID
    Status         VisitStatus
    ChiefComplaint *string
    StartedAt      *time.Time
    CompletedAt    *time.Time
    CreatedAt      time.Time
    UpdatedAt      time.Time
}

func (v *Visit) Start() error
func (v *Visit) Complete() error
func (v *Visit) Cancel() error
```

**`vital.go`:**
```go
type VisitVital struct {
    ID          uuid.UUID
    VisitID     uuid.UUID
    BpSystolic  *int      // huyết áp tâm thu (mmHg)
    BpDiastolic *int      // huyết áp tâm trương (mmHg)
    HeartRate   *int      // mạch (bpm)
    Temperature *float64  // nhiệt độ (°C)
    SpO2        *int      // SpO2 (%)
    WeightKg    *float64  // cân nặng (kg)
    HeightCm    *float64  // chiều cao (cm)
    RecordedAt  time.Time
    RecordedBy  uuid.UUID
}

// Ngưỡng cảnh báo bất thường
func (v *VisitVital) Alerts() []string
```

**`order.go`:**
```go
type OrderType string
const (
    OrderLab       OrderType = "LAB"
    OrderRadiology OrderType = "RADIOLOGY"
    OrderProcedure OrderType = "PROCEDURE"
)

type VisitOrder struct {
    ID        uuid.UUID
    VisitID   uuid.UUID
    OrderType OrderType
    RefID     *uuid.UUID  // link tới lab_order hoặc radiology_order (Sprint 5)
    Details   string      // JSON hoặc text mô tả
    Status    string      // "PENDING", "IN_PROGRESS", "COMPLETED"
    CreatedAt time.Time
}
```

**`repository.go`:**
```go
type VisitRepository interface {
    Save(ctx, *Visit) error
    FindByID(ctx, uuid.UUID) (*Visit, error)
    FindWorklist(ctx, doctorID uuid.UUID, date time.Time, status VisitStatus) ([]*Visit, error)
    UpdateStatus(ctx, id uuid.UUID, status VisitStatus) error
    
    SaveVital(ctx, *VisitVital) error
    FindVitalsByVisitID(ctx, visitID uuid.UUID) ([]*VisitVital, error)
    
    SaveOrder(ctx, *VisitOrder) error
    FindOrdersByVisitID(ctx, visitID uuid.UUID) ([]*VisitOrder, error)
    
    SearchICD10(ctx, query string, limit int) ([]*ICD10Code, error)
}

type ICD10Code struct {
    Code          string
    DescriptionVI string
    Category      string
}
```

### 2. Application Layer

**Commands:**
- `CreateVisitCommand{PatientID, DoctorID, QueueEntryID?}` → tạo Visit → publish `HIS.VISIT.VisitStarted`
- `UpdateVisitStatusCommand{VisitID, NewStatus}` → validate transition → update
- `RecordVitalsCommand{VisitID, ...vitals}` → tạo `VisitVital`
- `CreateVisitOrderCommand{VisitID, OrderType, Details}` → tạo `VisitOrder` → publish `HIS.VISIT.LabOrderCreated`
- `CloseVisitCommand{VisitID}` → status = COMPLETED → publish `HIS.VISIT.VisitClosed`

**Queries:**
- `GetDoctorWorklist{DoctorID, Date, Status?}` → trả danh sách visits + patient info summary
- `GetVisitDetail{VisitID}` → trả full visit detail (vitals, orders, patient info)
- `SearchICD10{Query}` → PostgreSQL full-text search trên `icd10_codes` table

### 3. Infrastructure Layer

**`visit_repository_pg.go`:**
- Full implementation, join với `patients` để lấy tên bệnh nhân trong worklist
- ICD-10 search dùng `to_tsquery` với `icd10_codes.description_tsv` index

### 4. Handlers

**`visit_handler.go`:**
```
GET  /api/v1/visits              [DOCTOR, NURSE]   → GetDoctorWorklist
POST /api/v1/visits              [RECEPTIONIST]    → CreateVisitCommand
GET  /api/v1/visits/:id          [DOCTOR, NURSE, LAB_TECH] → GetVisitDetail
PUT  /api/v1/visits/:id/status   [DOCTOR, NURSE]   → UpdateVisitStatusCommand
POST /api/v1/visits/:id/vitals   [DOCTOR, NURSE]   → RecordVitalsCommand
GET  /api/v1/visits/:id/vitals   [DOCTOR, NURSE]   → FindVitalsByVisitID
POST /api/v1/visits/:id/orders   [DOCTOR]          → CreateVisitOrderCommand
GET  /api/v1/visits/:id/orders   [DOCTOR, LAB_TECH, PHARMACIST]
POST /api/v1/visits/:id/close    [DOCTOR]          → CloseVisitCommand
GET  /api/v1/icd10/search        [DOCTOR]          → SearchICD10
```

### 5. Redis Stream Events
Dùng lại Redis client (native driver từ Sprint 2/3 ✅):
- `HIS.VISIT.VisitStarted` → payload: `{visit_id, patient_id, doctor_id, started_at}`
- `HIS.VISIT.LabOrderCreated` → payload: `{order_id, visit_id, order_type, details}`
- `HIS.VISIT.VisitClosed` → payload: `{visit_id, patient_id, completed_at}` → billing (Sprint 7)

## Database Migrations

```sql
-- visits table
CREATE TABLE visits (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    patient_id       UUID NOT NULL REFERENCES patients(id),
    doctor_id        UUID NOT NULL REFERENCES users(id),
    queue_entry_id   UUID REFERENCES queue_entries(id),
    status           VARCHAR(20) NOT NULL DEFAULT 'REGISTERED',
    chief_complaint  TEXT,
    started_at       TIMESTAMPTZ,
    completed_at     TIMESTAMPTZ,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_visits_worklist ON visits(doctor_id, (created_at::date), status);

-- visit_vitals table
CREATE TABLE visit_vitals (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    visit_id      UUID NOT NULL REFERENCES visits(id) ON DELETE CASCADE,
    bp_systolic   INT,
    bp_diastolic  INT,
    heart_rate    INT,
    temperature   DECIMAL(4,1),
    spo2          INT,
    weight_kg     DECIMAL(5,2),
    height_cm     DECIMAL(5,1),
    recorded_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    recorded_by   UUID NOT NULL REFERENCES users(id)
);

-- visit_orders table
CREATE TABLE visit_orders (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    visit_id    UUID NOT NULL REFERENCES visits(id) ON DELETE CASCADE,
    order_type  VARCHAR(20) NOT NULL,
    ref_id      UUID,
    details     TEXT,
    status      VARCHAR(20) NOT NULL DEFAULT 'PENDING',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- icd10_codes table (seed data sau)
CREATE TABLE icd10_codes (
    code            VARCHAR(10) PRIMARY KEY,
    description_vi  TEXT NOT NULL,
    category        VARCHAR(100),
    description_tsv TSVECTOR GENERATED ALWAYS AS (
        to_tsvector('simple', code || ' ' || description_vi)
    ) STORED
);
CREATE INDEX idx_icd10_fts ON icd10_codes USING GIN(description_tsv);
```

## Kiểm tra hoàn thành
- [ ] `go build ./...` thành công
- [ ] `POST /visits` tạo visit, trả đúng status `REGISTERED`
- [ ] `GET /visits?doctor_id=me&date=today` trả đúng worklist
- [ ] `POST /visits/:id/vitals` lưu vitals, `GET` trả lịch sử
- [ ] `POST /visits/:id/orders` publish Redis Stream `HIS.VISIT.LabOrderCreated`
- [ ] `POST /visits/:id/close` publish Redis Stream `HIS.VISIT.VisitClosed`
- [ ] `GET /icd10/search?q=tim` trả kết quả full-text search
