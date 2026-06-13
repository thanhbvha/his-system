# Sprint 3 — Patient & Appointment (Tuần 5–6)

> **Mục tiêu:** Quản lý bệnh nhân (PII encrypted) và đặt lịch khám — cả Desktop lẫn Web đều dùng.
> **Prerequisite:** Sprint 2 hoàn thành ✅, Auth middleware hoạt động ✅, token refresh OK ✅.
> **Kết thúc sprint:** Receptionist tìm/tạo bệnh nhân trên Desktop; Bệnh nhân đặt lịch được qua Web.

---

## 🏗️ KẾ THỪA TỪ SPRINT 2 (Đã hoàn thành — Không cần làm lại)

> Đây là nền tảng Sprint 3 sẽ kế thừa trực tiếp từ Sprint 2.

### Backend Foundation ✅
- **Domain Models sẵn sàng:** Entity `User`, `Role`, `Permission`, `Device`, `MFA`, `Department`, `Patient` đã được khởi tạo với Clean Architecture.
  - ⚠️ Entity `Patient` đã có cấu trúc cơ bản — Sprint 3 sẽ **mở rộng** (thêm PII encryption, value objects, relations).
- **JWT AES-GCM Infrastructure:** `pkg/auth/jwt.go` với mã hoá payload sẵn sàng — dùng để bảo vệ API Patient/Appointment.
- **Postgres Repositories (PGX):** Patterns repository đã được thiết lập — Sprint 3 follow cùng pattern.
- **Database Migrations:** Hệ thống migration đã chạy ổn định — Sprint 3 thêm migration mới cho Patient/Appointment tables.
- **Redis Infrastructure:** Kết nối Redis ổn định (đã fix double-prefix bug). Sẵn sàng cho Redis Stream events.

### Auth & RBAC Middleware ✅
- `JWTAuth` middleware: giải mã và verify JWT — **dùng ngay** cho tất cả Patient/Appointment APIs.
- `RequireRole` / `RequirePermission`: phân quyền RECEPTIONIST, DOCTOR, NURSE, PATIENT — **dùng ngay**.
- `RequestSignature`: ký request từ Desktop — **áp dụng** cho các API nhạy cảm.
- **Roles đã có:** `ADMIN`, `RECEPTIONIST`, `DOCTOR`, `NURSE`, `LAB_TECH`, `PHARMACIST`, `PATIENT`.

### Admin Infrastructure ✅
- CRUD User, Role, Permission, Department đã hoàn chỉnh.
- Department (Khoa/Phòng ban) schema có trường `code` — Doctor schedule sẽ link vào Department.

### Desktop Frontend (Wails + React) ✅
- `apiClient.ts` với interceptor tự động ký Hardware Key + Silent Refresh Token.
- Zustand persist storage đã cấu hình — Sprint 3 thêm `patientStore`, `appointmentStore`.
- Admin Dashboard layout sẵn sàng — Sprint 3 thêm section Receptionist/Doctor.

### Web Frontend (React) ✅
- Interceptor Cookie Refresh Token tự động.
- Zustand global state management.
- OTP/Login flow cho bệnh nhân — bệnh nhân đã có `patient_id` sau đăng ký, dùng luôn cho `/patients/me`.

---

## BACKEND

### Module `internal/patient`

> ⚠️ Entity `Patient` đã có stub từ Sprint 2. Sprint 3 **hoàn thiện** entity này với PII encryption đầy đủ.

**Domain layer:**
- [ ] Hoàn thiện Entity `Patient`: id, full_name, dob, gender, phone_encrypted, phone_hmac, cccd_encrypted, cccd_hmac, email_encrypted, email_hmac, address_detail_encrypted, blood_type, is_active
- [ ] Value object `PhoneNumber` — validate 10 số VN, tự encrypt/HMAC
- [ ] Value object `CCCD` — validate 12 số, tự encrypt/HMAC
- [ ] Value object `BHYTNumber` — validate format Bộ Y tế
- [ ] Entity `PatientInsurance`: patient_id, bhyt_number_encrypted, bhyt_hmac, valid_from, valid_to, coverage_level
- [ ] Entity `PatientContact`: patient_id, name, phone_encrypted, phone_hmac, relationship
- [ ] Repository interfaces: `PatientRepository`

**Application layer:**
- [ ] Command: `CreatePatientCommand`, `UpdatePatientCommand`
- [ ] Command: `UpdateInsuranceCommand`
- [ ] Query: `SearchPatients` (search by phone_hmac, cccd_hmac, full_name fuzzy)
- [ ] Query: `GetPatientByID`, `GetPatientHistory`
- [ ] Handlers + DTOs

