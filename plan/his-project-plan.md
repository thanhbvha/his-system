# KẾ HOẠCH DỰ ÁN HIS — Chi Tiết Đầy Đủ

> **Cập nhật:** 2026-06-10  
> **Kiến trúc:** Modular Monolith  
> **Team:** Solo Developer  
> **Interface:** Web (bệnh nhân) + Desktop Wails (nhân viên/bác sĩ)  
> **Mục tiêu:** Phòng khám tư → mở rộng bệnh viện công  

---

## 1. TỔNG QUAN KIẾN TRÚC

### Stack Công Nghệ

| Layer | Công nghệ |
|-------|-----------|
| Backend | Go (Fiber v2) |
| Desktop App | Wails v2 + React + TypeScript |
| Web App (Patient) | React + Vite + TypeScript |
| Database | PostgreSQL 15 (giao dịch) + MongoDB 7 (tài liệu y tế) |
| Cache / Session | Redis 7 |
| Message Queue | NATS JetStream |
| Object Storage | MinIO (DICOM, tài liệu) |
| Logging | Zerolog → Loki |
| Tracing | OpenTelemetry → Jaeger |
| Metrics | Prometheus → Grafana |
| Auth | JWT (Access 15m + Refresh 7d) + TOTP MFA |
| Migration | golang-migrate |
| API Docs | Swaggo (Swagger) |

### Kiến Trúc Dual Interface

```
┌─────────────────────────┐     ┌─────────────────────────┐
│   WEB APP (React+Vite)  │     │  DESKTOP (Wails+React)  │
│   Bệnh nhân / Khách     │     │  Staff: Doctor, Nurse,  │
│   - Đặt lịch online     │     │  Lab, Pharmacy, Admin   │
│   - Xem kết quả XN      │     │  - Worklist, EMR        │
│   - Hồ sơ cá nhân       │     │  - Queue, Billing       │
└────────────┬────────────┘     └────────────┬────────────┘
             │ HTTPS                          │ localhost HTTP
     ┌───────▼────────────────────────────────▼───────┐
     │         Go Fiber API — /api/v1                  │
     │   Auth · RBAC · RateLimit · Audit · OTel        │
     └───┬──────┬──────┬──────┬──────┬──────┬─────────┘
         │      │      │      │      │      │
      identity patient appt  visit  lis  billing ...
         │      │      │      │      │      │
     ┌───▼──────▼──────▼──────▼──────▼──────▼─────────┐
     │  PostgreSQL │ MongoDB │ Redis │ NATS │ MinIO     │
     └─────────────────────────────────────────────────┘
```

### Folder Structure

```
his-system/
├── cmd/
│   ├── api/            # HTTP server entry point
│   ├── worker/         # NATS consumer worker
│   └── migrate/        # DB migration runner
├── internal/
│   ├── shared/         # Cross-cutting: errors, events, valueobjects
│   ├── identity/       # Auth, User, RBAC, MFA
│   ├── patient/        # Patient profile, insurance, history
│   ├── appointment/    # Scheduling, slots, reminders
│   ├── reception/      # Check-in, queue management
│   ├── visit/          # Outpatient visit, diagnosis, orders
│   ├── emr/            # Electronic Medical Record (MongoDB)
│   ├── lis/            # Laboratory: order→sample→result
│   ├── ris/            # Radiology: order→study→report
│   ├── pacs/           # DICOM storage & viewer metadata
│   ├── pharmacy/       # Prescription, dispensing
│   ├── inventory/      # Drug & supply warehouse
│   ├── billing/        # Invoice, payment, insurance
│   ├── inpatient/      # Ward, bed, admission
│   ├── notification/   # SMS, Email, Push (NATS consumer)
│   └── audit/          # Audit log (NATS consumer → MongoDB)
├── pkg/
│   ├── database/       # PG + Mongo connection factories
│   ├── cache/          # Redis wrapper
│   ├── messaging/      # NATS JetStream wrapper
│   ├── storage/        # MinIO/S3 wrapper
│   └── logger/         # Zerolog structured logging
├── web/                # React+Vite patient web app
├── desktop/            # Wails + React staff desktop app
├── migrations/         # SQL migration files (numbered)
├── configs/            # app.yaml, docker-compose.yml
├── deployments/
│   ├── docker/
│   └── k8s/            # (Phase 2+)
└── docs/               # Swagger, ERD, architecture diagrams
```

