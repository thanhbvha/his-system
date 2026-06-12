# Sprint 1 — Báo cáo Hoàn thành (Completion Report)

> **Sprint:** 1 — Core Foundation & Architecture  
> **Thời gian:** Tuần 1–2  
> **Trạng thái:** ✅ HOÀN THÀNH  
> **Ngày tổng hợp:** 2026-06-12

---

## 📌 Tóm tắt tổng quan

Sprint 1 thiết lập toàn bộ nền tảng kỹ thuật (Technical Foundation) cho hệ thống HIS. Kết thúc sprint, dự án có đầy đủ 3 tầng ứng dụng hoạt động song song:

| Tầng | Công nghệ | Mục đích |
|------|-----------|----------|
| **Backend API** | Go + Fiber v2 | REST API server, Business Logic |
| **Desktop App** | Wails v2 + React + Ant Design | Giao diện nội bộ (bác sĩ, y tá, nhân viên) |
| **Web App** | Vite + React + Tailwind + shadcn/ui | Cổng thông tin bệnh nhân (đặt lịch) |

---

## ✅ Step 1 — Khởi tạo Project & Infrastructure

**Mục tiêu:** Dựng skeleton Go project và toàn bộ infra local bằng Docker Compose.

### Đã hoàn thành:
- `go mod init his-system` + cài đặt tất cả core dependencies (Fiber v2, pgx, MongoDB driver, `thanhbvha/go-common`)
- Tạo cấu trúc thư mục chuẩn: `cmd/`, `internal/`, `pkg/`, `migrations/`, `docs/`, `scripts/`
- `Makefile` với đầy đủ targets: `dev`, `migrate`, `test`, `lint`, `swag`, `build`
- **Docker Compose** với 7 services đầy đủ:

| Service | Image | Port |
|---------|-------|------|
| PostgreSQL | `postgres:15-alpine` | 5432 |
| MongoDB | `mongo:7` | 27017 |
| Redis | `redis:7-alpine` | 6379 |
| MinIO | `minio/minio` | 9000, 9001 |
| Jaeger | `jaegertracing/all-in-one` | 16686, 4317 |
| Prometheus | `prom/prometheus` | 9090 |
| Grafana | `grafana/grafana` | 3000 |

- Health check cho từng service (auto-retry 5 lần)
- File `.env` đầy đủ biến môi trường (DSN, Keys, Endpoints)
- Cổng API được load từ `.env` (`APP_PORT`)

> ⚠️ **Quyết định kiến trúc:** Không có service NATS riêng — Redis Stream (`go-common/queue`) đủ đáp ứng yêu cầu hiện tại, giúp giảm thiểu gánh nặng vận hành.

---

## ✅ Step 2 — Database & Migration

**Mục tiêu:** Thiết lập schema database cốt lõi và seed data.

### Đã hoàn thành:
- `cmd/migrate/main.go` — CLI migration runner (`--up`, `--down`, `--steps`)
- **3 file migration** chạy thành công, idempotent:

**`001_identity.sql`** — 9 bảng:
- `users`, `roles`, `permissions`, `user_roles`, `role_permissions`
- `refresh_tokens`, `mfa_secrets`, `login_attempts`
- `device_registry`, `departments`, `staff_profiles`, `audit_sessions`

**`002_patient.sql`** — 3 bảng với field-level encryption:
- `patients` (cột `phone_encrypted`, `phone_hmac`, `cccd_encrypted`, `cccd_hmac`, ...)
- `patient_contacts`, `patient_insurance`

> ⚠️ **Pattern bảo mật:** Cột `*_hmac` để tìm kiếm (lookup) không cần giải mã. Cột `*_encrypted` lưu AES-GCM ciphertext.

**`003_appointment.sql`** — 4 bảng:
- `appointments`, `appointment_slots`, `slot_templates`, `doctor_schedules`

**Seed data:**
- 6 roles mặc định: `admin`, `doctor`, `nurse`, `receptionist`, `pharmacist`, `patient`
- Permissions cơ bản cho từng role
- ICD-10 top 100 mã bệnh phổ biến
- Demo admin user: `admin` / `Admin@123`

---

## ✅ Step 3 — Security Package (Crypto)

**Mục tiêu:** Xây dựng `pkg/crypto` — layer mã hoá AES-GCM và HMAC cho toàn hệ thống.

### Đã hoàn thành:
- **`pkg/crypto/aes_gcm.go`**: `EncryptAESGCM` / `DecryptAESGCM`
  - Nonce 96-bit random mỗi lần encrypt
  - Output: `nonce(12B) || ciphertext || tag(16B)` → base64 URL-safe
- **`pkg/crypto/field_cipher.go`**: `FieldCipher` wrapper
  - `Encrypt(plaintext string)` / `Decrypt(ciphertext string)`
  - `HMAC(value string)` — deterministic, dùng cho search
