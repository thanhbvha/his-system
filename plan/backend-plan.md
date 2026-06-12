# HIS BACKEND PLAN — Go Fiber API

> **Cập nhật:** 2026-06-11
> **Stack:** Go (Fiber v2) · PostgreSQL 15 · MongoDB 7 · Redis 7 · Redis Stream · MinIO
> **Kiến trúc:** Modular Monolith + DDD

**Tài liệu liên quan:**
- 📱 Desktop client: [desktop-plan.md](./desktop-plan.md)
- 🌐 Web client: [web-plan.md](./web-plan.md)

---

## 1. KIẾN TRÚC TỔNG QUAN

```
┌─────────────────────────────────────────────────┐
│           Go Fiber API — /api/v1                │
│   Auth · RBAC · RateLimit · Audit · OTel        │
└───┬──────┬──────┬──────┬──────┬──────┬──────────┘
    │      │      │      │      │      │
 identity patient appt  visit  lis  billing ...
    │      │      │      │      │      │
┌───▼──────▼──────▼──────▼──────▼──────▼──────────┐
│  PostgreSQL │ MongoDB │ Redis │ MinIO      │
└────────────────────────────────────────────────-─┘
```

### Folder Structure

```
his-system/
├── cmd/
│   ├── api/            # HTTP server entry point
│   ├── worker/         # Redis Stream consumer worker
│   └── migrate/        # DB migration runner
├── internal/
│   ├── shared/         # errors, events, valueobjects
│   ├── identity/       # Auth, User, RBAC, MFA
│   ├── patient/        # Patient profile, insurance
│   ├── appointment/    # Scheduling, slots
│   ├── reception/      # Check-in, queue
│   ├── visit/          # Outpatient visit, diagnosis
│   ├── emr/            # EMR (MongoDB)
│   ├── lis/            # Laboratory
│   ├── ris/            # Radiology
│   ├── pacs/           # DICOM metadata
│   ├── pharmacy/       # Prescription, dispensing
│   ├── inventory/      # Drug & supply warehouse
│   ├── billing/        # Invoice, payment
│   ├── inpatient/      # Ward, bed, admission
│   ├── notification/   # SMS, Email (Redis Stream consumer)
│   └── audit/          # Audit log (Redis Stream → MongoDB)
├── pkg/
│   ├── crypto/         # AES-GCM encrypt/decrypt
│   ├── auth/           # JWT issue/verify
│   ├── database/       # PG + Mongo factories
│   ├── cache/          # Redis wrapper
│   ├── messaging/      # Redis Stream wrapper
│   ├── storage/        # MinIO/S3 wrapper
│   └── logger/         # Zerolog structured logging
├── migrations/         # SQL files (numbered)
├── configs/
└── docs/               # Swagger, ERD
```

### DDD Module Structure

```
internal/{module}/
├── domain/
│   ├── entity/          # Aggregate roots
│   ├── valueobject/     # CCCD, PhoneNumber, BHYT...
│   ├── repository/      # Interface (port)
│   ├── service/         # Domain service
│   └── event/           # Domain events
├── application/
│   ├── command/         # Write operations
│   ├── query/           # Read operations
│   ├── handler/         # Command/Query handlers
│   └── dto/             # Request/Response DTOs
├── infrastructure/
│   ├── postgres/
│   ├── mongodb/
│   └── cache/
└── presentation/
    └── http/            # Fiber route handlers
```

---

## 2. DATABASE DESIGN

### PostgreSQL Schema (~150 bảng)

#### identity (~12 bảng)
```sql
users, roles, permissions, user_roles, role_permissions,
refresh_tokens, mfa_secrets, audit_sessions,
departments, staff_profiles, login_attempts, device_registry
```
> ⚠️ `mfa_secret` phải được encrypt bằng AES-GCM trước khi lưu

#### patient (~18 bảng)
```sql
patients, patient_contacts, patient_insurance, patient_allergies,
patient_chronic_conditions, patient_documents, patient_identities,
patient_family_history, patient_social_history,
blood_types, marital_statuses, occupations, ethnicities,
provinces, districts, wards_address, countries
```
> ⚠️ Các cột `cccd_number`, `bhyt_number`, `phone_number`, `email`, `address_detail`
> phải dùng dual-column pattern: `*_encrypted` (AES-GCM) + `*_hmac` (searchable)

