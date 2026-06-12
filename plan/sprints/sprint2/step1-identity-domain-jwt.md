# Sprint 2 — Step 1: Identity Domain & JWT Package

> **Mục tiêu:** Xây dựng tầng Domain của module `internal/identity` và package `pkg/auth` làm nền tảng cho toàn bộ authentication ở Sprint 2.
> **Phụ thuộc:** `backend/pkg/crypto` (AES-GCM, FieldCipher) — ✅ đã có từ Sprint 1.
> **Output:** Entity, Repository interface, Value Object, và `pkg/auth/jwt.go` pass 100% unit test.

---

## Nền tảng Sprint 1 sử dụng trong step này

| Package | Dùng để |
|---------|---------|
| `pkg/crypto/field_cipher.go` | Mã hoá `email_encrypted` trong User entity |
| `pkg/crypto/aes_gcm.go` | Mã hoá JWT Claims payload + TOTP secret |
| `pkg/errors/errors.go` | `ErrUnauthorized`, `ErrForbidden` trong Repository |
| `backend/pkg/database/postgres.go` | pgxpool inject vào Repository |
| Schema `migrations/postgres/001_identity.sql` | Bảng `users`, `roles`, `permissions`, `mfa_secrets`, `device_registry` đã tồn tại |

---

## 1. Domain Layer — `internal/identity/domain/`

### Entity `User`

- [ ] `internal/identity/domain/user.go`:
  ```go
  type User struct {
      ID             uuid.UUID
      Username       string
      EmailEncrypted string        // AES-GCM encrypted
      EmailHMAC      string        // HMAC-SHA256 deterministic, dùng để lookup
      PasswordHash   string        // bcrypt
      RoleIDs        []uuid.UUID
      IsActive       bool
      MFAEnabled     bool
      CreatedAt      time.Time
      UpdatedAt      time.Time
  }
  ```
- [ ] Method `User.SetEmail(plain string, cipher *FieldCipher) error` — encrypt + HMAC
- [ ] Method `User.GetEmail(cipher *FieldCipher) (string, error)` — decrypt

### Entity `Device`

- [ ] `internal/identity/domain/device.go`:
  ```go
  type Device struct {
      ID                 uuid.UUID
      UserID             uuid.UUID
      DeviceFingerprint  string
      PublicKeyPEM       string   // ECDSA-P256 public key PEM (secp256r1)
      // ⚠️ Luôn dùng ECDSA-P256: đây là thuật toán duy nhất được hỗ trợ
      // native trên cả Windows CNG/TPM, macOS Secure Enclave, và Linux TPM 2.0.
      PublicKeyHash      string   // SHA-256 của PEM, dùng trong JWT cnf claim
      RegisteredAt       time.Time
      IsActive           bool
  }
  ```

### Entity `Role` + `Permission`

- [ ] `internal/identity/domain/role.go`:
  ```go
  type Role struct {
      ID          uuid.UUID
      Name        string       // "admin", "doctor", ...
      Permissions []Permission
  }

  type Permission struct {
      ID       uuid.UUID
      Resource string  // "patient", "appointment", ...
      Action   string  // "read", "write", "delete"
  }

  // Helper: "patient:read"
  func (p Permission) String() string { return p.Resource + ":" + p.Action }
  ```

### Value Object `Email`

- [ ] `internal/identity/domain/email.go`:
  ```go
  type Email struct {
      plain     string
      encrypted string
      hmac      string
  }

  func NewEmail(plain string, cipher *FieldCipher) (Email, error)
  func (e Email) Encrypted() string
  func (e Email) HMAC() string
  func (e Email) Reveal(cipher *FieldCipher) (string, error)
  ```
  - Validate format email trước khi tạo
  - Trả `ErrValidation` nếu format sai

### Repository Interfaces

- [ ] `internal/identity/domain/repository.go`:
  ```go
  type UserRepository interface {
      Create(ctx context.Context, user *User) error
      GetByID(ctx context.Context, id uuid.UUID) (*User, error)
      GetByUsername(ctx context.Context, username string) (*User, error)
      GetByEmailHMAC(ctx context.Context, emailHMAC string) (*User, error)
      Update(ctx context.Context, user *User) error
      List(ctx context.Context, page, limit int) ([]*User, int64, error)
  }

  type RoleRepository interface {
      GetByID(ctx context.Context, id uuid.UUID) (*Role, error)
      GetByName(ctx context.Context, name string) (*Role, error)
      List(ctx context.Context) ([]*Role, error)
      UpdatePermissions(ctx context.Context, roleID uuid.UUID, perms []Permission) error
  }

  type DeviceRepository interface {
      Upsert(ctx context.Context, device *Device) error
      GetByUserAndFingerprint(ctx context.Context, userID uuid.UUID, fingerprint string) (*Device, error)
      DeactivateByUser(ctx context.Context, userID uuid.UUID) error
  }
  ```

---

