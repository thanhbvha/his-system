# Sprint 2 — Step 3: Web Auth API (Patient OTP Flow)

> **Mục tiêu:** Xây dựng authentication flow cho bệnh nhân trên Web: đăng ký/đăng nhập bằng OTP qua SĐT (Zalo ZNS + SMS fallback), phát token với HttpOnly cookie.
> **Phụ thuộc:** Step 1 (Identity Domain + JWT hoàn thành).
> **Output:** Bệnh nhân đăng ký và đăng nhập thành công qua OTP; refresh token lưu HttpOnly cookie.

---

## Nền tảng Sprint 1 + Step 1 sử dụng

| Package / Schema | Dùng để |
|-----------------|---------|
| `pkg/auth/jwt.go` | `IssueAccessToken` không có `cnf` claim (Web token) |
| `internal/identity/domain/` | User entity (patient role) |
| `internal/identity/infrastructure/UserRepositoryPG` | Create + GetByPhoneHMAC |
| `go-common/redis` | Lưu OTP `otp:{phone_hmac}` TTL 5m, rate limit counter |
| `go-common/queue` | Gửi OTP async qua Queue worker |
| `pkg/crypto/field_cipher.go` | HMAC phone để lookup, encrypt phone/CCCD lưu DB |
| Schema `patients` | `phone_encrypted`, `phone_hmac` — đã có từ Sprint 1 |

---

## 1. Application Layer — Commands

### `SendOTPCommand`

- [ ] `internal/identity/application/command/send_otp.go`:
  ```go
  type SendOTPCommand struct {
      Phone    string   // raw phone, VD: "0912345678"
      ClientIP string
  }
  ```
  - Validate format: 10 chữ số, bắt đầu bằng `0`
  - HMAC phone → key `otp:{phone_hmac}`
  - **Rate limit:** check Redis key `rl:otp:{phone_hmac}` — max 3 lần/giờ
    ```
    Key: "rl:otp:{phone_hmac}"
    Value: INCR → nếu > 3 → trả ErrTooManyRequests
    TTL: 3600s (1 giờ)
    ```
  - Generate OTP: 6 chữ số random (`math/rand/v2`, seed từ `crypto/rand`)
  - Lưu Redis: `otp:{phone_hmac}` → `{otp, attempts: 0}` TTL 300s (5 phút)
  - Enqueue job gửi OTP vào Queue:
    ```go
    q.Enqueue(ctx, queue.Job{
        Type:    "send_otp",
        Payload: map[string]any{"phone": phone, "otp": otp},
    })
    ```

### OTP Delivery Worker — `cmd/worker/`

- [ ] `cmd/worker/handlers/send_otp_handler.go`:
  - Nhận job `send_otp` từ Queue
  - **Bước 1:** Gọi Zalo ZNS API
    ```go
    err := zaloClient.SendOTP(phone, otp)
    ```
  - **Bước 2 (fallback):** Nếu Zalo lỗi (timeout, error) → tự động gọi SMS Brandname
    ```go
    if err != nil {
        _ = smsClient.SendBrandname(phone, fmt.Sprintf("Ma OTP HIS cua ban la: %s", otp))
    }
    ```
  - [ ] `pkg/notify/zalo_zns.go` — interface + HTTP client gọi Zalo ZNS API
  - [ ] `pkg/notify/sms_brandname.go` — interface + HTTP client gọi SMS provider
  - [ ] Config credentials từ env:
    ```env
    ZALO_OA_ACCESS_TOKEN=<token>
    ZALO_TEMPLATE_ID=<id>
    SMS_PROVIDER_URL=<url>
    SMS_API_KEY=<key>
    SMS_BRANDNAME=HIS
    ```

### `VerifyOTPCommand`

- [ ] `internal/identity/application/command/verify_otp.go`:
  ```go
  type VerifyOTPCommand struct {
      Phone string
      OTP   string
  }

  type VerifyOTPResult struct {
      NeedsRegister bool     // true nếu chưa có tài khoản
      AccessToken   string   // rỗng nếu NeedsRegister = true
      RefreshToken  string   // rỗng nếu NeedsRegister = true
  }
  ```
  - HMAC phone → lookup Redis `otp:{phone_hmac}`
  - Tăng `attempts` trong Redis value → nếu ≥ 5 lần sai → xoá OTP, trả `ErrTooManyRequests`
  - So sánh OTP (constant time compare)
  - Nếu đúng: xoá key Redis
  - Tìm patient bằng `phone_hmac` trong DB
    - Nếu **không tồn tại** → trả `NeedsRegister: true`
    - Nếu **tồn tại** → Issue AccessToken + RefreshToken Web

### `RegisterPatientCommand`

