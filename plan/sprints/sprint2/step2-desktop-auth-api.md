# Sprint 2 — Step 2: Desktop Auth API (Challenge-Response + TOTP MFA)

> **Mục tiêu:** Xây dựng toàn bộ backend authentication flow cho nhân viên nội bộ: challenge-response hardware-bound login, TOTP MFA, và token management.
> **Phụ thuộc:** Step 1 (Identity Domain + JWT package hoàn thành).
> **Output:** Các API `/auth/login/init`, `/auth/login/complete`, `/auth/mfa/*`, `/auth/refresh`, `/auth/logout` hoạt động end-to-end.

---

## Nền tảng Sprint 1 + Step 1 sử dụng

| Package / Schema | Dùng để |
|-----------------|---------|
| `pkg/auth/jwt.go` | Issue + Verify Access Token có hardware binding |
| `internal/identity/domain/` | User, Device, Role entity + Repository interfaces |
| `internal/identity/infrastructure/` | UserRepositoryPG, DeviceRepositoryPG |
| `go-common/redis` | Lưu refresh token (TTL 7d), challenge (TTL 5m) |
| `go-common/queue` | Gửi email notification async (login từ thiết bị mới) |
| `pkg/errors/errors.go` | ErrUnauthorized, ErrForbidden, ErrConflict |
| `pkg/response/response.go` | OK(), Fail() |
| `pkg/middleware/logger.go` | Auto log auth events |
| Schema `mfa_secrets` | Lưu TOTP secret (đã encrypted) |
| Schema `device_registry` | Lưu Public Key theo device |
| Schema `login_attempts` | Brute-force protection counter |

---

## 1. Application Layer — Commands & Handlers

### `InitLoginCommand`

- [ ] `internal/identity/application/command/init_login.go`:
  ```go
  type InitLoginCommand struct {
      Username    string
      Password    string
      ClientIP    string
  }

  type InitLoginResult struct {
      ChallengeString string  // 256-bit random, lưu Redis TTL 5m
      MFARequired     bool
  }
  ```
  - Validate username/password → bcrypt compare
  - Check `login_attempts` — nếu ≥ 5 lần sai trong 15 phút → `ErrTooManyRequests (429)`
  - Generate `challenge_string` (32 random bytes → base64url)
  - Lưu Redis: `challenge:{challenge_hash}` → `{user_id}` TTL 5 phút
  - Trả `MFARequired: user.MFAEnabled`

### `CompleteLoginCommand`

- [ ] `internal/identity/application/command/complete_login.go`:
  ```go
  type CompleteLoginCommand struct {
      ChallengeString string
      Signature       string   // base64(ECDSA-P256 signature, DER-encoded)
      PublicKeyPEM    string   // ECDSA-P256 public key PEM (secp256r1)
      MFAToken        string   // optional, nếu MFA đã pass
      DeviceFingerprint string
      ClientIP        string
  }

  type CompleteLoginResult struct {
      AccessToken  string
      RefreshToken string
  }
  ```
  - Verify challenge còn hạn trong Redis
  - Verify ECDSA-P256 signature của challenge bằng PublicKeyPEM
    (`crypto/ecdsa` + `elliptic.P256()` — cố định, không detect algorithm)
  - Nếu user có `MFAEnabled=true`: verify `MFAToken` còn hạn (TTL 5m từ Redis)
  - Upsert Device vào `device_registry`
  - Issue AccessToken (với `cnf.jkt = SHA256(PublicKeyPEM)`)
  - Issue RefreshToken → lưu Redis: `refresh:{hash(token)}` → `{user_id, public_key_hash}` TTL 7d
  - Xoá challenge khỏi Redis
  - Reset `login_attempts`

### `LogoutCommand`

- [ ] `internal/identity/application/command/logout.go`:
  - Nhận `RefreshToken` → hash → xoá khỏi Redis
  - Trả 200 dù token không tồn tại (idempotent)

### `RefreshTokenCommand`

- [ ] `internal/identity/application/command/refresh_token.go`:
  ```go
  type RefreshTokenCommand struct {
      RefreshToken  string
      Signature     string
      PublicKeyPEM  string
  }
  ```
  - Hash refresh token → lookup Redis
  - Verify `public_key_hash` trong Redis match với `SHA256(PublicKeyPEM)`
  - Verify signature của `RefreshToken` bằng PublicKeyPEM (chống token theft)
  - Issue AccessToken mới, xoá refresh token cũ, cấp refresh token mới (rotation)

### `SetupMFACommand`

- [ ] `internal/identity/application/command/setup_mfa.go`:
  - `go get github.com/pquerna/otp`
  - Generate TOTP secret (base32, 32 bytes)
  - **Encrypt secret bằng AES-GCM trước khi lưu** bảng `mfa_secrets`
  - Generate 8 backup codes (random 12 hex chars), hash bcrypt, lưu DB
  - Trả `qr_uri` (format: `otpauth://totp/HIS:{username}?secret={raw_secret}&issuer=HIS`)
  - Trả `backup_codes` (plaintext, chỉ trả 1 lần)