#### appointment (~10 bảng)
```sql
appointments, appointment_slots, appointment_types,
slot_templates, doctor_schedules, appointment_cancellations,
appointment_reminders, waitlist_entries, referrals, holidays
```

#### visit (~20 bảng)
```sql
visits, visit_vitals, visit_chief_complaints, visit_notes,
diagnoses, icd10_codes, icd10_categories,
visit_orders, order_types, clinical_procedures,
procedure_catalog, care_plans, visit_status_history,
clinical_notes, visit_attachments
```

#### lis (~15 bảng)
```sql
lab_orders, lab_order_items, lab_samples, lab_sample_tracking,
lab_results, lab_result_items, lab_test_catalog, lab_panels,
lab_panel_tests, reference_ranges, lab_machines,
lab_qc_records, sample_rejection_reasons, lab_worklists,
lab_technician_assignments
```

#### ris (~10 bảng)
```sql
radiology_orders, radiology_studies, radiology_reports,
imaging_modalities, body_parts, radiology_technicians,
report_templates, contrast_records, radiation_doses,
radiology_status_history
```

#### pharmacy (~20 bảng)
```sql
drug_catalog, drug_categories, drug_units, drug_dosage_forms,
drug_routes, drug_interactions, drug_contraindications,
prescriptions, prescription_items, prescription_status_history,
dispensing_records, dispensing_items, drug_substitutions,
controlled_drug_logs, pharmacist_verifications,
drug_alerts, atc_codes, drug_manufacturers, drug_suppliers,
formulary_lists
```

#### inventory (~15 bảng)
```sql
warehouses, warehouse_locations, inventory_items, stock_lots,
stock_transactions, stock_adjustments, purchase_orders,
purchase_order_items, suppliers, supplier_contacts,
expiry_alerts, min_stock_configs, inventory_counts,
count_items, goods_receipts
```

#### billing (~15 bảng)
```sql
invoices, invoice_items, payments, payment_methods,
insurance_claims, insurance_claim_items,
price_catalogs, price_list_items, price_list_versions,
discounts, discount_rules, copayments,
payment_refunds, billing_configs, tax_configs
```
> ⚠️ `bank_account` trong `payment_methods` phải encrypt AES-GCM

#### inpatient (~20 bảng)
```sql
wards, rooms, beds, bed_types, bed_status_history,
admissions, admission_status_history,
nursing_assessments, nursing_notes, care_assignments,
medication_administrations, iv_records, fluid_charts,
discharge_summaries, transfer_records, transfer_reasons,
diet_orders, activity_orders, fall_risk_assessments,
pressure_injury_records
```

### MongoDB Collections

```js
medical_records    // EMR: SOAP, vitals, diagnoses, versioning
lab_result_documents   // LIS kết quả chi tiết
radiology_documents    // RIS báo cáo chi tiết
dicom_metadata         // PACS metadata
audit_logs             // Audit trail đầy đủ
report_snapshots       // Dashboard snapshots
notification_templates
notification_logs
```

---

## 3. SECURITY MODEL

### JWT — AES-GCM Encrypted Payload

```
Claims (JSON)
  └──► AES-GCM Encrypt (JWT_ENCRYPTION_KEY)
          └──► Base64URL(nonce+ciphertext+tag) → field "enc"
                  └──► Sign HMAC-SHA256
                          └──► JWT string
```

| Tham số | Giá trị |
|---------|---------|
| Algorithm | AES-256-GCM |
| Key size | 256-bit (32 bytes) |
| Nonce | 96-bit, random mỗi lần |
| Auth Tag | 128-bit |
| Key source | `JWT_ENCRYPTION_KEY` env/KMS |
| Key rotation | `kid` field trong header |

```go
// pkg/crypto/aes_gcm.go
func EncryptAESGCM(plaintext []byte, key []byte) ([]byte, error)
func DecryptAESGCM(ciphertext []byte, key []byte) ([]byte, error)

// pkg/auth/jwt.go
func IssueAccessToken(claims Claims) (string, error)
func VerifyAccessToken(token string) (Claims, error)
```

> ⚠️ **NOTE:** Dùng `crypto/aes` + `crypto/cipher` standard library. KHÔNG dùng third-party.

### Field-Level Encryption (AES-256-GCM)

