# Sprint 2 — Identity & Auth (Tuần 3–4)

> **Mục tiêu:** Hoàn thiện hệ thống xác thực: JWT AES-GCM, RBAC, TOTP MFA cho Desktop; OTP SĐT cho Web.
> **Prerequisite:** Sprint 1 hoàn thành, `pkg/crypto` pass test.
> **Kết thúc sprint:** Desktop login được + MFA; Web đăng ký/đăng nhập bằng OTP; Admin quản lý user.

---

## BACKEND

### Module `internal/identity`

**Domain layer:**
- [ ] Entity `User`: id, username, email_encrypted, password_hash, role_ids, is_active, mfa_enabled
- [ ] Entity `Device`: id, user_id, device_fingerprint, public_key_pem, registered_at, is_active
- [ ] Entity `Role`: id, name, permissions[]
- [ ] Entity `Permission`: id, resource, action
- [ ] Value object `Email` — validate format, encrypt/decrypt via FieldCipher
- [ ] Repository interfaces: `UserRepository`, `RoleRepository`, `DeviceRepository`

**Application layer:**
- [ ] Command: `InitLoginCommand` (Validate pass → Generate Challenge)
- [ ] Command: `CompleteLoginCommand` (Verify Challenge Signature → Register Device → Issue Token)
- [ ] Command: `LogoutCommand`, `RefreshTokenCommand`
- [ ] Command: `RegisterPatientCommand` (Web OTP flow)
- [ ] Command: `SendOTPCommand` (Zalo ZNS với Fallback SMS), `VerifyOTPCommand`
- [ ] Command: `SetupMFACommand`, `VerifyMFACommand`
- [ ] Query: `GetUserByID`, `ListUsers`, `GetRolePermissions`
- [ ] Handlers cho tất cả commands/queries trên

**Infrastructure:**
- [ ] `UserRepositoryPG` — implement interface, dùng FieldCipher cho email
- [ ] Redis: lưu refresh_token với TTL 7d, OTP với TTL 5m
- [ ] OTP generation: 6 chữ số random, lưu Redis với key `otp:{phone_hmac}`

### `pkg/auth/jwt.go`
- [ ] `IssueAccessToken(claims Claims, key []byte, publicKeyHash string) (string, error)`
  - Claims JSON → AES-GCM Encrypt → Base64URL → JWT payload field `"enc"`
  - Thêm `cnf` (confirmation) claim chứa hash của Public Key (DPoP concept) để bind token với thiết bị
  - Sign bằng HMAC-SHA256 (`JWT_SIGNING_KEY` từ env)
  - TTL 15 phút, include `kid` (key ID) trong header
- [ ] `VerifyAccessToken(token string, key []byte) (Claims, error)`
  - Verify HMAC signature trước
  - Decode payload → AES-GCM Decrypt → Claims
- [ ] `IssueRefreshToken() (string, error)` — random 256-bit token, không JWT
- [ ] Unit test: issue → verify round-trip, expired token, tampered token, bound public key

> ⚠️ **NOTE:** Refresh token là opaque random string, KHÔNG phải JWT.
> Lưu trong Redis: `refresh:{token_hash}` → `{user_id, public_key_hash}` với TTL 7d.

### TOTP MFA
- [ ] Cài `pquerna/otp`
- [ ] `SetupMFA(userID)` → generate TOTP secret → encrypt AES-GCM → lưu DB → trả QR code URI
- [ ] `VerifyMFA(userID, code)` → decrypt secret → validate TOTP code
- [ ] Backup codes: generate 8 single-use codes khi setup MFA

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
- [ ] Auth endpoints: max 5 requests/phút/IP (Redis sliding window)
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

- [ ] Desktop: Wails app có thể access TPM/Keychain để sinh key và ký data
- [ ] Desktop: login flow challenge-response thành công, MFA hoạt động
- [ ] Desktop: mọi API request đều được đính kèm signature hợp lệ
- [ ] Desktop: Admin tạo/xem/deactivate user được
- [ ] Web: đăng ký bằng SĐT + OTP thành công
- [ ] Web: đăng nhập bằng OTP thành công
- [ ] Token auto-refresh hoạt động đúng (cả Desktop lẫn Web)
- [ ] Request Signature Middleware hoạt động, chặn request giả mạo
- [ ] RBAC middleware block đúng route không có permission
- [ ] Rate limiting OTP hoạt động
- [ ] Unit test JWT round-trip pass với DPoP confirm claim
- [ ] TOTP MFA verify đúng Google Authenticator