- **`pkg/crypto/config.go`**: Load key từ env (`FIELD_ENCRYPTION_KEY`, `FIELD_HMAC_KEY`)
- **`scripts/gen-crypto-keys.sh`**: Helper tạo 256-bit random keys

**Unit Tests — 100% PASS:**

| Test | Kết quả |
|------|---------|
| `TestEncryptDecryptRoundTrip` | ✅ PASS |
| `TestEncryptProducesUniqueCiphertexts` | ✅ PASS |
| `TestInvalidKeySize` | ✅ PASS |
| `TestTamperedCiphertextDetected` | ✅ PASS |
| `TestHMACIsDeterministic` | ✅ PASS |
| `TestFieldCipherRoundTrip` | ✅ PASS |
| `TestEncryptEmptyString` | ✅ PASS |
| **Coverage** | **≥ 95%** |

---

## ✅ Step 4 — Core Framework & go-common Integration

**Mục tiêu:** Tích hợp `github.com/thanhbvha/go-common` và xây dựng các shared package.

### Đã hoàn thành:

**`pkg/database/`**
- `postgres.go`: `NewPostgresPool` / `MustNewPostgresPool` với connection pool config
- `mongo.go`: `NewMongoClient` / `MustNewMongoClient`

**`pkg/storage/minio.go`**
- `StorageClient` interface: `Upload`, `Download`, `Delete`, `GetURL`

**`pkg/utils/safego.go`**
- `SafeGo(fn func())` — goroutine an toàn, bắt panic không sập server

**`pkg/middleware/`**
- `recover.go` — bắt panic, log async, trả 500 chuẩn API
- `logger.go` — log `method`, `path`, `status`, `duration_ms` mỗi request

**`pkg/errors/errors.go`** — Error types chuẩn:
- `ErrNotFound (404)`, `ErrValidation (422)`, `ErrUnauthorized (401)`
- `ErrForbidden (403)`, `ErrConflict (409)`, `ErrInternal (500)`

**`pkg/response/response.go`** — Response wrapper chuẩn:
- `Response{Success, Data, Error, Meta}` — JSON đồng nhất cho toàn API
- `OK()`, `OKWithMeta()`, `Fail()` helpers

**go-common Integrations:**
| Package | Quyết định |
|---------|-----------|
| `go-common/logger` | Dùng trực tiếp, cấu hình `InfoAsync`/`ErrorAsync` 100% |
| `go-common/redis` | Dùng native client, KHÔNG bọc wrapper ngoài (dư thừa) |
| `go-common/queue` | Dùng trực tiếp Redis Streams (Enqueue, Consumer Groups, DLQ) |
| `go-common/websocket` | Dùng `adapter/fiber` trực tiếp, bỏ tự viết hub |

**`cmd/api/main.go` Bootstrap sequence:**
1. Load env → 2. Init Logger → 3. Init DB (PG + Mongo) → 4. Init Redis → 5. Init Queue → 6. Init WebSocket Manager → 7. Init Storage → 8. Bootstrap Fiber → 9. Register Middleware → 10. Register Routes → 11. Graceful Start

---

## ✅ Step 5 — Observability & API Foundation

**Mục tiêu:** Setup monitoring/tracing và API endpoints nền tảng.

### Đã hoàn thành:

**Observability Stack:**
- **OpenTelemetry + Jaeger**: `pkg/telemetry/tracer.go` — auto-span mỗi request
- **Prometheus**: `fiberprometheus` — `http_requests_total`, `http_request_duration_seconds`
- **Structured Logging (JSON)**: PII tự động masked (`phone → ***`, `cccd → ***`)

**API Endpoints nền tảng:**
| Endpoint | Mô tả |
|----------|-------|
| `GET /health` | Liveness check → `{"status":"ok","version":"1.0.0"}` |
| `GET /ready` | Readiness check PG + Mongo + Redis, trả 503 nếu fail |
| `GET /metrics` | Prometheus metrics |
| `GET /docs/tool` | Swagger UI (internal CSS/JS) |
| `GET /docs` | ReDoc UI (internal CSS/JS) |
| `GET /ws` | WebSocket endpoint |

**Middleware Stack (theo thứ tự):**
1. Request ID (UUID v4 → header + log)
2. Request Logger (custom, bỏ qua `/health`)
3. OpenTelemetry Tracer
4. CORS (whitelist `localhost:5173`)
5. Rate Limiter (100 req/min per IP → 429 JSON)
6. Recover (panic → 500 JSON)

---

## 📝 Step 6 — CI/CD (Ghi nhận để thiết lập sau)

**Trạng thái:** Đã lên kế hoạch đầy đủ, **chưa thiết lập** (sẽ làm khi có GitHub repository).