### DDD Module Structure (mỗi module)

```
internal/patient/
├── domain/
│   ├── entity/          # Aggregate roots, entities
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
│   ├── postgres/        # Repository implementations
│   ├── mongodb/
│   └── cache/
└── presentation/
    └── http/            # Fiber route handlers
```

---

## 2. DATABASE DESIGN

### PostgreSQL — Schema Groups (~150 bảng)

#### identity (~12 bảng)
```sql
users, roles, permissions, user_roles, role_permissions,
refresh_tokens, mfa_secrets, audit_sessions,
departments, staff_profiles, login_attempts, device_registry
```

#### patient (~18 bảng)
```sql
patients, patient_contacts, patient_insurance, patient_allergies,
patient_chronic_conditions, patient_documents, patient_identities,
patient_family_history, patient_social_history,
-- Reference:
blood_types, marital_statuses, occupations, ethnicities,
provinces, districts, wards_address, countries
```

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
// EMR (bệnh án điện tử)
medical_records {
  _id, patient_id, visit_id, version,
  soap: { subjective, objective, assessment, plan },
  vitals: { bp, hr, temp, spo2, weight, height },
  diagnoses: [...],
  orders: [...],
  prescriptions: [...],
  attachments: [...],
  created_by, updated_by, created_at, history: [...]
}

// LIS kết quả chi tiết
lab_result_documents { order_id, results: [...], raw_data, ... }

// RIS báo cáo chi tiết
radiology_documents { study_id, findings, impression, images: [...] }

// PACS metadata
dicom_metadata { study_uid, series: [...], instances: [...] }

// Audit trail
audit_logs { user_id, action, entity, entity_id, diff, ip, device, timestamp }

// Report snapshots
report_snapshots { type, period, data: {...}, generated_at }

// Notification
notification_templates { type, channel, subject, body_template }
notification_logs { template_id, recipient, status, sent_at }
```

---

## 3. SECURITY MODEL

### Authentication
- JWT Access Token: 15 phút
- Refresh Token: 7 ngày, Redis lưu + rotation
- TOTP MFA: compatible Google Authenticator
- Device fingerprint trong session

### RBAC Roles

| Role | Quyền chính |
|------|-------------|
| ADMIN | Toàn quyền |
| DOCTOR | Khám, EMR, chỉ định LIS/RIS, kê đơn |
| NURSE | Tiếp nhận, Vitals, chăm sóc |
| LAB_TECH | Nhận mẫu → kết quả |
| RADIOLOGIST | Đọc phim, viết báo cáo |
| PHARMACIST | Duyệt đơn, xuất thuốc |
| RECEPTIONIST | Đặt lịch, check-in, thu ngân |
| ACCOUNTANT | Xem hóa đơn, báo cáo tài chính |

### Data Security
- TLS 1.3 cho tất cả kết nối
- AES-256 field-level encryption: CCCD, số BHYT, SĐT
- PII masking trong log
- Backup daily + point-in-time recovery

---

## 4. EVENT-DRIVEN (NATS JetStream)

### Streams & Events

```
HIS.PATIENT.*
  PatientRegistered, PatientUpdated, PatientMerged

HIS.APPOINTMENT.*
  AppointmentScheduled, AppointmentConfirmed,
  AppointmentCancelled, AppointmentCheckedIn

HIS.VISIT.*
  VisitStarted, VisitDiagnosed,
  LabOrderCreated, RadiologyOrderCreated,
  PrescriptionCreated, VisitClosed

HIS.LIS.*
  SampleCollected, LabResultReady, LabResultVerified

HIS.RIS.*
  StudyScheduled, StudyCompleted, ReportSigned

