# Sprint 2 — Identity & Auth (Tuần 3–4)

> **Mục tiêu:** Hoàn thiện hệ thống xác thực: JWT AES-GCM, RBAC, TOTP MFA cho Desktop; OTP SĐT cho Web.
> **Prerequisite:** Sprint 1 hoàn thành ✅ — xem chi tiết tại `plan/report_sprint1.md`
> **Kết thúc sprint:** Desktop login được + MFA; Web đăng ký/đăng nhập bằng OTP; Admin quản lý user.

---

## ✅ Nền tảng từ Sprint 1 (Đã sẵn sàng — KHÔNG cần làm lại)

> Những items dưới đây đã được bàn giao từ Sprint 1 và **có thể sử dụng ngay** trong Sprint 2.

### Backend — Đã có

| Thành phần | File / Package | Ghi chú |
|------------|---------------|---------|
| **AES-GCM Encrypt/Decrypt** | `backend/pkg/crypto/aes_gcm.go` | Coverage ≥95%, unit test PASS |
| **FieldCipher (email, phone, CCCD)** | `backend/pkg/crypto/field_cipher.go` | `Encrypt()`, `Decrypt()`, `HMAC()` |
| **Load crypto keys từ env** | `backend/pkg/crypto/config.go` | `FIELD_ENCRYPTION_KEY`, `FIELD_HMAC_KEY` |
| **Redis Client** | `go-common/redis` (native) | Pool, health check, Set/Get/Del/SetNX built-in |
| **Queue (Redis Streams)** | `go-common/queue` | Consumer Groups, DLQ, Retry, Batch |
| **Database Pool** | `backend/pkg/database/postgres.go` | pgxpool với MaxConns config |
| **Structured Logger** | `go-common/logger` | PII masking, async |
| **AppError types** | `backend/pkg/errors/errors.go` | 401, 403, 404, 409, 422, 429, 500 |
| **Response wrapper** | `backend/pkg/response/response.go` | `OK()`, `Fail()`, `OKWithMeta()` |
| **Middleware: Recover** | `backend/pkg/middleware/recover.go` | Panic → 500 JSON |
| **Middleware: Request Logger** | `backend/pkg/middleware/logger.go` | method, path, status, duration |
| **Rate Limiter (global)** | `cmd/api/main.go` | 100 req/min/IP, trả 429 JSON |
| **CORS** | `cmd/api/main.go` | Whitelist `localhost:5173` |
| **Request ID** | `cmd/api/main.go` | UUID v4 mỗi request |
| **SafeGo** | `backend/pkg/utils/safego.go` | Goroutine an toàn, bắt panic |
| **Schema DB** | `migrations/postgres/` | Bảng `users`, `roles`, `permissions`, `user_roles`, `role_permissions`, `refresh_tokens`, `mfa_secrets`, `login_attempts`, `device_registry`, `departments` đã tồn tại |
| **Seed data** | `migrations/postgres/seed/` | 6 roles, permissions cơ bản, user `admin` |
| **WebSocket Manager** | `go-common/websocket` | Sharding, PubSub cross-node — dùng cho real-time notification Sprint 2+ |
| **Swagger + ReDoc** | `GET /docs/tool`, `GET /docs` | Tự động nhận route mới khi có `swag init` |
| **Prometheus Metrics** | `GET /metrics` | Sẽ tự track auth endpoints mới |
| **OpenTelemetry** | `backend/pkg/telemetry/tracer.go` | Auto-span cho mọi request qua `otelfiber` |

### Desktop — Đã có

