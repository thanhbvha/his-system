# Sprint 4 — Reception & Visit (Tuần 7–8)

> **Mục tiêu:** Check-in, quản lý hàng đợi realtime (WebSocket), tạo visit khám bệnh, nhập vitals.
> **Prerequisite:** Sprint 3 hoàn thành ✅ — Patient + Appointment APIs đã sẵn sàng.
> **Kết thúc sprint:** Receptionist check-in và quản lý queue realtime; Doctor thấy worklist và nhập vitals.

---

## 🏗️ KẾ THỪA TỪ SPRINT 3 (Đã hoàn thành — Không cần làm lại)

> Các thành phần sau đây đã được Sprint 3 xây dựng và kiểm tra. Sprint 4 kế thừa trực tiếp.

### Backend ✅
- **Patient API sẵn sàng:** `GET/POST /patients`, `GET /patients/:id` hoạt động đầy đủ với PII encrypt/decrypt và masking.
- **Appointment API sẵn sàng:** Booking flow có `SELECT FOR UPDATE`, trả `409 Conflict` đúng khi trùng slot.
- **Redis Stream hoạt động:** `AppointmentScheduled`, `PatientRegistered` đã publish thành công.
- **RBAC Middleware:** `JWTAuth` + `RequireRole` từ Sprint 2/3 — dùng ngay cho Queue & Visit API.
- **Public API:** `/public/doctors`, `/public/services`, `/public/clinic-info` sẵn sàng.

### Desktop Frontend ✅
- **`apiClient.ts`:** Tự động ký Hardware Key + Silent Refresh Token — Sprint 4 dùng ngay.
- **`patientStore` + `appointmentStore`:** Đã có, Sprint 4 bổ sung thêm `queueStore`, `visitStore`.
- **Shared Components:** `PatientSearchModal`, `PatientRegForm` đã hoàn thiện — dùng lại cho Check-in Flow.
- **i18n chuẩn:** Toàn bộ Desktop đã đồng bộ đa ngôn ngữ — Sprint 4 chỉ cần thêm keys mới cho Queue/Visit.

### Web Frontend ✅
- **Booking Flow hoàn chỉnh:** 4 bước đặt lịch end-to-end, xử lý 409 conflict.
- **My Appointments:** Tab Sắp tới / Lịch sử, nút Hủy lịch.
- Sprint 4 tập trung polish Web (không thêm feature lớn).

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

### Redis Stream Events
- [ ] `HIS.VISIT.VisitStarted` — audit worker ghi log
- [ ] `HIS.VISIT.LabOrderCreated` — LIS module sẽ lắng nghe (Sprint 5)
- [ ] `HIS.VISIT.VisitClosed` → billing worker tạo invoice (Sprint 7)

---

## DESKTOP

### Prerequisite
- Queue WebSocket endpoint phải sẵn sàng (cần xây mới trong Sprint 4)
- Visit creation, Vitals API phải ready (cần xây mới trong Sprint 4)
- **Đã sẵn sàng từ Sprint 3 ✅:** `PatientSearchModal` (dùng lại cho Check-in), `patientStore`, `appointmentStore`, `apiClient.ts`, i18n keys chuẩn hóa

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
- [ ] Bước 1: Tìm bệnh nhân — **tái sử dụng `<PatientSearchModal>` đã có từ Sprint 3**
- [ ] Bước 2: Chọn dịch vụ (dropdown — lấy từ `publicStore.services` đã có)
- [ ] Bước 3 (optional): Link với appointment (nếu bệnh nhân có lịch hẹn hôm nay — query `appointmentStore`)
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

## WEB (Polish Sprint 3)

> Sprint 4 không có feature mới lớn cho Web. Tập trung **hoàn thiện** những gì Sprint 3 đã build.

### Hoàn thiện UX (Sprint 3 → Production-ready)
- [ ] Polish Booking Flow: thêm loading skeleton, disable button khi đang submit
- [ ] My Appointments: thêm empty state + CTA "Đặt lịch ngay"
- [ ] Error boundary toàn bộ pages (tránh white screen khi API lỗi)
- [ ] Responsive layout: test và fix trên mobile viewport (360px, 390px)
- [ ] SEO: thêm `<meta>` title/description cho Landing page (`/`) và Book page (`/book`)
- [ ] Toast notifications thay vì `window.alert` (dùng `sonner` hoặc shadcn `Toast`)

---

## ĐIỂM KẾT NỐI Sprint 4

| Vấn đề | Backend | Desktop | Web |
|--------|---------|---------|-----|
| WebSocket auth | Validate JWT từ `?token=` query param | Gửi token khi connect | — |
| WS token expire | Server close connection khi token invalid | Detect close → refresh → reconnect | — |
| Queue broadcast | Hub broadcast tới tất cả connected clients | Subscribe và update UI | — |
| Check-in tìm bệnh nhân | `GET /patients?q=` (đã có từ Sprint 3 ✅) | Tái dùng `<PatientSearchModal>` (đã có ✅) | — |
| VisitClosed event | Publish → billing (Sprint 7) | Doctor nút "Kết thúc khám" | — |
| i18n Queue/Visit | — | Thêm keys mới vào `vi.json` / `en.json` (cấu trúc đã chuẩn ✅) | — |

---

## DEFINITION OF DONE

- [ ] WebSocket queue broadcast hoạt động: check-in → tất cả Desktop clients cập nhật realtime
- [ ] WebSocket reconnect đúng khi mất kết nối (exponential backoff)
- [ ] Check-in flow: tìm bệnh nhân → assign queue number → hiện trên Queue Dashboard
- [ ] Doctor worklist hiển thị bệnh nhân đang chờ, real-time update qua WS
- [ ] Vitals form nhập và lưu thành công, highlight giá trị bất thường
- [ ] ICD-10 search API trả đúng kết quả
- [ ] Visit orders API tạo thành công, publish Redis Stream event
- [ ] Redis Stream events publish đúng: `VisitStarted`, `LabOrderCreated`, `VisitClosed`
- [ ] Web: tất cả pages có skeleton loading + error boundary
- [ ] Web: my-appointments có empty state, toast thay alert
- [ ] Web: responsive đúng trên mobile (≥ 360px)