HIS.BILLING.*
  InvoiceCreated, PaymentReceived, ClaimSubmitted

HIS.NOTIFICATION.*
  NotificationRequested
```

### Workers
- **notification_worker**: Subscribe events → gửi SMS/Email/Zalo
- **audit_worker**: Subscribe tất cả → ghi MongoDB audit_logs
- **report_worker**: Subscribe billing/visit → cập nhật snapshots

---

## 5. WEB APP — Bệnh Nhân (React + Vite)

### Pages & Features

```
/ — Landing: Giới thiệu phòng khám, bác sĩ, dịch vụ
/login — Đăng nhập / OTP qua SĐT
/register — Đăng ký tài khoản
/book — Đặt lịch: chọn dịch vụ → bác sĩ → ngày/giờ
/my-appointments — Lịch hẹn: upcoming, history, hủy
/results — Kết quả xét nghiệm (sau khi lab verify)
/account — Hồ sơ: thông tin cá nhân, BHYT, lịch sử khám
```

### Tech Stack Web
```
React 18 + TypeScript
Vite
TanStack Query (data fetching)
Zustand (global state)
React Hook Form + Zod (validation)
shadcn/ui + Tailwind CSS
React Router v6
i18next (VI/EN)
```

---

## 6. DESKTOP APP — Nhân Viên (Wails + React)

### Màn Hình Theo Role

**Lễ tân (Receptionist)**
- Queue Dashboard realtime (WebSocket)
- Đặt lịch thủ công / check-in walk-in
- Tìm kiếm bệnh nhân, tạo hồ sơ mới
- Thu ngân: tạo hóa đơn, nhận thanh toán, in biên lai

**Bác sĩ (Doctor)**
- Worklist: danh sách bệnh nhân theo ca
- Khám bệnh: Vitals → SOAP → ICD-10 → Chỉ định
- EMR editor: bệnh án có versioning
- Xem kết quả LIS/RIS inline
- Kê đơn thuốc + drug interaction check

**Dược sĩ (Pharmacist)**
- Prescription queue theo trạng thái
- Drug interaction checker realtime
- Xuất thuốc, in nhãn, ghi log
- Quản lý tồn kho thuốc

**Kỹ thuật viên XN (Lab Tech)**
- Sample tracking: nhận → xử lý → kết quả
- Nhập kết quả theo reference range
- Verify & approve, in phiếu
- Barcode scanner support

**Bác sĩ CĐHA (Radiologist)**
- Worklist chụp chiếu
- DICOM Viewer (Cornerstone.js embedded trong Wails)
- Soạn báo cáo theo template

**Admin**
- User & Role management
- Cấu hình hệ thống (giờ làm việc, dịch vụ, giá)
- Dashboard & Reporting
- Audit log viewer

### Tech Stack Desktop
```
Wails v2 + React 18 + TypeScript
TanStack Query + Zustand
Ant Design (phù hợp nghiệp vụ phức tạp)
Cornerstone.js (DICOM viewer)
Recharts (biểu đồ)
React-to-print (in ấn)
i18next
```

---

## 7. INFRASTRUCTURE

### Docker Compose (Development)

```yaml
services:
  api:        # Go app (air hot-reload)
  worker:     # NATS consumer
  web:        # React Vite (patient web)
  postgres:   # PostgreSQL 15
  mongodb:    # MongoDB 7
  redis:      # Redis 7
  nats:       # NATS JetStream
  minio:      # Object storage
  jaeger:     # Distributed tracing
  prometheus: # Metrics
  grafana:    # Dashboard