| Trường | Bảng | Pattern |
|--------|------|---------|
| `cccd_number` | `patients` | dual-column |
| `bhyt_number` | `patient_insurance` | dual-column |
| `phone_number` | `patients`, `patient_contacts` | dual-column |
| `email` | `users`, `patients` | dual-column |
| `address_detail` | `patients` | encrypt only |
| `mfa_secret` | `mfa_secrets` | encrypt only |
| `bank_account` | `payment_methods` | encrypt only |

**Dual-column pattern (searchable):**
```
phone_encrypted  TEXT  — AES-GCM ciphertext
phone_hmac       TEXT  — HMAC-SHA256(phone, SEARCH_KEY)
→ Query: WHERE phone_hmac = HMAC(input, SEARCH_KEY)
```

```go
// pkg/crypto/field_cipher.go
type FieldCipher struct { key []byte }
func (f *FieldCipher) Encrypt(plaintext string) (string, error)
func (f *FieldCipher) Decrypt(ciphertext string) (string, error)
func (f *FieldCipher) HMAC(value string) string
```

**Storage format:** `base64(nonce | ciphertext | tag)` trong TEXT column

### RBAC Roles

| Role | Quyền |
|------|-------|
| ADMIN | Toàn quyền |
| DOCTOR | Khám, EMR, LIS/RIS order, kê đơn |
| NURSE | Tiếp nhận, Vitals, chăm sóc |
| LAB_TECH | Nhận mẫu → kết quả |
| RADIOLOGIST | Đọc phim, viết báo cáo |
| PHARMACIST | Duyệt đơn, xuất thuốc |
| RECEPTIONIST | Đặt lịch, check-in, thu ngân |
| ACCOUNTANT | Xem hóa đơn, báo cáo tài chính |

> ⚠️ **NOTE:** RBAC check phải dùng middleware per-route, không check trong handler.

---

## 4. API CONTRACT (Dùng chung cho Desktop & Web)

> ⚠️ **CRITICAL:** Mọi thay đổi API phải đồng bộ với [desktop-plan.md](./desktop-plan.md) và [web-plan.md](./web-plan.md). Dùng Swagger làm source of truth.

### Auth APIs
```
POST   /api/v1/auth/login              → { access_token, refresh_token }
POST   /api/v1/auth/refresh            → { access_token }
POST   /api/v1/auth/logout
POST   /api/v1/auth/mfa/setup          → { qr_code, secret }
POST   /api/v1/auth/mfa/verify         → { access_token } (sau khi verify TOTP)
POST   /api/v1/auth/register           → Web only (OTP flow)
POST   /api/v1/auth/otp/send           → Web only
POST   /api/v1/auth/otp/verify         → Web only
```

### Patient APIs
```
GET    /api/v1/patients                → list (search: phone_hmac, cccd_hmac, name)
POST   /api/v1/patients                → tạo mới
GET    /api/v1/patients/:id            → chi tiết (decrypt PII)
PUT    /api/v1/patients/:id            → cập nhật
GET    /api/v1/patients/:id/history    → lịch sử khám
GET    /api/v1/patients/:id/results    → kết quả XN (Web dùng)
```
> 🔗 Desktop dùng: tất cả | Web dùng: GET chi tiết + results (scope bệnh nhân tự xem)

### Appointment APIs
```
GET    /api/v1/appointments            → list (filter: date, doctor, status)
POST   /api/v1/appointments            → đặt lịch
PUT    /api/v1/appointments/:id        → cập nhật
DELETE /api/v1/appointments/:id        → hủy
GET    /api/v1/appointments/slots      → available slots (Web dùng)
GET    /api/v1/doctors/:id/schedule    → lịch bác sĩ (Web dùng)
```
> 🔗 Desktop dùng: tất cả | Web dùng: slots + book + my-appointments

### Visit APIs
```
POST   /api/v1/visits                  → tạo visit (check-in)
GET    /api/v1/visits/:id              → chi tiết visit
PUT    /api/v1/visits/:id/status       → cập nhật trạng thái
POST   /api/v1/visits/:id/vitals       → lưu vitals
POST   /api/v1/visits/:id/orders       → tạo lab/radiology order
GET    /api/v1/visits/:id/orders       → danh sách orders
POST   /api/v1/visits/:id/close        → kết thúc khám → trigger billing
```

### Queue APIs (WebSocket)
```
GET    /api/v1/queue                   → current queue state (HTTP)
WS     /api/v1/queue/ws               → realtime updates
POST   /api/v1/queue/checkin           → check-in, assign queue number
POST   /api/v1/queue/call/:id          → gọi số (bác sĩ gọi bệnh nhân)
```
> 🔗 Desktop dùng: WebSocket realtime | Web: không dùng

