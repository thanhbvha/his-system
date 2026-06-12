# Sprint 1 — Step 2: Database & Migration

> **Mục tiêu:** Thiết lập schema database cốt lõi và seed data ban đầu.
> **Phụ thuộc:** Step 1 (Docker Compose chạy PostgreSQL healthy).
> **Output:** Migration 001–003 chạy thành công, seed data có sẵn.

---

## 1. Cài golang-migrate

```bash
go get -tags 'postgres' github.com/golang-migrate/migrate/v4
```

- [x] Tạo `cmd/migrate/main.go` — CLI runner nhận `--up` / `--down` / `--steps`
- [x] Thư mục migrations: `migrations/postgres/`

---

## 2. Migration Files

### `001_identity.sql`

- [x] Tạo các bảng:
  | Bảng | Mô tả |
  |------|-------|
  | `users` | Tài khoản hệ thống |
  | `roles` | Vai trò (admin, doctor, nurse, receptionist, patient) |
  | `permissions` | Quyền chi tiết |
  | `user_roles` | N-N: user ↔ role |
  | `role_permissions` | N-N: role ↔ permission |
  | `refresh_tokens` | JWT refresh token store |
  | `mfa_secrets` | TOTP secrets |
  | `login_attempts` | Brute-force protection |
  | `device_registry` | Trusted devices |
  | `departments` | Khoa/phòng ban |
  | `staff_profiles` | Thông tin nhân viên |
  | `audit_sessions` | Session audit trail |

### `002_patient.sql`

- [x] Tạo các bảng:
  | Bảng | Ghi chú |
  |------|---------|
  | `patients` | Có các cột mã hoá: `phone_encrypted`, `phone_hmac`, `cccd_encrypted`, `cccd_hmac`, `email_encrypted`, `email_hmac` |
  | `patient_contacts` | Người liên hệ khẩn cấp |
  | `patient_insurance` | Thông tin BHYT |

  > ⚠️ **NOTE:** Cột `*_hmac` dùng để tìm kiếm (lookup) mà không cần giải mã. Cột `*_encrypted` lưu giá trị AES-GCM.

### `003_appointment.sql`

- [x] Tạo các bảng:
  | Bảng | Mô tả |
  |------|-------|
  | `appointments` | Lịch hẹn chính |
  | `appointment_slots` | Slot cụ thể theo ngày |
  | `slot_templates` | Template lịch làm việc mẫu |
  | `doctor_schedules` | Lịch làm việc của bác sĩ |

---

## 3. Quy tắc viết Migration

> ⚠️ **Migration phải idempotent — chạy nhiều lần không lỗi.**

```sql
-- Ví dụ đúng
CREATE TABLE IF NOT EXISTS users (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username    VARCHAR(100) NOT NULL UNIQUE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
```

---

## 4. Seed Data

File: `migrations/postgres/seed/001_seed.sql`

- [x] Roles mặc định: `admin`, `doctor`, `nurse`, `receptionist`, `pharmacist`, `patient`
- [x] Permissions cơ bản cho từng role
- [x] ICD-10 cơ bản (top 100 mã bệnh phổ biến)
- [x] Demo user:
  ```sql
  -- Password: Admin@123 (bcrypt hash)
  INSERT INTO users (username, password_hash, is_active) VALUES
    ('admin', '$2a$10$...', true)
  ON CONFLICT (username) DO NOTHING;
  ```

---

## Definition of Done (Step 2)

- [x] `make migrate` chạy không lỗi
- [x] Migration 001–003 apply thành công, có thể chạy lại (idempotent)
- [x] Bảng `users`, `patients`, `appointments` tồn tại trong DB
- [x] Seed data: login với `admin` / `Admin@123` thành công (kiểm tra sau khi có Auth)
- [x] `make migrate` với `--down` rollback sạch