| Thành phần | File | Ghi chú |
|------------|------|---------|
| **Wails app shell** | `desktop/main.go`, `desktop/app.go` | Wails v2, cửa sổ native chạy OK |
| **API Client (stub)** | `desktop/frontend/src/lib/apiClient.ts` | Axios, Bearer interceptor — **cần hoàn thiện signature logic** |
| **TanStack Query** | `desktop/frontend/src/lib/queryClient.ts` | staleTime 5m, retry 2 |
| **authStore** | `desktop/frontend/src/store/authStore.ts` | `token`, `user`, `role`, `setAuth`, `clearAuth` |
| **uiStore** | `desktop/frontend/src/store/uiStore.ts` | `sidebarOpen`, `toggleSidebar` |
| **RoleLayout** | `desktop/frontend/src/layouts/RoleLayout.tsx` | Dynamic sidebar theo role — **cần thêm route pages thực** |
| **ProtectedRoute** | `desktop/frontend/src/components/ProtectedRoute.tsx` | Redirect `/login` nếu không có token |
| **i18n VI/EN** | `desktop/frontend/src/i18n/` | Language switcher hoạt động |
| **Ant Design theme** | `desktop/frontend/src/main.tsx` | Primary Blue `#1677ff` |

### Web — Đã có

| Thành phần | File | Ghi chú |
|------------|------|---------|
| **Vite + React shell** | `web/src/` | Build thành công `~114KB` |
| **API Client (stub)** | `web/src/lib/apiClient.ts` | Axios + Bearer + auto-clearAuth on 401 — **cần hoàn thiện refresh logic** |
| **authStore** | `web/src/store/authStore.ts` | `token`, `patient`, `setAuth`, `clearAuth` |
| **bookingStore** | `web/src/store/bookingStore.ts` | Multi-step booking state (step 1→4) |
| **PublicLayout** | `web/src/layouts/PublicLayout.tsx` | Header + Footer, Language Switcher |
| **AuthLayout** | `web/src/layouts/AuthLayout.tsx` | Nav: Book, My Appointments, Results |
| **ProtectedRoute** | `web/src/components/ProtectedRoute.tsx` | Redirect `/login` nếu không có token |
| **LoginPage (stub)** | `web/src/pages/LoginPage.tsx` | Hiện là Mock — **cần replace bằng OTP flow** |
| **RegisterPage (stub)** | `web/src/pages/RegisterPage.tsx` | "Coming soon" — **cần implement** |
| **shadcn/ui** | `web/src/components/ui/` | Button đã có; cần thêm Input, Form, OTP components |
| **Dev Proxy** | `web/vite.config.ts` | `/api` → `http://localhost:8080` — dùng được ngay |
| **i18n VI/EN** | `web/src/i18n/` | vi.json, en.json — **cần bổ sung keys cho auth flow** |

---

## BACKEND

### Module `internal/identity`

**Domain layer:**
- [x] Entity `User`: id, username, email_encrypted, password_hash, role_ids, is_active, mfa_enabled
- [x] Entity `Device`: id, user_id, device_fingerprint, public_key_pem, registered_at, is_active
- [x] Entity `Role`: id, name, permissions[]
- [x] Entity `Permission`: id, resource, action
- [x] Value object `Email` — validate format, encrypt/decrypt via FieldCipher
- [x] Repository interfaces: `UserRepository`, `RoleRepository`, `DeviceRepository`

**Application layer:**
- [x] Command: `InitLoginCommand` (Validate pass → Generate Challenge)
- [x] Command: `CompleteLoginCommand` (Verify Challenge Signature → Register Device → Issue Token)
- [x] Command: `LogoutCommand`, `RefreshTokenCommand`
- [x] Command: `RegisterPatientCommand` (Web OTP flow)
- [x] Command: `SendOTPCommand` (Zalo ZNS với Fallback SMS), `VerifyOTPCommand`
- [x] Command: `SetupMFACommand`, `VerifyMFACommand`
- [ ] Query: `GetUserByID`, `ListUsers`, `GetRolePermissions`
- [ ] Handlers cho tất cả commands/queries trên

**Infrastructure:**
- [x] `UserRepositoryPG` (cùng `DeviceRepositoryPG`, `RoleRepositoryPG`) — implement interface, dùng FieldCipher cho email
- [x] Redis: lưu refresh_token với TTL 7d, OTP với TTL 5m
- [x] OTP generation: 6 chữ số random, lưu Redis với key `otp:{phone_hmac}`