**Infrastructure:**
- [ ] `PatientRepositoryPG` — encrypt/decrypt PII tự động trong repo layer (kế thừa pattern PGX từ Sprint 2)
- [ ] Full-text search: PostgreSQL `tsvector` cho `full_name`
- [ ] Masked response: list endpoint trả `phone_masked = "09x***xxx"`
- [ ] Database Migration: thêm columns PII + tsvector, index `phone_hmac`, `cccd_hmac`

> ⚠️ **NOTE:** Search phải dùng `phone_hmac` / `cccd_hmac` cho exact-match.
> Full-text search chỉ cho `full_name`. Không bao giờ query plaintext PII.

### APIs — Patient

> Tất cả routes được bảo vệ bởi `JWTAuth` + `RequireRole` middleware từ Sprint 2.

```
GET  /api/v1/patients
  query: q (tìm theo name full-text), phone (HMAC internally), cccd (HMAC internally)
  query: page, limit
  response: list patients với PII masked
  auth: RECEPTIONIST, DOCTOR, NURSE, LAB_TECH, PHARMACIST

POST /api/v1/patients                    [RECEPTIONIST, ADMIN]
  body: { full_name, dob, gender, phone, cccd?, email?, address? }
  → encrypt PII → create → publish PatientRegistered event

GET  /api/v1/patients/:id                [staff roles]
  response: full patient detail (decrypt PII cho staff)
  
GET  /api/v1/patients/me                 [PATIENT - Web]
  response: patient detail (masked PII)

PUT  /api/v1/patients/:id                [RECEPTIONIST, ADMIN]
PUT  /api/v1/patients/me                 [PATIENT - Web]
  → re-encrypt PII nếu thay đổi phone/cccd, update HMAC

GET  /api/v1/patients/:id/history        [DOCTOR, NURSE]
  response: danh sách visits summary

GET  /api/v1/patients/me/visits          [PATIENT - Web]
  response: danh sách visits của chính bệnh nhân

GET  /api/v1/patients/:id/insurance      [staff roles]
GET  /api/v1/patients/me/insurance       [PATIENT]
PUT  /api/v1/patients/me/insurance       [PATIENT]
```

> ⚠️ **NOTE:** `/patients/:id` cho staff xem đầy đủ PII (decrypt).
> `/patients/me` cho patient chỉ trả masked. Backend phải enforce theo JWT role.

### Redis Stream Event
- [ ] Publish `HIS.PATIENT.PatientRegistered` khi tạo mới (dùng Redis native driver đã ổn định từ Sprint 2)
- [ ] Publish `HIS.PATIENT.PatientUpdated` khi cập nhật

---

### Module `internal/appointment`

**Domain layer:**
- [ ] Entity `Appointment`: id, patient_id, doctor_id, service_id, slot_id, status, note, scheduled_at
- [ ] Entity `AppointmentSlot`: id, doctor_id, date, start_time, end_time, is_booked
- [ ] Entity `DoctorSchedule`: doctor_id, department_id (→ link Department từ Sprint 2), day_of_week, start_time, end_time, slot_duration_min
- [ ] Entity `SlotTemplate`: template rules → generate slots
- [ ] Status machine: `PENDING` → `CONFIRMED` → `CHECKED_IN` → `COMPLETED` | `CANCELLED`
- [ ] Repository interfaces: `AppointmentRepository`, `SlotRepository`

**Application layer:**
- [ ] Command: `BookAppointmentCommand`, `CancelAppointmentCommand`, `ConfirmAppointmentCommand`
- [ ] Command: `GenerateSlotsCommand` (từ template)
- [ ] Query: `GetAvailableSlots`, `ListAppointments`, `GetDoctorSchedule`
- [ ] Conflict detection: check `is_booked = false` trước khi book → atomic update

> ⚠️ **NOTE:** Booking phải dùng **SELECT FOR UPDATE** (PG row lock) để tránh double booking.
> Sau khi book: `UPDATE appointment_slots SET is_booked = true WHERE id = $1 AND is_booked = false`
> Nếu affected rows = 0 → slot đã bị book → trả `409 Conflict`.

### APIs — Appointment

