# Sprint 4 — Reception & Visit (Tuần 7–8)

> **Mục tiêu:** Check-in, quản lý hàng đợi realtime (WebSocket), tạo visit khám bệnh, nhập vitals.
> **Prerequisite:** Sprint 3 hoàn thành. Patient + Appointment APIs ready.
> **Kết thúc sprint:** Receptionist check-in và quản lý queue realtime; Doctor thấy worklist và nhập vitals.

---

## BACKEND

### Module `internal/reception`

**Domain layer:**
- [ ] Entity `QueueEntry`: id, patient_id, visit_id, service_type, queue_number, status, called_at, completed_at
- [ ] Queue number generation: format `{prefix}{number}` theo service (e.g. `KB001`, `XN001`)
- [ ] Status: `WAITING` → `CALLED` → `IN_PROGRESS` → `DONE` | `SKIPPED`
- [ ] Repository interface: `QueueRepository`

**Application layer:**
- [ ] Command: `CheckInCommand` — tạo QueueEntry + link với appointment (nếu có)
- [ ] Command: `CallQueueCommand` — gọi số → push WebSocket event
- [ ] Command: `SkipQueueCommand`
- [ ] Command: `CompleteQueueCommand`
- [ ] Query: `GetCurrentQueue`, `GetQueueStats`

### WebSocket — Queue Realtime
- [ ] Fiber WebSocket endpoint: `GET /api/v1/queue/ws`
  - Authenticate: validate JWT từ query param `?token=...`
  - On connect: gửi current queue state
  - Broadcast khi queue thay đổi
- [ ] WebSocket hub: `pkg/ws/hub.go`
  ```go
  type Hub struct {
    clients    map[*Client]bool
    broadcast  chan []byte
    register   chan *Client
    unregister chan *Client
  }
  func (h *Hub) Run()
  func (h *Hub) Broadcast(event WSEvent)
  ```
- [ ] Event types: `queue.updated`, `queue.called`, `queue.completed`

> ⚠️ **NOTE:** WebSocket phải handle disconnect/reconnect gracefully.
> Implement heartbeat ping/pong (interval 30s). Nếu client không pong → close connection.

### APIs — Reception

```
GET  /api/v1/queue              [RECEPTIONIST, DOCTOR, NURSE]
  response: current queue state (all entries với status)

POST /api/v1/queue/checkin      [RECEPTIONIST]
  body: { patient_id, service_type, appointment_id? }
  → Generate queue_number → create QueueEntry → publish HIS.VISIT.QueueCheckedIn
  → Broadcast queue.updated via WebSocket

POST /api/v1/queue/call/:id     [RECEPTIONIST, DOCTOR]
  → Update status CALLED → Broadcast queue.called event

POST /api/v1/queue/skip/:id     [RECEPTIONIST]
POST /api/v1/queue/complete/:id [DOCTOR]

GET  /api/v1/queue/stats        [ADMIN, RECEPTIONIST]
  response: { waiting_count, called_count, avg_wait_minutes }
```

---

### Module `internal/visit`

**Domain layer:**
- [ ] Entity `Visit`: id, patient_id, doctor_id, queue_entry_id, status, chief_complaint, started_at, completed_at
- [ ] Status machine: `REGISTERED` → `WAITING` → `IN_PROGRESS` → `ORDERED` → `COMPLETED` | `CANCELLED`
- [ ] Entity `VisitVital`: visit_id, bp_systolic, bp_diastolic, heart_rate, temperature, spo2, weight_kg, height_cm, recorded_at, recorded_by
- [ ] Entity `VisitOrder`: visit_id, order_type (LAB/RADIOLOGY/PROCEDURE), ref_id, status
- [ ] ICD-10 search: PostgreSQL full-text search trên `icd10_codes` (code + description_vi)
- [ ] Repository interfaces: `VisitRepository`

**Application layer:**
- [ ] Command: `CreateVisitCommand`, `UpdateVisitStatusCommand`
- [ ] Command: `RecordVitalsCommand`
- [ ] Command: `CreateVisitOrderCommand`
- [ ] Command: `CloseVisitCommand` → trigger billing
- [ ] Query: `GetDoctorWorklist`, `GetVisitDetail`, `SearchICD10`

### APIs — Visit

```
GET  /api/v1/visits             [DOCTOR, NURSE]
  query: status, doctor_id, date
  → Doctor worklist: visits của bác sĩ hiện tại trong ngày

POST /api/v1/visits             [RECEPTIONIST]
  body: { patient_id, doctor_id, queue_entry_id }
  → Tạo visit → publish HIS.VISIT.VisitStarted

GET  /api/v1/visits/:id         [DOCTOR, NURSE, LAB_TECH]

PUT  /api/v1/visits/:id/status  [DOCTOR, NURSE]
  body: { status }

POST /api/v1/visits/:id/vitals  [DOCTOR, NURSE]
  body: { bp_systolic, bp_diastolic, heart_rate, temperature, spo2, weight_kg, height_cm }

GET  /api/v1/visits/:id/vitals  [DOCTOR, NURSE]
  response: vitals history cho visit này

POST /api/v1/visits/:id/orders  [DOCTOR]
  body: { order_type, details }
  → Tạo VisitOrder → publish HIS.VISIT.LabOrderCreated hoặc RadiologyOrderCreated

GET  /api/v1/visits/:id/orders  [DOCTOR, LAB_TECH, PHARMACIST]

POST /api/v1/visits/:id/close   [DOCTOR]
  → Cập nhật status COMPLETED → publish HIS.VISIT.VisitClosed → trigger billing

GET  /api/v1/icd10/search       [DOCTOR]
  query: q (full-text search)
  response: list { code, description_vi, category }
```