- [ ] `internal/identity/application/command/register_patient.go`:
  ```go
  type RegisterPatientCommand struct {
      Phone     string
      FullName  string
      DOB       time.Time
      Gender    string   // "male" | "female" | "other"
      Email     string   // optional
  }
  ```
  - Validate: FullName không rỗng, DOB hợp lệ, Gender hợp lệ
  - Encrypt phone (AES-GCM) + HMAC phone
  - Tạo User record (`role = patient`) + Patient record (phone_encrypted, phone_hmac)
  - Nếu có email: encrypt + HMAC email
  - Issue AccessToken (Web, không `cnf`) + RefreshToken
  - Refresh token lưu Redis: `refresh:{hash}` → `{user_id}` TTL 7d

---

## 2. API Endpoints — `internal/api/auth/web_handler.go`

### Route Setup

- [ ] Đăng ký routes trong `cmd/api/main.go`:
  ```go
  // Web OTP — Public endpoints
  auth.Post("/otp/send",   webAuthHandler.SendOTP)
  auth.Post("/otp/verify", webAuthHandler.VerifyOTP)
  auth.Post("/register",   webAuthHandler.Register)
  auth.Post("/refresh",    webAuthHandler.RefreshWeb)   // đọc cookie
  auth.Post("/logout",     webAuthHandler.LogoutWeb)    // clear cookie
  ```

### Handler Implementations

- [ ] `POST /api/v1/auth/otp/send`:
  ```
  Request:  { "phone": "0912345678" }
  Response: { "success": true, "message": "OTP đã được gửi" }
  Error 422: format SĐT sai
  Error 429: quá 3 lần/giờ với SĐT này
  ```

- [ ] `POST /api/v1/auth/otp/verify`:
  ```
  Request:  { "phone": "0912345678", "otp": "123456" }

  Case 1 - user chưa có TK:
  Response: { "needs_register": true }

  Case 2 - user đã có TK:
  Response: { "needs_register": false, "access_token": "eyJ..." }
  Set-Cookie: refresh_token=...; HttpOnly; Secure; SameSite=Strict; Path=/api/v1/auth; Max-Age=604800

  Error 401: OTP sai
  Error 410: OTP hết hạn
  Error 429: quá 5 lần sai OTP
  ```

- [ ] `POST /api/v1/auth/register`:
  ```
  Request: {
    "phone": "0912345678",   // đã verify OTP ở bước trước
    "full_name": "Nguyễn Văn A",
    "dob": "1990-05-15",
    "gender": "male",
    "email": "a@example.com"
  }
  Response: { "access_token": "eyJ..." }
  Set-Cookie: refresh_token=...; HttpOnly; Secure; SameSite=Strict
  Error 409: SĐT đã đăng ký
  ```

- [ ] `POST /api/v1/auth/refresh` (Web):
  ```
  Cookie: refresh_token=<opaque_token>   // browser tự đính kèm
  Response: { "access_token": "eyJ..." }
  → Cấp refresh token mới, set cookie lại (rotation)
  Error 401: cookie không hợp lệ / hết hạn
  ```

- [ ] `POST /api/v1/auth/logout` (Web):
  ```
  Response: 200 OK
  Set-Cookie: refresh_token=; HttpOnly; Max-Age=0   // clear cookie
  ```

> ⚠️ **NOTE:** Web refresh token lưu trong **HttpOnly Cookie** (`Path=/api/v1/auth`).
> Desktop refresh token gửi trong **request body**.
> CORS config: `AllowCredentials: true` cho Web origin.

---

## 3. CORS Update cho Web Auth

- [ ] Cập nhật CORS trong `cmd/api/main.go` để hỗ trợ cookie:
  ```go
  app.Use(cors.New(cors.Config{
      AllowOrigins:     "http://localhost:5173,https://his-system.vn",
      AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
      AllowMethods:     "GET, POST, PUT, PATCH, DELETE",
      AllowCredentials: true,   // ← BẮT BUỘC cho HttpOnly cookie
  }))
  ```

---

## Definition of Done (Step 3)

- [ ] `POST /api/v1/auth/otp/send` → OTP được gửi qua Queue worker (Zalo/SMS fallback)
- [ ] `POST /api/v1/auth/otp/verify` → trả đúng `needs_register` flag
- [ ] `POST /api/v1/auth/register` → tạo patient + set HttpOnly cookie
- [ ] `POST /api/v1/auth/refresh` (Web) → đọc cookie, issue token mới, rotate cookie
- [ ] `POST /api/v1/auth/logout` (Web) → clear cookie
- [ ] Rate limit OTP: SĐT gửi quá 3 lần/giờ → 429
- [ ] OTP sai 5 lần → key bị xoá, phải gửi OTP mới
- [ ] Phone encrypt+HMAC lưu đúng trong bảng `patients`
- [ ] CORS: request từ `localhost:5173` với `withCredentials: true` không bị block