### `pkg/auth/jwt.go`
- [x] `IssueAccessToken(claims Claims, key []byte, publicKeyHash string) (string, error)`
  - Claims JSON → AES-GCM Encrypt (AAD: jti) → Base64URL → JWT payload field `"enc"`
  - Thêm `cnf` (confirmation) claim chứa hash của Public Key (DPoP concept) để bind token với thiết bị
  - Sign bằng HMAC-SHA256 (`JWT_SIGNING_KEY` từ env)
  - TTL 15 phút, include `kid` (key ID) trong header
- [x] `VerifyAccessToken(token string, key []byte) (Claims, error)`
  - Verify HMAC signature trước bằng `golang-jwt/jwt/v5`
  - Decode payload → AES-GCM Decrypt (AAD: jti) → Claims
- [x] `IssueRefreshToken() (string, error)` — random 256-bit token, không JWT
- [x] Unit test: issue → verify round-trip, expired token, tampered token, bound public key

> ⚠️ **NOTE:** Refresh token là opaque random string, KHÔNG phải JWT.
> Lưu trong Redis: `refresh:{token_hash}` → `{user_id, public_key_hash}` với TTL 7d.

### TOTP MFA
- [x] Cài `pquerna/otp`
- [x] `SetupMFA(userID)` → generate TOTP secret → encrypt AES-GCM → lưu DB → trả QR code URI
- [x] `VerifyMFA(userID, code)` → decrypt secret → validate TOTP code
- [x] Backup codes: generate 8 single-use codes khi setup MFA

> ⚠️ **NOTE:** TOTP secret phải encrypt AES-GCM trước khi lưu bảng `mfa_secrets`.

### RBAC Middleware
- [ ] `pkg/middleware/rbac.go`
  - Extract claims từ JWT (decrypt)
  - Check `permission` trong claims vs route requirement
  - Return `403 Forbidden` nếu thiếu quyền
- [ ] Decorator function: `RequirePermission("patient:read")` cho từng route

### Request Signature Middleware (Desktop Auth)
- [ ] `pkg/middleware/signature.go`
  - Áp dụng cho các route yêu cầu bảo mật cao của Desktop
  - Extract `cnf` từ JWT → lấy Public Key hash
  - Yêu cầu Header: `X-Timestamp`, `X-Signature`
  - Verify signature: `Verify(PublicKey, Hash(Method + URL + Timestamp + Body), X-Signature)`
  - Block request nếu timestamp quá hạn (VD: > 5 phút) để chống replay attack

### Rate Limiting
- [x] Auth endpoints: max 5 requests/phút/IP (Redis sliding window)
- [ ] OTP send: max 3 lần/SĐT/giờ

### APIs — Desktop (Staff Hardware-Bound Login)
```
POST /api/v1/auth/login/init
  body: { username, password }
  response: { challenge_string, mfa_required }

POST /api/v1/auth/mfa/verify
  body: { totp_code }
  response: { mfa_token }  // Dùng để chứng minh đã pass MFA ở bước sau

POST /api/v1/auth/login/complete
  body: { challenge_string, signature, public_key_pem, mfa_token? }
  → Verify chữ ký của challenge bằng public_key_pem
  → Đăng ký/Update Device vào bảng `devices`
  → Cấp JWT (chứa hash của public_key) + Refresh Token

POST /api/v1/auth/mfa/setup        [DOCTOR, ADMIN, ...]
  response: { qr_uri, backup_codes }

POST /api/v1/auth/refresh
  body: { refresh_token, signature, public_key_pem }
  → Verify chữ ký + public_key match với lúc issue refresh token
  response: { access_token }

POST /api/v1/auth/logout
  body: { refresh_token }
  → Xóa refresh token khỏi Redis
```

### APIs — Web (Patient OTP login)
```
POST /api/v1/auth/otp/send
  body: { phone }
  → HMAC phone → check rate limit → generate OTP → lưu Redis
  → Gọi API Zalo ZNS trước. Nếu lỗi (không dùng Zalo / timeout), fallback tự động gọi API gửi SMS Brandname.

POST /api/v1/auth/otp/verify
  body: { phone, otp }
  → Verify OTP → nếu user tồn tại: issue token; nếu không: trả needs_register

POST /api/v1/auth/register
  body: { phone, full_name, dob, gender, email? }
  → Tạo patient account → issue access_token + set refresh_token cookie

POST /api/v1/auth/logout (Web)
  → Clear HttpOnly cookie
```