### NATS Events
- [ ] `HIS.VISIT.VisitStarted` — audit worker ghi log
- [ ] `HIS.VISIT.LabOrderCreated` — LIS module sẽ lắng nghe (Sprint 5)
- [ ] `HIS.VISIT.VisitClosed` → billing worker tạo invoice (Sprint 7)

---

## DESKTOP

### Prerequisite
- Queue WebSocket endpoint phải sẵn sàng
- Visit creation, Vitals API phải ready

### Queue Dashboard (Receptionist — màn hình chính)
- [ ] Kết nối WS: `ws://localhost:8080/api/v1/queue/ws?token={access_token}`
- [ ] Reconnect logic: exponential backoff 1s→2s→4s→8s→30s (max)
- [ ] Hiển thị queue theo cột service_type:
  - Badge số thứ tự hiện đang được gọi (lớn, nổi bật)
  - Danh sách đang chờ bên dưới
- [ ] Nút "Gọi số tiếp" → `POST /queue/call/:id`
- [ ] Nút "Bỏ qua" → `POST /queue/skip/:id`
- [ ] WS event `queue.called` → animation highlight số được gọi

> ⚠️ **NOTE:** WS token sẽ expire sau 15 phút (JWT TTL).
> Khi WS disconnect do token expire: refresh token → reconnect với token mới.

### Check-in Flow (Receptionist)
- [ ] Button "Check-in mới" → Drawer/Modal
- [ ] Bước 1: Tìm bệnh nhân (dùng shared Patient Search Modal)
- [ ] Bước 2: Chọn dịch vụ (dropdown)
- [ ] Bước 3 (optional): Link với appointment (nếu bệnh nhân có lịch hẹn hôm nay)
- [ ] Submit → `POST /queue/checkin` → hiển thị số thứ tự được cấp

### Doctor Worklist
- [ ] Table/List: bệnh nhân chờ khám của bác sĩ hiện tại
- [ ] Gọi: `GET /visits?status=WAITING&doctor_id=me&date=today`
- [ ] Cột: số TT, tên bệnh nhân, giờ check-in, lý do khám
- [ ] Click → mở Visit Screen
- [ ] Real-time update: subscribe WS event `queue.updated`

### Visit Screen (Doctor)
- [ ] Header: tên bệnh nhân, tuổi, giới tính, dị ứng nổi bật
- [ ] Tab 1: **Vitals** — form nhập huyết áp, mạch, nhiệt độ, SpO2, cân nặng, chiều cao
  - Submit → `POST /visits/:id/vitals`
  - Highlight bất thường (ngoài ngưỡng bình thường)
- [ ] Tab 2: **Chẩn đoán & Chỉ định** (placeholder cho Sprint 5 EMR)
- [ ] Tab 3: **Lịch sử khám** — `GET /patients/:id/history`

---

## WEB

> Sprint 4 không có feature mới cho Web.
> Web team dùng thời gian này để:

- [ ] Polish Sprint 3 features (booking flow UX, my-appointments)
- [ ] Implement Skeleton loading states cho tất cả pages
- [ ] Error boundary setup
- [ ] Responsive: test trên mobile viewport
- [ ] SEO meta tags cho Landing page và Book page

---

## ĐIỂM KẾT NỐI Sprint 4

| Vấn đề | Backend | Desktop | Web |
|--------|---------|---------|-----|
| WebSocket auth | Validate JWT từ `?token=` query param | Gửi token khi connect | — |
| WS token expire | Server close connection khi token invalid | Detect close → refresh → reconnect | — |
| Queue broadcast | Hub broadcast tới tất cả connected clients | Subscribe và update UI | — |
| VisitClosed event | Publish → billing (Sprint 7) | Doctor nút "Kết thúc khám" | — |

## DEFINITION OF DONE

- [ ] WebSocket queue broadcast hoạt động: check-in → tất cả Desktop clients cập nhật realtime
- [ ] WebSocket reconnect đúng khi mất kết nối
- [ ] Check-in flow: tìm bệnh nhân → assign queue number → hiện trên Queue Dashboard
- [ ] Doctor worklist hiển thị bệnh nhân đang chờ
- [ ] Vitals form nhập và lưu thành công
- [ ] ICD-10 search API trả đúng kết quả
- [ ] Visit orders API tạo thành công
- [ ] NATS events publish đúng: VisitStarted, LabOrderCreated
