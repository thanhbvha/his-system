# Sprint 3 — Step 1: Database Schema Design & Migration ✅

## Mục tiêu
Thiết kế toàn bộ các bảng database cần thiết cho Sprint 3, đảm bảo đầy đủ field, index đúng chỗ, và chạy migrate thành công.

## Migration đã chạy
File: `migrations/postgres/000009_sprint3_patient_appointment_schema.up.sql`

---

## Schema chi tiết

### 1. `patients` — Thông tin bệnh nhân (PII encrypted)

| Column | Type | Ghi chú |
|--------|------|---------|
| `id` | UUID PK | auto gen |
| `full_name` | VARCHAR(255) | cleartext, dùng full-text search |
| `full_name_search` | TSVECTOR | auto-update qua trigger |
| `dob` | DATE | Ngày sinh |
| `gender` | VARCHAR(10) | `MALE` / `FEMALE` / `OTHER` |
| `blood_type` | VARCHAR(5) | `A+`, `B-`, `O`... |
| `is_active` | BOOLEAN | soft delete |
| `phone_encrypted` | TEXT | AES-GCM mã hóa |
| `phone_hmac` | VARCHAR(64) | SHA-256 hex — **UNIQUE index** |
| `cccd_encrypted` | TEXT | AES-GCM mã hóa |
| `cccd_hmac` | VARCHAR(64) | SHA-256 hex — **UNIQUE index** |
| `email_encrypted` | TEXT | optional |
| `email_hmac` | VARCHAR(64) | index |
| `address_detail_encrypted` | TEXT | optional |
| `avatar_url` | TEXT | optional |

**Indexes:**
- `UNIQUE idx_patients_phone_hmac` — exact-match search SĐT (siêu tốc)
- `UNIQUE idx_patients_cccd_hmac` — exact-match search CCCD
- `GIN idx_patients_full_name_search` — full-text search họ tên

**Trigger:** `trg_patients_search_vector` — tự động cập nhật `full_name_search` khi insert/update.

---

### 2. `patient_insurance` — Bảo hiểm y tế (BHYT)

| Column | Type | Ghi chú |
|--------|------|---------|
| `patient_id` | UUID FK | references patients |
| `bhyt_number_encrypted` | TEXT | mã hóa |
| `bhyt_hmac` | VARCHAR(64) | index |
| `valid_from` / `valid_to` | DATE | thời hạn |
| `coverage_level` | VARCHAR(50) | `80%`, `100%` |
| `issuing_province` | VARCHAR(100) | tỉnh cấp |

---

### 3. `patient_contacts` — Người liên hệ khẩn cấp

| Column | Type | Ghi chú |
|--------|------|---------|
| `patient_id` | UUID FK | references patients |
| `name` | VARCHAR(255) | tên người thân |
| `relationship` | VARCHAR(100) | FATHER / MOTHER / SPOUSE / OTHER |
| `phone_encrypted` | TEXT | mã hóa |
| `phone_hmac` | VARCHAR(64) | |
| `is_primary` | BOOLEAN | liên hệ chính |

---

### 4. `services` — Danh mục dịch vụ/chuyên khoa

| Column | Type | Ghi chú |
|--------|------|---------|
| `code` | VARCHAR(50) UNIQUE | mã dịch vụ |
| `name` | VARCHAR(255) | tên dịch vụ |
| `price` | NUMERIC(12,2) | giá |
| `duration_min` | INT | thời gian 1 slot (mặc định 30p) |
| `is_active` | BOOLEAN | |

---

### 5. `staff_profiles` (ALTER — thêm fields)

| Column mới | Type | Ghi chú |
|-----------|------|---------|
| `specialty` | VARCHAR(255) | chuyên môn bác sĩ |
| `avatar_url` | TEXT | ảnh đại diện |
| `bio` | TEXT | giới thiệu ngắn |

---

### 6. `doctor_schedules` — Lịch làm việc (template tuần)

| Column | Type | Ghi chú |
|--------|------|---------|
| `doctor_id` | UUID FK | references users |
| `department_id` | UUID FK | references departments |
| `day_of_week` | SMALLINT | 0=Mon … 6=Sun |
| `start_time` / `end_time` | TIME | ca làm việc |
| `slot_duration_min` | INT | thời gian 1 slot |
| `is_active` | BOOLEAN | |

**Unique:** `(doctor_id, day_of_week, start_time)` — không tạo trùng lịch.

---

### 7. `appointment_slots` — Slot cụ thể (generated từ schedule)

| Column | Type | Ghi chú |
|--------|------|---------|
| `doctor_id` | UUID FK | |
| `schedule_id` | UUID FK | nguồn gốc từ schedule |
| `slot_date` | DATE | ngày cụ thể |
| `start_time` / `end_time` | TIME | giờ cụ thể |
| `is_booked` | BOOLEAN | **chống double booking** |

**Indexes:**
- `(doctor_id, slot_date)` — lấy slots của bác sĩ theo ngày
- Partial index `WHERE is_booked = false` — tối ưu query slot còn trống

**⚠️ Anti-double-booking:** Khi đặt lịch, phải dùng `SELECT FOR UPDATE` trong transaction và kiểm tra `affected_rows = 0` → `409 Conflict`.

---

### 8. `appointments` — Lịch hẹn

| Column | Type | Ghi chú |
|--------|------|---------|
| `patient_id` | UUID FK | |
| `doctor_id` | UUID FK | |
| `service_id` | UUID FK | |
| `slot_id` | UUID FK | |
| `scheduled_at` | TIMESTAMPTZ | denormalized để display nhanh |
| `status` | VARCHAR(20) | `PENDING`→`CONFIRMED`→`CHECKED_IN`→`COMPLETED`\|`CANCELLED` |
| `note` | TEXT | |
| `cancel_reason` | TEXT | |
| `booked_by` | UUID FK | NULL nếu bệnh nhân tự đặt |
| `confirmed_at`, `checked_in_at`, `completed_at`, `cancelled_at` | TIMESTAMPTZ | audit trail |

**Constraint:** `CHECK (status IN ('PENDING','CONFIRMED','CHECKED_IN','COMPLETED','CANCELLED'))`

---

## Quy tắc PII bất biến
> ❌ Không bao giờ lưu SĐT / CCCD / Email plaintext vào database.
> ✅ Luôn lưu dạng `_encrypted` + `_hmac`.
> ✅ Search exact-match qua `_hmac`, full-text qua `tsvector`.
> ✅ Không bao giờ log PII.

---

## Các Step tiếp theo

| Step | Nội dung |
|------|---------|
| **Step 2** | Patient Domain: Entity, Value Objects, Repository interface |
| **Step 3** | Patient Application Layer: Commands, Queries, Handlers |
| **Step 4** | Patient API: HTTP Handlers, Routes, Middleware |
| **Step 5** | Appointment Domain + Application Layer + API |
| **Step 6** | Desktop Frontend: Patient search, registration, appointment calendar |
| **Step 7** | Web Frontend: Landing page, Booking flow 4 bước, My appointments |
