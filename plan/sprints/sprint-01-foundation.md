# Sprint 1 — Foundation (Tuần 1–2)

> **Mục tiêu:** Dựng toàn bộ nền tảng dự án: backend skeleton, infra local, crypto package, và khởi tạo 2 app client.
> **Kết thúc sprint:** Hệ thống chạy được local, health check OK, 2 app client hiển thị layout cơ bản.

---

## BACKEND

### Khởi tạo project
- [ ] `go mod init his-system`
- [ ] Cài Fiber v2, zerolog, pgxpool, mongo-driver, redis client, NATS client
- [ ] Cấu trúc thư mục theo plan: `cmd/`, `internal/`, `pkg/`
- [ ] `Makefile` với targets: `dev`, `migrate`, `test`, `lint`, `swag`, `build`

### Infrastructure (Docker Compose)
- [ ] `docker-compose.yml` với các service:
  - `postgres:15`, `mongodb:7`, `redis:7`
  - `nats` (JetStream enabled)
  - `minio`, `jaeger`, `prometheus`, `grafana`
- [ ] Volume + network config
- [ ] Health check cho từng service

### Database & Migration
- [ ] Cài `golang-migrate`
- [ ] `001_identity.sql` — bảng: users, roles, permissions, user_roles, role_permissions, refresh_tokens, mfa_secrets, login_attempts, device_registry, departments, staff_profiles, audit_sessions
- [ ] `002_patient.sql` — bảng: patients (có cột `phone_encrypted`, `phone_hmac`, `cccd_encrypted`, `cccd_hmac`, `email_encrypted`, `email_hmac`), patient_contacts, patient_insurance, v.v.
- [ ] `003_appointment.sql` — bảng: appointments, appointment_slots, slot_templates, doctor_schedules, v.v.
- [ ] Seed data: roles, permissions, ICD-10 cơ bản, demo user

> ⚠️ **NOTE:** Migration phải idempotent — chạy nhiều lần không lỗi. Dùng `IF NOT EXISTS`.

### Security Package (ưu tiên cao)
- [ ] `pkg/crypto/aes_gcm.go`
  ```go
  func EncryptAESGCM(plaintext, key []byte) ([]byte, error)
  func DecryptAESGCM(ciphertext, key []byte) ([]byte, error)
  ```
  - Nonce 96-bit random mỗi lần
  - Output format: `nonce(12B) | ciphertext | tag(16B)` → base64
  - Key 256-bit từ env `FIELD_ENCRYPTION_KEY`
- [ ] `pkg/crypto/field_cipher.go`
  ```go
  type FieldCipher struct { key []byte }
  func (f *FieldCipher) Encrypt(plaintext string) (string, error)
  func (f *FieldCipher) Decrypt(ciphertext string) (string, error)
  func (f *FieldCipher) HMAC(value string) string
  ```
- [ ] Unit test cho cả 2 package: round-trip encrypt/decrypt, HMAC deterministic

> ⚠️ **NOTE:** `pkg/crypto` phải xong và pass test 100% trước khi viết bất kỳ module nào khác.

### Core Framework
- [ ] `pkg/database/` — pgxpool factory, mongo client factory
- [ ] `pkg/cache/` — Redis wrapper (Set, Get, Del, SetEX)
- [ ] `pkg/messaging/` — NATS JetStream wrapper (Publish, Subscribe)
- [ ] `pkg/storage/` — MinIO wrapper (Upload, Download, Delete)
- [ ] `pkg/logger/` — Zerolog structured logger với PII masking hook
- [ ] Base error types (`pkg/errors/`): NotFound, Validation, Unauthorized, Conflict
- [ ] Response wrapper: `{ success, data, error, meta }`
- [ ] Pagination helper: `{ page, limit, total, items }`
- [ ] `cmd/api/main.go` — Fiber app bootstrap

### Observability
- [ ] OpenTelemetry setup → Jaeger exporter
- [ ] Prometheus middleware cho Fiber (latency, status code metrics)
- [ ] Zerolog → stdout (Docker log driver → Loki sau)

### API Foundation
- [ ] `GET /health` — liveness check
- [ ] `GET /ready` — readiness check (kiểm tra PG, Mongo, Redis connect)
- [ ] Swagger (swaggo) setup + `/swagger/*` route
- [ ] CORS middleware config
- [ ] Rate limit middleware (global)
- [ ] Request ID middleware

### GitHub Actions CI
- [ ] `.github/workflows/ci.yml`: lint (`golangci-lint`) + test + build check

---

## DESKTOP (Wails + React)