### EMR APIs
```
GET    /api/v1/emr/:patient_id         → lấy medical record
POST   /api/v1/emr                     → tạo/cập nhật SOAP (versioned)
GET    /api/v1/emr/:patient_id/history → version history
POST   /api/v1/emr/attachments         → upload file (MinIO)
```

### LIS APIs
```
GET    /api/v1/lab/orders              → worklist (Lab Tech)
GET    /api/v1/lab/orders/:id          → chi tiết order
PUT    /api/v1/lab/orders/:id/sample   → nhận mẫu
POST   /api/v1/lab/orders/:id/results  → nhập kết quả
PUT    /api/v1/lab/orders/:id/verify   → verify & approve
GET    /api/v1/lab/results/:id         → kết quả (Web + Desktop xem)
```
> 🔗 Web dùng: GET results (bệnh nhân xem sau verify) | Desktop dùng: tất cả

### Pharmacy APIs
```
GET    /api/v1/pharmacy/prescriptions  → queue theo status
GET    /api/v1/pharmacy/prescriptions/:id
PUT    /api/v1/pharmacy/prescriptions/:id/dispense
GET    /api/v1/drugs/search            → tìm thuốc (autocomplete)
GET    /api/v1/drugs/:id/interactions  → kiểm tra tương tác
```

### Billing APIs
```
GET    /api/v1/billing/invoices        → danh sách hóa đơn
GET    /api/v1/billing/invoices/:id    → chi tiết
POST   /api/v1/billing/invoices/:id/pay → thanh toán
GET    /api/v1/billing/invoices/:id/pdf → xuất PDF
```
> 🔗 Web dùng: GET invoice/:id (bệnh nhân xem hóa đơn của mình)

### Reporting APIs (Desktop Admin)
```
GET    /api/v1/reports/revenue         → doanh thu theo kỳ
GET    /api/v1/reports/patients        → thống kê bệnh nhân
GET    /api/v1/reports/services        → top dịch vụ
GET    /api/v1/audit/logs              → audit trail (Admin only)
```

---

## 5. EVENT-DRIVEN (Redis Stream)

```
HIS.PATIENT.*     → PatientRegistered, PatientUpdated, PatientMerged
HIS.APPOINTMENT.* → AppointmentScheduled, AppointmentConfirmed,
                    AppointmentCancelled, AppointmentCheckedIn
HIS.VISIT.*       → VisitStarted, VisitDiagnosed,
                    LabOrderCreated, RadiologyOrderCreated,
                    PrescriptionCreated, VisitClosed
HIS.LIS.*         → SampleCollected, LabResultReady, LabResultVerified
HIS.RIS.*         → StudyScheduled, StudyCompleted, ReportSigned
HIS.BILLING.*     → InvoiceCreated, PaymentReceived, ClaimSubmitted
HIS.NOTIFICATION.* → NotificationRequested
```

**Workers:**
- `notification_worker`: subscribe → SMS/Email/Zalo
- `audit_worker`: subscribe all → MongoDB audit_logs
- `report_worker`: subscribe billing/visit → update snapshots

> ⚠️ **NOTE:** Desktop WebSocket queue update được trigger từ `HIS.VISIT.*` events.
> Không poll từ client, phải push từ server qua WS.

---

## 6. INFRASTRUCTURE

### Docker Compose (Dev)
```yaml
services:
  api:        # Go (air hot-reload)
  worker:     # Redis Stream consumer
  postgres:   # PostgreSQL 15
  mongodb:    # MongoDB 7
  redis:      # Redis 7
  minio:      # Object storage
  jaeger:     # Distributed tracing
  prometheus:
  grafana:
```