**Public (không cần auth):**
```
GET /api/v1/public/clinic-info          → tên, địa chỉ, giờ làm việc
GET /api/v1/public/doctors              → danh sách bác sĩ (public profile, ảnh, chuyên môn)
GET /api/v1/public/services             → danh sách dịch vụ + giá
GET /api/v1/public/doctors?service_id=  → bác sĩ theo dịch vụ
```

> ℹ️ Doctor profile lấy từ `User` + `Department` đã có từ Sprint 2 — thêm fields: specialty, avatar_url, bio.

**Auth required:**
```
GET  /api/v1/appointments/slots
  query: doctor_id, date
  response: list slots còn trống (is_booked = false)

GET  /api/v1/appointments
  query: date, doctor_id, status, patient_id
  auth: staff xem tất cả; PATIENT chỉ xem của mình

POST /api/v1/appointments               [PATIENT, RECEPTIONIST]
  body: { doctor_id, service_id, slot_id, note? }
  → SELECT FOR UPDATE slot → book → publish AppointmentScheduled

PUT  /api/v1/appointments/:id           [RECEPTIONIST, ADMIN]
  → update status (CONFIRMED, ...)

DELETE /api/v1/appointments/:id         [PATIENT, RECEPTIONIST]
  → Chỉ cancel được nếu scheduled_at > now + 24h
  → Publish AppointmentCancelled
```

### Redis Stream Events
- [ ] `HIS.APPOINTMENT.AppointmentScheduled` → notification worker gửi SMS xác nhận
- [ ] `HIS.APPOINTMENT.AppointmentCancelled` → notification worker gửi SMS thông báo
- [ ] `HIS.APPOINTMENT.AppointmentConfirmed` → notification worker gửi SMS confirm

---

## DESKTOP (Receptionist + Doctor)

### Prerequisite
- Patient APIs và Appointment APIs phải sẵn sàng
- `apiClient.ts` với Hardware Key signing ✅ (đã có từ Sprint 2 — dùng ngay)
- Zustand store ✅ (đã có từ Sprint 2 — thêm `patientStore`, `appointmentStore`)
- Admin Dashboard layout ✅ (đã có từ Sprint 2 — thêm menu items Receptionist)

### Patient Search Modal (Shared Component)
- [ ] Input debounce 300ms → `GET /patients?q={input}`
- [ ] Hiển thị: tên, SĐT masked, ngày sinh, mã BN
- [ ] Select → callback với patient object
- [ ] Dùng chung cho Receptionist check-in, Doctor worklist

> ⚠️ **NOTE:** Kết quả search hiển thị SĐT masked (09x***xxx) — không bao giờ hiển thị số đầy đủ trong list.

### Patient Registration Form (Receptionist)
- [ ] Ant Design Form: Họ tên, Ngày sinh (DatePicker), Giới tính, SĐT, CCCD (optional), Email (optional)
- [ ] Validate: SĐT 10 số, CCCD 12 số
- [ ] Submit → `POST /patients`
- [ ] Hiển thị mã BN sau khi tạo

### Patient Detail View (Receptionist/Doctor)
- [ ] Tab: Thông tin cá nhân, BHYT, Lịch sử khám
- [ ] Staff xem được SĐT đầy đủ (decrypt từ backend)
- [ ] Edit button → mở form cập nhật

### Appointment Calendar (Receptionist)
- [ ] Calendar view (Ant Design Calendar hoặc custom)
- [ ] Chọn ngày → `GET /appointments?date=...`
- [ ] Hiển thị danh sách lịch hẹn trong ngày (theo bác sĩ)
- [ ] Status badge màu theo trạng thái

### Manual Booking (Receptionist)
- [ ] Modal: chọn bác sĩ → chọn dịch vụ → chọn slot
- [ ] `GET /appointments/slots?doctor_id=&date=`
- [ ] Hiển thị slot dạng time-grid
- [ ] Submit → `POST /appointments`
- [ ] Xử lý 409 Conflict: "Slot vừa được đặt, vui lòng chọn lại"

### Confirm/Update Appointment (Receptionist)
- [ ] Table: lịch hẹn chờ xác nhận → nút "Xác nhận"
- [ ] `PUT /appointments/:id` với status `CONFIRMED`

---

## WEB (Patient)

### Prerequisite
- Public endpoints phải ready: `/public/clinic-info`, `/public/doctors`, `/public/services`
- Appointment slots API phải ready
- OTP Login/Register flow ✅ (đã có từ Sprint 2 — bệnh nhân đã có `patient_id`)
- Cookie Refresh Token interceptor ✅ (đã có từ Sprint 2 — dùng ngay)
- Zustand store ✅ (đã có từ Sprint 2 — thêm `bookingStore`)