### Khởi tạo project
- [ ] `wails init -n desktop -t react-ts` trong thư mục `desktop/`
- [ ] Cài dependencies: `axios`, `@tanstack/react-query`, `zustand`, `antd`, `react-router-dom`, `zod`, `i18next`
- [ ] Cấu hình Ant Design 5.x theme token (màu primary, font)
- [ ] Setup alias `@/` → `src/`

### Design System
- [ ] Ant Design theme config: màu primary `#1677ff` (y tế), border radius, font Inter/Roboto
- [ ] Global CSS reset + custom variables
- [ ] Typography scale: h1–h4, body, caption

### Core Setup
- [ ] `src/lib/apiClient.ts` — Axios instance, baseURL từ env/Wails config
  - Attach `Authorization: Bearer {token}` header
  - Interceptor placeholder cho auto-refresh (implement Sprint 2)
- [ ] `src/lib/queryClient.ts` — TanStack Query config: staleTime, retry
- [ ] `src/lib/websocket.ts` — WS client skeleton (connect/disconnect/on)
- [ ] `src/store/authStore.ts` — Zustand: `{ token, user, role, setAuth, clearAuth }`
- [ ] `src/store/uiStore.ts` — Zustand: `{ sidebarOpen, toggleSidebar }`

### Routing & Layout
- [ ] React Router v6 setup
- [ ] `src/layouts/RoleLayout.tsx` — Sidebar + Header + Content (layout chung)
- [ ] Sidebar items render theo `role` từ authStore
- [ ] Route guard: redirect `/login` nếu chưa auth
- [ ] Placeholder pages cho từng role (hiển thị "Coming soon")

### i18n
- [ ] i18next setup, file `vi.json` và `en.json`
- [ ] Language switcher component

### Dev Tools
- [ ] Wails dev mode với hot-reload React

---

## WEB (React + Vite)

### Khởi tạo project
- [ ] `npm create vite@latest web -- --template react-ts` trong thư mục `web/`
- [ ] Cài dependencies: `axios`, `@tanstack/react-query`, `zustand`, `react-hook-form`, `zod`, `@hookform/resolvers`, `react-router-dom`, `i18next`, `tailwindcss`, shadcn/ui
- [ ] Tailwind CSS + shadcn/ui init

### Design System
- [ ] shadcn/ui theme: màu primary dạng y tế (blue-600), border radius
- [ ] Font: Inter (Google Fonts)
- [ ] Global CSS variables cho color tokens
- [ ] Responsive breakpoints

### Core Setup
- [ ] `src/lib/apiClient.ts` — Axios, `withCredentials: true` (cho cookie refresh)
  - baseURL: `import.meta.env.VITE_API_URL`
  - Interceptor placeholder
- [ ] `src/lib/queryClient.ts`
- [ ] `src/store/authStore.ts` — Zustand: `{ token, patient, setAuth, clearAuth }`
- [ ] `src/store/bookingStore.ts` — Zustand: multi-step booking state

### Routing & Layout
- [ ] React Router v6 setup
- [ ] `PublicLayout.tsx` — Header (logo, nav, login button) + Footer
- [ ] `AuthLayout.tsx` — Header (logo, nav, user avatar) + Footer
- [ ] Protected route wrapper
- [ ] Placeholder pages: `/`, `/login`, `/register`, `/book`, `/my-appointments`, `/results`, `/account`

### i18n
- [ ] i18next setup, mặc định VI

### Build & Env
- [ ] `.env.development`: `VITE_API_URL=http://localhost:8080`
- [ ] `vite.config.ts`: proxy `/api` → backend (dev mode)

---

## ĐIỂM KẾT NỐI Sprint 1

| Hạng mục | Backend | Desktop | Web |
|----------|---------|---------|-----|
| API base URL | `localhost:8080` | config trong Wails | `.env` VITE_API_URL |
| Health check | `GET /health` sẵn sàng | Test gọi từ Wails | Test gọi từ Vite proxy |
| Crypto package | Phải xong trước Sprint 2 | — | — |
| Design system | — | Ant Design theme | shadcn/ui + Tailwind |

## DEFINITION OF DONE

- [ ] `docker-compose up` → tất cả service healthy
- [ ] `GET /health` trả `200 OK`
- [ ] `GET /ready` trả `200 OK` (PG + Mongo + Redis connected)
- [ ] `pkg/crypto` unit test pass 100%
- [ ] Migration 001–003 chạy thành công
- [ ] Desktop app build và hiển thị layout cơ bản
- [ ] Web app chạy Vite dev server, hiển thị landing placeholder
- [ ] CI GitHub Actions pass