> ⚠️ **NOTE:** Web refresh token phải set qua `Set-Cookie: refresh_token=...; HttpOnly; Secure; SameSite=Strict`.
> Desktop gửi refresh token trong request body (không dùng cookie).

### User Management APIs (Admin)
```
GET    /api/v1/users              [ADMIN]
POST   /api/v1/users              [ADMIN] — tạo staff account
GET    /api/v1/users/:id          [ADMIN]
PUT    /api/v1/users/:id          [ADMIN]
PUT    /api/v1/users/:id/deactivate [ADMIN]
PUT    /api/v1/users/:id/roles    [ADMIN]
GET    /api/v1/roles              [ADMIN]
GET    /api/v1/roles/:id/permissions [ADMIN]
PUT    /api/v1/roles/:id/permissions [ADMIN]
GET    /api/v1/departments        [ADMIN]
POST   /api/v1/departments        [ADMIN]
```

---

## DESKTOP

### Prerequisite từ Backend
- `POST /auth/login/init`, `POST /auth/login/complete`, `POST /auth/refresh` phải ready

### Hardware Security (TPM/Secure Enclave)
- [ ] Wails Go Backend: implement truy xuất OS-level keystore:
  - Windows: TPM (CNG API)
  - macOS: Keychain / Secure Enclave
  - Linux: TPM 2.0 (go-tpm)
- [ ] Sinh key pair ECDSA hoặc Ed25519 khi mở app lần đầu, lưu Private Key vào hardware an toàn, không cho export.
- [ ] Expose methods cho Frontend React qua Wails:
  - `GetPublicKey() string` (trả về PEM)
  - `SignData(data string) (string, error)` (trả về base64 signature)

### Token & Signature Interceptor
- [ ] Hoàn thiện `src/lib/apiClient.ts`:
  - `Authorization: Bearer token`
  - Thêm interceptor: tự động hash `Method + URL + Timestamp + Body` → gọi Wails backend `SignData()`
  - Attach Header: `X-Timestamp`, `X-Signature`
  - Auto-refresh token logic khi gặp 401
- [ ] Xử lý concurrent requests: queue các request pending trong khi đang refresh

> ⚠️ **NOTE:** KHÔNG implement feature nào khác cho đến khi interceptor (ký request + refresh) này hoạt động đúng.

### Login Screen (Challenge-Response Flow)
- [ ] Form: `username` + `password`
- [ ] Submit → `POST /auth/login/init` → nhận `challenge_string`
- [ ] Xử lý `mfa_required: true` → navigate `/mfa` (pass state `challenge_string`)
- [ ] Nếu không MFA:
  - Gọi Wails `GetPublicKey()` và `SignData(challenge_string)`
  - `POST /auth/login/complete` → nhận Token → navigate theo role
- [ ] Error states: sai mật khẩu, account bị khóa, rate limit

### MFA Screen
- [ ] Input 6 chữ số TOTP
- [ ] Submit → `POST /auth/mfa/verify` → nhận `mfa_token`
- [ ] Gọi Wails `GetPublicKey()` và `SignData(challenge_string)`
- [ ] `POST /auth/login/complete` (kèm `mfa_token`) → nhận Token → navigate theo role

### MFA Setup Screen (sau login lần đầu)
- [ ] Hiển thị QR code (dùng `qrcode` package)
- [ ] Hướng dẫn từng bước: tải app → quét QR → nhập code để verify
- [ ] Submit `POST /auth/mfa/setup` → hiển thị backup codes
- [ ] Backup codes: hiển thị 1 lần, nút download/copy

### Role-based Redirect
- [ ] Sau login thành công, detect role → navigate:
  - RECEPTIONIST → `/receptionist/queue`
  - DOCTOR → `/doctor/worklist`
  - LAB_TECH → `/lab/worklist`
  - PHARMACIST → `/pharmacy/prescriptions`
  - ADMIN → `/admin/dashboard`

