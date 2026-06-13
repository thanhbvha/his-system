# Sprint 3 — Step 5: Appointment Domain + Application + API (Backend)

## Mục tiêu
Implement toàn bộ module Appointment — từ Domain, Application, đến HTTP API. Điểm quan trọng nhất: **chống Double Booking** với `SELECT FOR UPDATE`.

## Cây thư mục cần tạo
```
internal/appointment/
├── domain/
│   ├── appointment.go       -- Entities + State machine
│   └── repository.go        -- Repository interfaces
├── application/
│   ├── command/
│   │   ├── book_appointment.go
│   │   ├── cancel_appointment.go
│   │   ├── confirm_appointment.go
│   │   └── generate_slots.go
│   └── query/
│       ├── get_available_slots.go
│       ├── list_appointments.go
│       └── get_doctor_schedule.go
├── infrastructure/
│   ├── appointment_repository_pg.go
│   └── slot_repository_pg.go
└── (handler đặt trong internal/api/appointment/)
```

---

## Domain Layer

### State Machine Appointment
```
PENDING → CONFIRMED → CHECKED_IN → COMPLETED
                ↘
              CANCELLED (từ PENDING hoặc CONFIRMED, nếu còn > 24h)
```

### Repository Interfaces
```go
type AppointmentRepository interface {
    Create(ctx, appointment) error
    GetByID(ctx, id) (*Appointment, error)
    List(ctx, filter ListFilter) ([]*Appointment, int64, error)
    UpdateStatus(ctx, id, status, updatedAt) error
}

type SlotRepository interface {
    Generate(ctx, schedule, date) error                   // sinh slots từ template
    GetAvailable(ctx, doctorID, date) ([]*Slot, error)
    BookSlot(ctx context.Context, slotID uuid.UUID) error // SELECT FOR UPDATE
    ReleaseSlot(ctx, slotID) error
}
```

---

## Anti-Double-Booking — `BookSlot` Implementation
```sql
-- Trong một transaction
BEGIN;

SELECT id FROM appointment_slots
WHERE id = $1 AND is_booked = false
FOR UPDATE;           -- Khóa row, chặn concurrent request

-- Nếu không có row → slot đã bị đặt → Rollback → return 409

UPDATE appointment_slots
SET is_booked = true
WHERE id = $1 AND is_booked = false;

-- Nếu affected_rows = 0 → race condition → return 409

COMMIT;
```

---

## HTTP Routes

### Public Routes (Không cần Auth)
```
GET /api/v1/public/clinic-info     → thông tin phòng khám
GET /api/v1/public/doctors         → danh sách bác sĩ (profile + specialty)
GET /api/v1/public/services        → danh sách dịch vụ + giá
GET /api/v1/public/doctors?service_id={id}  → bác sĩ theo dịch vụ
```

### Protected Routes
```
[JWTAuth]
├── GET  /api/v1/appointments/slots?doctor_id=&date=
│        [RequireRole: any authenticated]
├── GET  /api/v1/appointments?date=&doctor_id=&status=&patient_id=
│        [RequireRole: staff thấy all, PATIENT chỉ thấy của mình]
├── POST /api/v1/appointments
│        [RequireRole: PATIENT, RECEPTIONIST]
├── PUT  /api/v1/appointments/:id   → update status
│        [RequireRole: RECEPTIONIST, ADMIN]
└── DELETE /api/v1/appointments/:id → cancel (> 24h trước)
           [RequireRole: PATIENT, RECEPTIONIST]
```

---

## Redis Events
Publish sau các thao tác thành công:
- `HIS.APPOINTMENT.AppointmentScheduled` → SMS xác nhận
- `HIS.APPOINTMENT.AppointmentConfirmed` → SMS confirm
- `HIS.APPOINTMENT.AppointmentCancelled` → SMS hủy

Mock Worker (log ra console trong Sprint 3, tích hợp SMS thật ở sprint sau).

---

## Error Cases cần xử lý
| Case | HTTP Code | Thông báo |
|------|-----------|-----------|
| Slot đã bị đặt | 409 Conflict | "Slot này vừa được đặt bởi người khác" |
| Hủy quá muộn (< 24h) | 422 | "Không thể hủy lịch trong vòng 24h" |
| Patient không tồn tại | 404 | |
| Doctor không có lịch ngày đó | 404 | "Bác sĩ không làm việc ngày này" |