### Landing Page (`/`)
- [ ] Section: Giới thiệu phòng khám — `GET /public/clinic-info`
- [ ] Section: Danh sách bác sĩ — `GET /public/doctors` (card: ảnh, tên, chuyên khoa)
- [ ] Section: Dịch vụ & giá — `GET /public/services`
- [ ] CTA button: "Đặt lịch ngay" → `/book` (nếu chưa login → `/login` với returnUrl)
- [ ] Footer: địa chỉ, SĐT, giờ làm việc

### Booking Flow (`/book`) — 4 bước
- [ ] Zustand `bookingStore`: lưu state qua các bước {service, doctor, slot, note}
- [ ] Progress indicator (Step 1/4...)

**Step 1: Chọn dịch vụ**
- `GET /public/services` → Grid card dịch vụ
- Click → lưu service_id → next step

**Step 2: Chọn bác sĩ**
- `GET /public/doctors?service_id={id}` → Card bác sĩ (ảnh, tên, chuyên môn, rating)
- Click → lưu doctor_id → next step

**Step 3: Chọn ngày & giờ**
- Calendar picker: highlight ngày có slot còn trống
- Chọn ngày → `GET /appointments/slots?doctor_id=&date=`
- Hiển thị slot dạng grid (08:00, 08:30, 09:00...)
- Slot đã book: disabled, màu xám
- Click slot → lưu slot_id → next step

> ⚠️ **NOTE:** Slot phải refetch khi đổi ngày. Xử lý 409 khi submit:
> "Slot này vừa được đặt, vui lòng chọn giờ khác"

**Step 4: Xác nhận**
- Tóm tắt: tên bác sĩ, dịch vụ, ngày giờ, phí
- Textarea ghi chú (optional)
- Submit → `POST /appointments`
- Success screen → nút "Xem lịch hẹn" → `/my-appointments`

### My Appointments (`/my-appointments`)
- [ ] Tab Upcoming: `GET /appointments?patient_id=me&status=upcoming`
- [ ] Tab History: `GET /appointments?patient_id=me&status=past`
- [ ] Appointment card: bác sĩ, dịch vụ, ngày giờ, status badge
- [ ] Nút Hủy (chỉ hiện nếu `scheduled_at > now + 24h`)
- [ ] Confirm dialog trước khi hủy → `DELETE /appointments/:id`
- [ ] Empty state với CTA "Đặt lịch ngay"

---

## ĐIỂM KẾT NỐI Sprint 3

| Vấn đề | Backend | Desktop | Web |
|--------|---------|---------|-----|
| PII search | HMAC internally | Search modal debounce | — |
| PII display | Decrypt cho staff, Masked cho patient | Full PII cho Receptionist/Doctor | Masked trong account |
| Slot conflict | SELECT FOR UPDATE + 409 | Hiển thị error rõ ràng | Hiển thị error + refetch slots |
| SMS sau booking | Redis Stream → notification worker | — | Bệnh nhân nhận SMS |
| Public endpoints | `/public/*` không cần auth | — | Landing page + Booking step 1–2 |
| RBAC enforcement | JWTAuth + RequireRole (từ Sprint 2 ✅) | Hardware Key signing (từ Sprint 2 ✅) | Cookie interceptor (từ Sprint 2 ✅) |
| Department link | Doctor → Department (từ Sprint 2 ✅) | Hiển thị Khoa/Phòng ban của bác sĩ | Chuyên khoa bác sĩ trên landing |

---

## DEFINITION OF DONE

- [ ] Patient CRUD hoạt động (PII encrypt/decrypt đúng)
- [ ] Search patient theo SĐT/CCCD/tên hoạt động
- [ ] Double booking được prevent (test concurrent requests)
- [ ] Desktop: tìm kiếm + đăng ký bệnh nhân thành công
- [ ] Desktop: xem lịch hẹn theo ngày, đặt lịch thủ công
- [ ] Web: landing page hiển thị đúng thông tin
- [ ] Web: booking flow 4 bước hoạt động end-to-end
- [ ] Web: my-appointments hiển thị và hủy được
- [ ] SMS xác nhận lịch hẹn gửi thành công (hoặc mock)
- [ ] Redis Stream events được publish đúng
- [ ] RBAC đúng: PATIENT không truy cập được `/patients/:id` (staff only)
- [ ] Không có plaintext PII nào lọt vào DB hoặc logs