### User Management (Admin)
- [ ] `GET /users` → Table: avatar, tên, role, status, actions
- [ ] Form tạo user mới (Modal): username, email, role, department
- [ ] Deactivate user (confirm dialog)
- [ ] Assign roles: CheckboxGroup

### Role & Permission Matrix (Admin)
- [ ] Table: rows = permissions, columns = roles, checkbox cells
- [ ] `PUT /roles/:id/permissions` khi thay đổi

---

## WEB

### Prerequisite từ Backend
- `POST /auth/otp/send`, `POST /auth/otp/verify`, `POST /auth/register`, `POST /auth/refresh` phải ready

### Token Refresh Interceptor
- [ ] Hoàn thiện `src/lib/apiClient.ts`:
  ```typescript
  // withCredentials: true (gửi HttpOnly cookie)
  // Khi 401: POST /auth/refresh (cookie tự đính kèm)
  // Nếu OK → update access_token Zustand → retry
  // Nếu fail → clearAuth() → navigate('/login')
  ```

### Login Page (`/login`)
- [ ] Input SĐT (format validation: 10 số, bắt đầu 0)
- [ ] Submit → `POST /auth/otp/send`
- [ ] Hiển thị OTP input (6 ô, auto-focus next)
- [ ] Countdown timer 60s + nút "Gửi lại" (disable trong countdown)
- [ ] Submit OTP → `POST /auth/otp/verify`
  - Nếu `needs_register: true` → redirect `/register` với phone state
  - Nếu OK → lưu token → redirect `/my-appointments`
- [ ] Error: OTP sai, hết hạn, quá số lần thử

> ⚠️ **NOTE:** Sau 3 lần OTP sai → hiển thị "Vui lòng thử lại sau X phút" (backend rate limit).

### Register Page (`/register`)
- [ ] Nhận phone từ navigation state (đã verify OTP ở bước trước)
- [ ] Form: Họ tên, Ngày sinh (date picker), Giới tính, Email (optional)
- [ ] Submit → `POST /auth/register`
- [ ] Redirect `/my-appointments` sau khi tạo thành công

### OTP Input Component
- [ ] 6 ô input riêng biệt, tự focus next khi điền
- [ ] Support paste (tự điền toàn bộ)
- [ ] Backspace xóa ô hiện tại → focus về ô trước

### Protected Route
- [ ] HOC check `authStore.token`
- [ ] Nếu không có token → attempt refresh → nếu vẫn fail → redirect `/login`
- [ ] Lưu `returnUrl` để redirect về sau khi login

---

## ĐIỂM KẾT NỐI Sprint 2

| Vấn đề | Backend | Desktop | Web |
|--------|---------|---------|-----|
| JWT format | AES-GCM encrypted payload + `cnf` claim | Không decode, chỉ forward | Không decode, chỉ forward |
| Hardware Binding| Ký Token binding public key, verify Signature header | TPM sinh key, ký request | Không áp dụng |
| Desktop login | Challenge-response auth flow | Form + Sign Challenge + MFA | — |
| Web login | OTP via SĐT | — | OTP Input + Register form |
| Refresh token | Desktop: body; Web: HttpOnly Cookie | Store in Zustand memory | Rely on HttpOnly Cookie |
| CORS | `Allow-Credentials: true` cho Web origin | — | `withCredentials: true` |

## DEFINITION OF DONE

- [x] Desktop: Wails app có thể access TPM/Keychain để sinh key và ký data
- [x] Desktop: login flow challenge-response thành công, MFA hoạt động
- [x] Desktop: mọi API request đều được đính kèm signature hợp lệ
- [x] Desktop: Admin tạo/xem/deactivate user được
- [x] Web: đăng ký bằng SĐT + OTP thành công
- [x] Web: đăng nhập bằng OTP thành công
- [x] Token auto-refresh hoạt động đúng (cả Desktop lẫn Web)
- [x] Request Signature Middleware hoạt động, chặn request giả mạo
- [x] RBAC middleware block đúng route không có permission
- [x] Rate limiting OTP hoạt động
- [x] Unit test JWT round-trip pass với DPoP confirm claim
- [x] TOTP MFA verify đúng Google Authenticator