```

### Production (VPS — MVP)
- 1 VPS: Ubuntu 22.04, 4 CPU, 8GB RAM
- Docker Compose production mode
- Nginx reverse proxy + SSL (Let's Encrypt)
- Automated backup to S3/Backblaze

### CI/CD (GitHub Actions)
```
PR:   lint → unit test → integration test → build check
Main: all tests → build image → push to GHCR
Tag:  build → push → deploy to VPS (SSH + docker compose pull)
```

### Observability
```
Logging:  zerolog → stdout → Loki → Grafana
Metrics:  Prometheus → Grafana (response time, error rate, queue depth)
Tracing:  OpenTelemetry → Jaeger (request traces)
Alerts:   Grafana alerting → Telegram/Email
```

---

## 8. ROADMAP CHI TIẾT

### PHASE 1 — MVP Phòng Khám Tư (Tuần 1–16)

> Solo developer · 2 tuần/sprint · Web + Desktop

#### Sprint 1 (Tuần 1–2): Foundation
**Backend:**
- [ ] Init Go module, Fiber, Air hot-reload
- [ ] Docker Compose: PG, Mongo, Redis, NATS, MinIO
- [ ] DB connection factories (pgxpool, mongo driver, redis)
- [ ] Zerolog + OpenTelemetry setup
- [ ] golang-migrate + migration scripts
- [ ] Base error types, response wrapper, pagination
- [ ] Makefile: run, test, migrate, lint
- [ ] GitHub Actions CI cơ bản
- [ ] Health check `/health`, `/ready`
- [ ] Swagger (swaggo) setup

**Frontend (cả 2 apps):**
- [ ] Init Wails v2 + React + TypeScript (desktop)
- [ ] Init React + Vite + TypeScript (web)
- [ ] Design system: color palette, typography, spacing
- [ ] Layout cơ bản: Sidebar, Header, Content
- [ ] Shared API client (axios + TanStack Query)

**Database migrations:**
- [ ] 001_identity.sql
- [ ] 002_patient.sql
- [ ] 003_appointment.sql
- [ ] Seed: roles, permissions, ICD-10 codes (basic), demo data

---

#### Sprint 2 (Tuần 3–4): Identity & Auth
**Backend:**
- [ ] `internal/identity`: User entity, Role, Permission
- [ ] JWT access + refresh token (Redis)
- [ ] TOTP MFA (pquerna/otp)
- [ ] RBAC middleware (permission check per route)
- [ ] APIs: login, logout, refresh, MFA setup/verify
- [ ] User CRUD: create, list, update, deactivate
- [ ] Department management
- [ ] Password: bcrypt, reset flow via email/SMS
- [ ] Rate limiting on auth endpoints

**Frontend:**
- [ ] Desktop: Login screen, MFA screen
- [ ] Desktop: User management (Admin)
- [ ] Desktop: Role & Permission matrix UI
- [ ] Web: Login/Register (OTP via SĐT)

---

#### Sprint 3 (Tuần 5–6): Patient & Appointment
**Backend:**
- [ ] `internal/patient`: Patient aggregate, Insurance, Contacts
- [ ] Patient CRUD + search (full-text, CCCD, SĐT)
- [ ] Insurance validation logic (BHYT format)
- [ ] `internal/appointment`: Slot templates, Doctor schedules
- [ ] Appointment CRUD: book, cancel, reschedule
- [ ] Conflict detection (double booking)
- [ ] `PatientRegistered`, `AppointmentScheduled` events

**Frontend:**
- [ ] Desktop: Patient search & registration form
- [ ] Desktop: Appointment calendar view (doctor schedule)
- [ ] Web: Booking flow (service → doctor → slot → confirm)
- [ ] Web: My appointments page

---

#### Sprint 4 (Tuần 7–8): Reception & Visit
**Backend:**
- [ ] `internal/reception`: Check-in, Queue
- [ ] Queue number generation (per service type)
- [ ] WebSocket endpoint: realtime queue update
- [ ] `internal/visit`: Visit creation, Vitals, Status machine
- [ ] Visit orders (lab, radiology, procedure)
- [ ] ICD-10 search API (full catalog seeded)
- [ ] `VisitStarted`, `LabOrderCreated` events

**Frontend:**
- [ ] Desktop: Queue dashboard (realtime, số thứ tự)
- [ ] Desktop: Check-in flow (scan CCCD / search patient)
- [ ] Desktop: Doctor worklist (bệnh nhân chờ khám)
- [ ] Desktop: Visit screen (Vitals form)

---

#### Sprint 5 (Tuần 9–10): EMR
**Backend:**
- [ ] `internal/emr`: Medical record MongoDB schema
- [ ] SOAP note editor API (create, update with versioning)
- [ ] Vitals history API
- [ ] Diagnosis finalize + ICD-10 link
- [ ] Attachment upload (MinIO)
- [ ] EMR audit: ai sửa, sửa gì, version history

**Frontend:**
- [ ] Desktop: SOAP editor (rich text hoặc structured form)
- [ ] Desktop: Diagnosis picker (ICD-10 search)
- [ ] Desktop: Order creation UI (lab/radiology/prescription)
- [ ] Desktop: Patient history timeline
- [ ] Web: Patient view own results (read-only)

---

#### Sprint 6 (Tuần 11–12): Pharmacy & Inventory
**Backend:**
- [ ] `internal/pharmacy`: Drug catalog, Prescription
- [ ] Drug interaction check (local DB rules)
- [ ] Prescription CRUD + status workflow
- [ ] Dispensing record
- [ ] `internal/inventory`: Warehouse, Stock, Transactions
- [ ] Stock deduction on dispense
- [ ] Low stock alert event

**Frontend:**
- [ ] Desktop: Drug search + Prescription editor
- [ ] Desktop: Drug interaction warning modal
- [ ] Desktop: Pharmacist dispensing queue
- [ ] Desktop: Inventory stock view

---

#### Sprint 7 (Tuần 13–14): Billing & Notification
**Backend:**
- [ ] `internal/billing`: Price catalog, Invoice, Payment
- [ ] Auto-create invoice on visit close
- [ ] Payment methods: cash, transfer (manual)
- [ ] Invoice PDF generation (chromedp)
- [ ] `internal/notification`: NATS consumer worker
- [ ] SMS gateway integration (VNPT/Twilio)
- [ ] Email (SMTP/SendGrid)
- [ ] Notification triggers: appointment reminder, result ready

**Frontend:**
- [ ] Desktop: Billing screen (hóa đơn, thanh toán)
- [ ] Desktop: In hóa đơn / biên lai
- [ ] Desktop: Payment history
- [ ] Web: Invoice view cho bệnh nhân

---

#### Sprint 8 (Tuần 15–16): Audit, Dashboard, Polish
**Backend:**
- [ ] `internal/audit`: NATS consumer → MongoDB audit_logs
- [ ] Reporting API: revenue summary, patient count, top services
- [ ] Report snapshot worker
- [ ] Integration test suite (testify + httptest)
- [ ] Load test cơ bản (k6)
- [ ] API documentation hoàn chỉnh (Swagger)

**Frontend:**
- [ ] Desktop: Admin dashboard (Recharts)
- [ ] Desktop: Audit log viewer
- [ ] Desktop: System config UI
- [ ] Web: Final polish, responsive
- [ ] E2E test cơ bản

**DevOps:**
- [ ] Production Docker Compose
- [ ] Nginx config + SSL
- [ ] Backup script (PostgreSQL dump + MongoDB dump)
- [ ] Monitoring: Prometheus + Grafana dashboard

---

### PHASE 2 — Inpatient & Advanced (Tuần 17–32)

| Module | Nội dung |
|--------|----------|
| Inpatient | Ward/Room/Bed management, Admission, Discharge |
| Nursing | Nursing notes, Medication administration, Vitals charting |
| Advanced Billing | BHYT manual claim, copayment calculation |
| Advanced Inventory | Purchase order, Supplier management, Expiry alerts |
| Reporting | Advanced dashboard, export Excel/PDF |
| Shift Management | Ca trực, bàn giao ca |
| Queue Enhancement | TV display app (WebSocket), TTS gọi số |
| Print Enhancement | Barcode/QR, label printer, thermal printer |

---

### PHASE 3 — LIS & RIS Full (Tuần 33–48)

| Module | Nội dung |
|--------|----------|
| LIS Full | Machine interface (HL7 v2.x), Auto-result import |
| RIS Full | Scheduling, Worklist management |
| PACS | DICOM storage (MinIO), Viewer (OHIF/Cornerstone.js) |
| PACS Protocol | WADO-RS, STOW-RS, QIDO-RS |
| Drug DB | DrugBank integration, Advanced interaction check |

---

### PHASE 4 — Digital & Scale (Tuần 49–64)

| Module | Nội dung |
|--------|----------|
| Patient Portal Web | Nâng cấp web app đầy đủ tính năng |
| Telehealth | Video consultation (WebRTC) |
| Mobile App | React Native: bệnh nhân đặt lịch, xem kết quả |
| Multi-clinic | Multi-tenant: nhiều cơ sở, shared patient |
| Kubernetes | Migrate từ Docker Compose sang K8s |
| Microservices | Extract LIS, PACS thành service riêng |

---

### PHASE 5 — Compliance & Integration (Tuần 65+)

| Module | Nội dung |
|--------|----------|
| EMR chuẩn | Thông tư 46/2018 Bộ Y tế |
| Chữ ký số | VNPT-CA, Viettel-CA |
| BHYT điện tử | API cổng giám định BHXH |
| FHIR R4 | FHIR API endpoint (Patient, Encounter, Observation...) |
| HL7 v2.x | Interface engine với thiết bị xét nghiệm |
| HIS quốc gia | Kết nối VN HIS nếu có yêu cầu |

---

## 9. CHIẾN LƯỢC SOLO DEVELOPER

### Nguyên tắc

```
✅ Làm:
  - DDD nhưng không over-engineer từ đầu
  - YAGNI: chỉ build thứ cần cho sprint hiện tại
  - Unit test domain layer (business logic)
  - Docker Compose cho MVP, VPS đơn giản
  - Swagger auto-gen từ code
  - Commit thường xuyên, 1 branch per feature