## 2. Package `pkg/auth/jwt.go`

> ⚠️ **NOTE:** JWT payload được **mã hoá AES-GCM** toàn bộ trước khi sign. Không ai có thể đọc claims nếu không có `JWT_ENCRYPTION_KEY`. Refresh token là **opaque random string**, KHÔNG phải JWT.

- [ ] Thêm env variables vào `.env`:
  ```env
  JWT_SIGNING_KEY=<hmac-sha256-key-base64>      # ký JWT header+payload
  JWT_ENCRYPTION_KEY=<aes-256-key-base64>       # mã hoá claims payload
  JWT_ACCESS_TTL=15m
  JWT_REFRESH_TTL=168h                           # 7 ngày
  ```

- [ ] `pkg/auth/claims.go`:
  ```go
  type Claims struct {
      UserID      uuid.UUID   `json:"sub"`
      Username    string      `json:"usr"`
      Roles       []string    `json:"roles"`
      Permissions []string    `json:"perms"`  // ["patient:read", "appointment:write"]
      IssuedAt    int64       `json:"iat"`
      ExpiresAt   int64       `json:"exp"`
  }
  ```

- [ ] `pkg/auth/jwt.go`:
  ```go
  // IssueAccessToken mã hoá Claims bằng AES-GCM, sau đó tạo JWT.
  // publicKeyHash = SHA-256(PublicKeyPEM), bind token với thiết bị (DPoP pattern).
  // Nếu publicKeyHash rỗng → Web token (không cần hardware binding).
  func IssueAccessToken(claims Claims, signingKey, encKey []byte, publicKeyHash string) (string, error)

  // VerifyAccessToken verify chữ ký HMAC trước, sau đó giải mã payload.
  func VerifyAccessToken(token string, signingKey, encKey []byte) (Claims, error)

  // IssueRefreshToken tạo 256-bit random opaque string.
  func IssueRefreshToken() (string, error)

  // HashToken tạo SHA-256 hash của token để lưu Redis (không lưu raw token).
  func HashToken(token string) string
  ```

- [ ] JWT structure:
  ```
  Header: { "alg": "HS256", "typ": "JWT", "kid": "<key-id>" }
  Payload: {
    "enc": "<base64url(AES-GCM(claims_json))>",
    "cnf": { "jkt": "<sha256(public_key_pem)>" },  // Desktop only
    "exp": <unix_timestamp>
  }
  ```

### Unit Tests — `pkg/auth/jwt_test.go`

- [ ] `TestIssueVerifyRoundTrip` — issue → verify trả đúng claims
- [ ] `TestExpiredToken` — token hết hạn → verify trả error
- [ ] `TestTamperedToken` — sửa payload → verify trả error
- [ ] `TestHardwareBoundToken` — verify claims có `cnf.jkt` đúng
- [ ] `TestWebTokenNoCnf` — Web token không có `cnf` claim → verify OK
- [ ] `TestRefreshTokenUnique` — 2 lần gọi `IssueRefreshToken` trả khác nhau
- [ ] Coverage ≥ 90%

---

## 3. Infrastructure — `internal/identity/infrastructure/`

### `UserRepositoryPG`

- [ ] `internal/identity/infrastructure/user_repository_pg.go`:
  - Implement `UserRepository` interface
  - `Create`: insert với `email_encrypted` và `email_hmac` từ `FieldCipher`
  - `GetByEmailHMAC`: `WHERE email_hmac = $1` (không cần decrypt để tìm)
  - `GetByUsername`: `WHERE username = $1`
  - `List`: với pagination OFFSET/LIMIT, total count
  - Transaction support qua `pgxpool.Tx`

### `DeviceRepositoryPG`

- [ ] `internal/identity/infrastructure/device_repository_pg.go`:
  - `Upsert`: `INSERT ... ON CONFLICT (user_id, device_fingerprint) DO UPDATE`
  - Lưu `public_key_pem` và `public_key_hash` (SHA-256 của PEM)

### `RoleRepositoryPG`

- [ ] `internal/identity/infrastructure/role_repository_pg.go`:
  - `GetByName`: join `roles` + `role_permissions` + `permissions`
  - `UpdatePermissions`: delete + insert batch trong transaction

---

## Definition of Done (Step 1)

- [ ] `go build ./internal/identity/...` không lỗi
- [ ] `go test ./pkg/auth/... -v -cover` → **tất cả test PASS**, coverage ≥ 90%
- [ ] Entity `User` có thể set/get email qua FieldCipher (encrypt/decrypt round-trip)
- [ ] `IssueAccessToken` → `VerifyAccessToken` round-trip đúng claims
- [ ] Token bị tamper → `VerifyAccessToken` trả error (không panic)
- [ ] Token hết hạn → trả error rõ ràng
- [ ] Repository interface compile được (chưa cần test tích hợp DB)
- [ ] `UserRepositoryPG` có thể compile và inject pgxpool