### Production (VPS MVP)
- Ubuntu 22.04, 4 CPU, 8GB RAM
- Docker Compose production mode
- Nginx reverse proxy + SSL (Let's Encrypt)
- Automated backup to S3/Backblaze

### CI/CD (GitHub Actions)
```
PR:   lint → unit test → integration test → build check
Main: all tests → build image → push to GHCR
Tag:  build → push → deploy VPS (SSH + docker compose pull)
```

### Observability
```
Logging:  zerolog → stdout → Loki → Grafana
Metrics:  Prometheus → Grafana
Tracing:  OpenTelemetry → Jaeger
Alerts:   Grafana → Telegram/Email
```

### Tooling
```
air              — Go hot reload
sqlc             — Type-safe SQL codegen
swaggo/swag      — Swagger auto-gen
golang-migrate   — DB migrations
golangci-lint    — Linter
testify/mockery  — Testing + mocks
```

---

## 7. ROADMAP BACKEND — PHASE 1 (Tuần 1–16)

### Sprint 1 (Tuần 1–2): Foundation
- [ ] Init Go module, Fiber v2, Air hot-reload
- [ ] Docker Compose: PG, Mongo, Redis, MinIO
- [ ] DB connection factories (pgxpool, mongo driver, redis)
- [ ] Zerolog + OpenTelemetry setup
- [ ] golang-migrate + migration scripts
- [ ] Base error types, response wrapper, pagination helper
- [ ] Makefile targets
- [ ] GitHub Actions CI cơ bản
- [ ] Health check `/health`, `/ready`
- [ ] Swagger (swaggo) setup
- [ ] **`pkg/crypto/aes_gcm.go`** — AES-GCM core
- [ ] **`pkg/crypto/field_cipher.go`** — field-level encryption
- [ ] Migrations: `001_identity.sql`, `002_patient.sql`, `003_appointment.sql`
- [ ] Seed: roles, permissions, ICD-10 (basic), demo data

> ⚠️ **NOTE:** `pkg/crypto` phải xong trước Sprint 2 vì Identity module dùng ngay.

### Sprint 2 (Tuần 3–4): Identity & Auth
- [ ] `internal/identity`: User, Role, Permission entities
- [ ] **`pkg/auth/jwt.go`** — IssueAccessToken + VerifyAccessToken (AES-GCM payload)
- [ ] JWT access (15m) + refresh token (7d, Redis)
- [ ] TOTP MFA (`pquerna/otp`) — mfa_secret encrypt AES-GCM
- [ ] RBAC middleware (permission check per route)
- [ ] APIs: login, logout, refresh, MFA setup/verify
- [ ] Web-specific: OTP via SĐT, register flow
- [ ] User CRUD: create, list, update, deactivate
- [ ] Department management
- [ ] Password: bcrypt hash, reset flow via email/SMS
- [ ] Rate limiting on auth endpoints
- [ ] Field encryption cho `email` trong `users`

> 🔗 **Desktop cần:** login, MFA verify, user management APIs
> 🔗 **Web cần:** OTP login, register APIs
> ⚠️ **NOTE:** Test kỹ AES-GCM JWT trước khi Desktop/Web bắt đầu implement auth client.

### Sprint 3 (Tuần 5–6): Patient & Appointment
- [ ] `internal/patient`: Patient aggregate, Insurance, Contacts
- [ ] Patient CRUD + search (full-text, cccd_hmac, phone_hmac)
- [ ] AES-GCM encrypt cho tất cả PII fields
- [ ] Insurance validation (BHYT format)
- [ ] `internal/appointment`: Slot templates, Doctor schedules
- [ ] Appointment CRUD: book, cancel, reschedule
- [ ] Conflict detection (double booking prevention)
- [ ] Available slots API (Web dùng cho booking flow)
- [ ] Events: `PatientRegistered`, `AppointmentScheduled`

> 🔗 **Desktop cần:** Patient CRUD, Appointment management APIs
> 🔗 **Web cần:** GET /slots, POST /appointments, GET /my-appointments
> ⚠️ **NOTE:** Search patient phải dùng `phone_hmac`/`cccd_hmac`, không scan plaintext.

### Sprint 4 (Tuần 7–8): Reception & Visit
- [ ] `internal/reception`: Check-in, Queue number generation
- [ ] WebSocket endpoint: `/api/v1/queue/ws` — realtime queue push
- [ ] `internal/visit`: Visit lifecycle state machine
- [ ] Visit orders (lab, radiology, procedure)
- [ ] ICD-10 search API (full catalog seeded)
- [ ] Events: `VisitStarted`, `LabOrderCreated`

> 🔗 **Desktop cần:** WebSocket queue, check-in API, visit creation APIs
> ⚠️ **NOTE:** WebSocket phải stable trước khi Desktop Queue Dashboard implement.

### Sprint 5 (Tuần 9–10): EMR
- [ ] `internal/emr`: Medical record MongoDB schema + versioning
- [ ] SOAP note API (create, update — immutable versioning)
- [ ] Vitals history API
- [ ] Diagnosis finalize + ICD-10 link
- [ ] Attachment upload → MinIO
- [ ] EMR audit: diff tracking per version

> 🔗 **Desktop cần:** tất cả EMR APIs
> 🔗 **Web cần:** GET kết quả XN (read-only, sau Lab verify)
> ⚠️ **NOTE:** EMR versioning logic phải test kỹ — không được phép xóa/overwrite version cũ.

### Sprint 6 (Tuần 11–12): Pharmacy & Inventory
- [ ] `internal/pharmacy`: Drug catalog, Prescription CRUD + workflow
- [ ] Drug interaction check (local DB rules)
- [ ] Dispensing record
- [ ] `internal/inventory`: Warehouse, Stock, Transactions
- [ ] Stock deduction on dispense (atomic transaction)
- [ ] Low stock alert event → Redis Stream

> 🔗 **Desktop cần:** prescription APIs, drug search, dispensing APIs
> ⚠️ **NOTE:** Stock deduction phải atomic (PG transaction) để tránh race condition.

### Sprint 7 (Tuần 13–14): Billing & Notification
- [ ] `internal/billing`: Price catalog, Invoice auto-create on VisitClosed
- [ ] Payment: cash, transfer (manual record)
- [ ] Invoice PDF generation (chromedp)
- [ ] `internal/notification`: Redis Stream consumer worker
- [ ] SMS gateway (VNPT/Twilio)
- [ ] Email (SMTP/SendGrid)
- [ ] Triggers: appointment reminder, lab result ready

> 🔗 **Desktop cần:** billing APIs, invoice PDF
> 🔗 **Web cần:** GET invoice/:id (bệnh nhân xem)
> ⚠️ **NOTE:** Invoice auto-create trigger từ `VisitClosed` event — phải idempotent (dùng visit_id làm unique key).

### Sprint 8 (Tuần 15–16): Audit, Dashboard, Polish
- [ ] `internal/audit`: Redis Stream consumer → MongoDB audit_logs
- [ ] Reporting APIs: revenue, patient count, top services
- [ ] Report snapshot worker
- [ ] Integration test suite (testify + httptest)
- [ ] Load test cơ bản (k6)
- [ ] API documentation hoàn chỉnh (Swagger)
- [ ] Production Docker Compose
- [ ] Nginx config + SSL
- [ ] Backup script (PG dump + Mongo dump)
- [ ] Prometheus + Grafana dashboard

> 🔗 **Desktop Admin cần:** reporting APIs, audit log API

---

## 8. PHASE 2–5 OVERVIEW

| Phase | Tuần | Backend Focus |
|-------|------|---------------|
| Phase 2 | 17–32 | Inpatient, Advanced Billing (BHYT), Advanced Inventory, Shift |
| Phase 3 | 33–48 | LIS Full (HL7 v2.x), RIS Full, PACS (WADO/STOW/QIDO) |
| Phase 4 | 49–64 | Multi-tenant, Kubernetes migration, Microservices extract |
| Phase 5 | 65+  | TT 46/2018, Chữ ký số, BHYT điện tử, FHIR R4 |

---

## 9. RỦI RO & LƯU Ý

| Rủi ro | Mức | Biện pháp |
|--------|-----|-----------|
| Nghiệp vụ y tế sai quy trình | Cao | Validate với bác sĩ trước khi code |
| AES key bị lộ | Cao | Lưu key trong env/KMS, không commit vào git |
| Race condition stock/billing | Cao | Dùng PG transaction + advisory lock |
| WebSocket không stable | Trung bình | Implement reconnect + heartbeat |
| EMR version bị overwrite | Cao | Immutable append-only, test kỹ |
| Invoice duplicate | Trung bình | Idempotency key = visit_id |
| BHYT integration | Cao | MVP manual, tự động hóa Phase 5 |
| Data loss | Cao | Daily backup + point-in-time recovery |

---

## 10. OPEN QUESTIONS

1. **SMS Provider:** VNPT SMS, Twilio, hay Zalo ZNS?
2. **Deployment:** VPS cloud hay on-premise tại phòng khám?
3. **Multi-tenant:** MVP 1 cơ sở hay nhiều ngay từ đầu?
4. **Payment gateway:** VNPay/MoMo, hay chỉ cash/transfer?
5. **Offline:** Wails desktop cần offline cache không?