❌ Tránh:
  - Microservices quá sớm
  - K8s cho MVP (overkill)
  - Build nhiều module song song
  - Perfection over done
```

### Tooling

```
air              — Go hot reload
sqlc             — Type-safe SQL codegen
swaggo/swag      — Swagger auto-gen
golang-migrate   — DB migrations
golangci-lint    — Linter
testify/mockery  — Testing + mocks
Makefile         — Task automation
```

### Makefile Targets

```makefile
make dev         # Start all Docker services + air
make migrate     # Run pending migrations
make test        # Run all tests
make lint        # Run golangci-lint
make swag        # Generate Swagger docs
make build       # Build binary
make deploy      # Deploy to VPS
```

---

## 10. RỦI RO & GIẢM THIỂU

| Rủi ro | Mức độ | Biện pháp |
|--------|--------|-----------|
| Nghiệp vụ y tế phức tạp, sai quy trình | Cao | Validate với bác sĩ/quản lý phòng khám trước khi code |
| Solo dev burnout | Cao | Scope MVP nhỏ, iterate nhanh, ưu tiên nghiêm ngặt |
| BHYT integration phức tạp | Cao | MVP chỉ ghi nhận manual, tự động hóa sau |
| Data loss | Cao | Daily backup + off-site, point-in-time recovery |
| Downtime ảnh hưởng bệnh nhân | Trung bình | Health check, restart policy, offline cache |
| DICOM storage cost | Trung bình | MinIO self-hosted, tiered storage sau |
| Regulatory compliance | Cao | Bám sát TT 46/2018, chuẩn bị từ Phase 5 |

---

## 11. OPEN QUESTIONS

1. **Offline**: Phòng khám có internet ổn định? Cần offline cache trong Wails?
2. **SMS Provider**: VNPT SMS, Twilio, hay Zalo ZNS?
3. **Deployment**: VPS cloud hay server on-premise tại phòng khám?
4. **Multi-tenant**: MVP cho 1 cơ sở hay nhiều cơ sở ngay từ đầu?
5. **Payment gateway**: VNPay/MoMo, hay chỉ ghi nhận tiền mặt/chuyển khoản?