### `VerifyMFACommand`

- [ ] `internal/identity/application/command/verify_mfa.go`:
  - Decrypt TOTP secret từ DB
  - Validate TOTP code với window ±1 step (30s)
  - Nếu đúng: tạo `mfa_token` → lưu Redis `mfa:{token}` → `{user_id}` TTL 5m
  - Hỗ trợ backup code: check hash → invalidate sau khi dùng

---

## 2. Rate Limiting (Auth-specific)

- [ ] `pkg/middleware/auth_rate_limit.go`:
  - **Login endpoint:** max 5 req/IP/phút (Redis sliding window)
    ```go
    func AuthRateLimit(rdb *redis.Client) fiber.Handler
    // Key: "rl:login:{ip}" → INCR + EXPIRE
    ```
  - **Nếu sai password:** tăng counter `login_attempts:{username}` TTL 15m
  - Khác với global rate limiter (100 req/min) từ Sprint 1 — đây là per-endpoint

---

## 3. API Endpoints — `internal/api/auth/desktop_handler.go`

### Route Setup

- [ ] Đăng ký routes trong `cmd/api/main.go`:
  ```go
  auth := api.Group("/api/v1/auth")
  auth.Use(authRateLimit)   // 5 req/phút/IP

  // Desktop — Hardware-bound login
  auth.Post("/login/init",     desktopAuthHandler.InitLogin)
  auth.Post("/login/complete", desktopAuthHandler.CompleteLogin)
  auth.Post("/mfa/verify",     desktopAuthHandler.VerifyMFA)
  auth.Post("/mfa/setup",      jwtMiddleware, desktopAuthHandler.SetupMFA)
  auth.Post("/refresh",        desktopAuthHandler.RefreshToken)
  auth.Post("/logout",         desktopAuthHandler.Logout)
  ```

### Handler Implementations

- [ ] `POST /api/v1/auth/login/init`:
  ```
  Request:  { "username": "drnguyenvan", "password": "Secret@123" }
  Response: { "challenge_string": "...", "mfa_required": true }
  Error 401: sai username/password
  Error 429: quá nhiều lần thử
  Error 423: account bị khóa
  ```

- [ ] `POST /api/v1/auth/mfa/verify`:
  ```
  Request:  { "username": "drnguyenvan", "totp_code": "123456" }
  Response: { "mfa_token": "..." }  // TTL 5 phút
  Error 401: sai TOTP code
  ```

- [ ] `POST /api/v1/auth/login/complete`:
  ```
  Request: {
    "challenge_string": "...",
    "signature": "<base64>",
    "public_key_pem": "-----BEGIN PUBLIC KEY-----...",
    "mfa_token": "...",                   // optional
    "device_fingerprint": "win-machine-1"
  }
  Response: { "access_token": "...", "refresh_token": "..." }
  Error 401: signature không hợp lệ
  Error 401: challenge hết hạn
  Error 403: MFA required nhưng không có mfa_token
  ```

- [ ] `POST /api/v1/auth/mfa/setup`:
  ```
  Auth: Bearer JWT (đã login)
  Response: { "qr_uri": "otpauth://totp/...", "backup_codes": ["abc123", ...] }
  ```

- [ ] `POST /api/v1/auth/refresh`:
  ```
  Request: { "refresh_token": "...", "signature": "...", "public_key_pem": "..." }
  Response: { "access_token": "..." }
  Error 401: refresh token không hợp lệ / hết hạn
  Error 401: public key không khớp
  ```

- [ ] `POST /api/v1/auth/logout`:
  ```
  Request: { "refresh_token": "..." }
  Response: 200 OK (luôn luôn)
  ```

---

## 4. Swagger Annotations

- [ ] Thêm `// @Summary`, `// @Tags`, `// @Accept`, `// @Produce`, `// @Param`, `// @Success`, `// @Failure` cho tất cả handlers trên
- [ ] Chạy `swag init -g ./cmd/api/main.go --output ./docs` để generate lại docs

---

## Definition of Done (Step 2)

- [ ] `POST /api/v1/auth/login/init` nhận username/password, trả challenge
- [ ] `POST /api/v1/auth/login/complete` verify signature, issue JWT + Refresh Token
- [ ] `POST /api/v1/auth/mfa/verify` validate TOTP code, trả mfa_token
- [ ] `POST /api/v1/auth/mfa/setup` tạo TOTP secret encrypt + trả QR URI
- [ ] `POST /api/v1/auth/refresh` rotate refresh token, issue access token mới
- [ ] `POST /api/v1/auth/logout` xoá refresh token khỏi Redis
- [ ] Rate limit: request thứ 6 liên tiếp cùng IP → 429
- [ ] Brute force: sai password 5 lần → account lock
- [ ] TOTP secret được lưu AES-GCM encrypted trong DB (kiểm tra bằng DB viewer)
- [ ] Swagger docs hiển thị đúng tất cả endpoints mới