**Pipeline đã thiết kế** (`.github/workflows/ci.yml`):
- Trigger: push/PR vào `main`, `develop`
- Services: PostgreSQL 15 + Redis 7 trong CI runner
- Steps: `go mod download` → `golangci-lint` → `go test -race -cover` → `go build`
- Tích hợp Codecov cho coverage report

---

## ✅ Step 7 — Desktop Application (Wails + React + Ant Design)

**Mục tiêu:** Khởi tạo ứng dụng Desktop native cho nội bộ bệnh viện.

### Đã hoàn thành:

**Bootstrapping:**
- Khởi tạo Wails v2 project với `react-ts` template (`desktop/`)
- `wails dev` chạy thành công, cửa sổ native hiển thị

**Enterprise Architecture Setup:**
- Path alias `@/` → `src/` (vite.config + tsconfig)
- Ant Design `ConfigProvider` với theme màu y tế chính (`#1677ff`)
- Google Fonts **Inter** tích hợp qua `index.html`

**Core Libraries:**
| Library | Mục đích |
|---------|---------|
| `antd` | UI Component System |
| `axios` | HTTP Client |
| `@tanstack/react-query` | Server State & Caching |
| `zustand` | Client State Management |
| `react-router-dom` | Routing |
| `i18next` / `react-i18next` | Đa ngôn ngữ (VI/EN) |

**Files tạo mới:**
```
desktop/frontend/src/
├── lib/
│   ├── apiClient.ts      # Axios + JWT Bearer interceptor
│   └── queryClient.ts    # TanStack Query config (staleTime 5m)
├── store/
│   ├── authStore.ts      # token, user, role, setAuth, clearAuth
│   └── uiStore.ts        # sidebarOpen, toggleSidebar
├── layouts/
│   └── RoleLayout.tsx    # Dynamic sidebar theo role (admin/doctor/nurse/...)
├── components/
│   └── ProtectedRoute.tsx
├── i18n/
│   ├── index.ts
│   ├── vi.json
│   └── en.json
└── pages/ (placeholder)
```

**Verification:** `tsc && vite build` pass không lỗi ✅

---

## ✅ Step 8 — Web Application (Vite + React + Tailwind + shadcn/ui)

**Mục tiêu:** Khởi tạo Patient-facing Web App.

### Đã hoàn thành:

**Bootstrapping:**
- Khởi tạo Vite project `react-ts` (`web/`)
- Downgrade Vite xuống `5.4.x` để tương thích Node.js `20.18.x` hiện tại
- `npm run build` → bundle production `~114KB` (gzipped) ✅

**Design System:**
- **Tailwind CSS v4** với `@tailwindcss/vite` plugin (không cần `tailwind.config.ts`)
- **shadcn/ui** khởi tạo thành công (`components.json`, `lib/utils.ts`)
- Font **Inter** (Geist) tích hợp tự động qua shadcn
- Màu primary: Medical Blue (`--primary: blue-600`)

**Core Setup:**
| File | Mô tả |
|------|-------|
| `src/lib/apiClient.ts` | Axios + JWT interceptor + auto-clearAuth on 401 |
| `src/lib/queryClient.ts` | TanStack Query (staleTime 5m, retry 2) |
| `src/store/authStore.ts` | `token`, `patient`, `setAuth`, `clearAuth` |
| `src/store/bookingStore.ts` | Multi-step booking state (step 1→4, department, doctor, slot, patientInfo) |

**Dev Proxy:**
- `vite.config.ts`: `/api` → `http://localhost:8080` (giải quyết CORS dev)
- `.env.development` / `.env.production` tách biệt rõ ràng

**Routing & Layout:**
```
/ (PublicLayout)
├── /           → LandingPage    (Hero section + Đặt khám CTA)
├── /login      → LoginPage      (Mock login Sprint 1)
└── /register   → RegisterPage   (Coming soon)

/ (ProtectedRoute → AuthLayout)
├── /book           → BookingPage        (4-step form placeholder)
├── /my-appointments → MyAppointmentsPage
├── /results        → ResultsPage
└── /account        → AccountPage        (Patient profile)
```

**i18n:** VI (mặc định) + EN với Language Switcher trên Header

---

## 🗂️ Cấu trúc dự án sau Sprint 1

```
his-system/
├── backend/                # Go Backend API
│   ├── cmd/
│   │   ├── api/main.go     # Entry point, Fiber bootstrap
│   │   ├── worker/         # Redis Stream consumer
│   │   └── migrate/        # Migration CLI
│   ├── internal/
│   │   ├── api/            # Route handlers
│   │   ├── identity/       # (scaffold, Sprint 2)
│   │   └── patient/        # (scaffold, Sprint 3)
│   ├── pkg/
│   │   ├── crypto/         # AES-GCM, FieldCipher ✅ tested
│   │   ├── database/       # PG pool, MongoDB client
│   │   ├── storage/        # MinIO client
│   │   ├── telemetry/      # OpenTelemetry tracer
│   │   ├── errors/         # AppError types
│   │   ├── response/       # JSON response wrapper
│   │   ├── middleware/      # Logger, Recover
│   │   ├── logger/         # PII masking wrapper
│   │   └── utils/          # SafeGo
│   ├── migrations/postgres/ # 3 migration files
│   ├── docs/               # Swagger spec (auto-generated)
│   ├── scripts/            # gen-crypto-keys.sh
│   ├── docker-compose.yml
│   ├── .env
│   ├── go.mod
│   └── Makefile
│
├── desktop/                # Wails Native Desktop App
│   ├── main.go
│   ├── app.go
│   ├── wails.json
│   └── frontend/           # React + Ant Design
│       └── src/
│           ├── lib/        # apiClient, queryClient
│           ├── store/      # authStore, uiStore
│           ├── layouts/    # RoleLayout
│           ├── components/ # ProtectedRoute
│           └── i18n/       # vi.json, en.json
│
├── web/                    # Vite Patient Web App
│   ├── src/
│   │   ├── lib/            # apiClient, queryClient
│   │   ├── store/          # authStore, bookingStore
│   │   ├── layouts/        # PublicLayout, AuthLayout
│   │   ├── components/     # ProtectedRoute, shadcn/ui
│   │   ├── pages/          # 7 pages placeholder
│   │   └── i18n/           # vi.json, en.json
│   ├── .env.development
│   ├── .env.production
│   ├── vite.config.ts
│   └── components.json
│
├── plan/
│   ├── sprints/sprint1/    # Tài liệu kế hoạch (8 steps)
│   └── report_sprint1.md  # File này
│
└── README.md               # Hướng dẫn chạy tổng hợp
```

---

## 📊 Definition of Done — Tổng hợp

| Step | Mô tả | Trạng thái |
|------|-------|------------|
| Step 1 | Project init + Docker Compose infrastructure | ✅ Done |
| Step 2 | Database migration (001–003) + Seed data | ✅ Done |
| Step 3 | `pkg/crypto` AES-GCM + FieldCipher + Unit Tests ≥95% | ✅ Done |
| Step 4 | Core framework + go-common integration + Fiber bootstrap | ✅ Done |
| Step 5 | Observability (OTel/Jaeger/Prometheus) + API endpoints | ✅ Done |
| Step 6 | CI/CD pipeline design | 📝 Planned (thiết lập khi có GitHub repo) |
| Step 7 | Desktop App (Wails + React + Ant Design) | ✅ Done |
| Step 8 | Web App (Vite + React + Tailwind + shadcn/ui) | ✅ Done |

---

## 🔍 Ghi chú kỹ thuật quan trọng

### Quyết định kiến trúc nổi bật trong Sprint 1:

1. **Không dùng wrapper thừa:** `pkg/cache/cache.go` và `pkg/messaging/messaging.go` đã bị loại bỏ. Dùng `go-common/redis` và `go-common/queue` trực tiếp để tận dụng toàn bộ sức mạnh native.

2. **API Port từ .env:** Cổng API server (`APP_PORT`, mặc định `8080`) được load từ `.env` tránh conflict với Grafana (port 3000).

3. **Module path giữ nguyên:** Sau khi tái cấu trúc thư mục (gộp vào `backend/`), `go.mod` vẫn giữ `module his-system` ở root. Import paths như `his-system/pkg/...` vẫn hoạt động bình thường.

4. **go-common/websocket đánh giá:** Architecture Sharding (`maxShards=1000`, `maxTotalConnections=200,000` per node) + Redis PubSub cross-node đủ đáp ứng hàng chục ngàn events/giây cho quy mô HIS bệnh viện. **Không cần NATS** ở giai đoạn hiện tại.

5. **Node.js compatibility:** Web app dùng Vite `5.4.x` thay vì latest `8.x` để tương thích Node.js `20.18.x` (latest yêu cầu `≥20.19.0`).

---

## 🚀 Chuẩn bị cho Sprint 2

Sprint 2 tiếp theo: **Identity & Auth (Tuần 3–4)**

**Prerequisite từ Sprint 1 đã sẵn sàng:**
- ✅ Schema bảng `users`, `roles`, `permissions`, `device_registry`, `mfa_secrets` đã migrate
- ✅ `pkg/crypto` (AES-GCM + FieldCipher) sẵn sàng mã hoá email trong `users`
- ✅ Redis client sẵn sàng lưu `refresh_token` (TTL 7d) và OTP (TTL 5m)
- ✅ Queue sẵn sàng gửi email/OTP async
- ✅ Fiber middleware stack (CORS, Rate Limit, Request ID) sẵn sàng cho auth routes
- ✅ Desktop và Web app đã có `authStore` và `ProtectedRoute` chờ kết nối API thật
