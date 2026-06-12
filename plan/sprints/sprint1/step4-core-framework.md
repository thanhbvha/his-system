# Sprint 1 — Step 4: Core Framework (go-common Integration)

> **Mục tiêu:** Tích hợp `github.com/thanhbvha/go-common` và xây dựng các shared package cốt lõi.
> **Phụ thuộc:** Step 1 (Go project init), Step 3 (Crypto package xong).
> **Output:** Tất cả shared package trong `pkg/` hoạt động, `cmd/api/main.go` bootstrap Fiber thành công.

---

## 1. Database Package — `pkg/database/`

### PostgreSQL (pgxpool)

- [x] `pkg/database/postgres.go`:
  ```go
  func NewPostgresPool(dsn string) (*pgxpool.Pool, error)
  func MustNewPostgresPool(dsn string) *pgxpool.Pool // panic nếu không kết nối được
  ```
  - Cấu hình pool: `MaxConns`, `MinConns`, `MaxConnLifetime`
  - Health check: `pool.Ping(ctx)`

### MongoDB

- [x] `pkg/database/mongo.go`:
  ```go
  func NewMongoClient(uri string) (*mongo.Client, error)
  func MustNewMongoClient(uri string) *mongo.Client
  ```

---

## 2. Tích hợp `github.com/thanhbvha/go-common`

### 2.1 Logger

- [x] Init logger trong `cmd/api/main.go`:
  ```go
  import "github.com/thanhbvha/go-common/logger"

  log := logger.New(logger.Config{
      Level:  "info",
      Format: "json",       // structured JSON cho production
  })
  ```
- [x] Thêm **PII masking hook** — tự động mask các field nhạy cảm khi log:
  ```go
  // Tạo wrapper trong pkg/logger/pii.go
  // Mask các field: phone, cccd, email, password khi xuất hiện trong log entry
  ```
- [x] Attach logger vào Fiber context để dùng xuyên suốt request lifecycle

### 2.2 Redis (Cache)

- [x] Init Redis client qua `go-common/redis`:
  ```go
  import "github.com/thanhbvha/go-common/redis"

  rdb := redis.New(redis.Config{
      Addr:     os.Getenv("REDIS_ADDR"),
      Password: os.Getenv("REDIS_PASSWORD"),
      DB:       0,
  })
  ```
- [x] Expose qua `pkg/cache/cache.go` wrapper nếu cần thêm logic:
  > **Note:** File `pkg/cache/cache.go` đã được gỡ bỏ vì `thanhbvha/go-common/redis` cung cấp native wrapper siêu mạnh trên struct `*redis.Client` (gồm các hàm `Set`, `Get`, `SetNX`, `Delete`, `HSet`, v.v. với context, expiration built-in). Việc tạo wrapper ngoài là dư thừa và giới hạn sức mạnh của client. Dùng trực tiếp `commonRedis.Client` ở các service.

### 2.3 Queue (Redis Stream)

- [x] Init queue qua `go-common/queue`:
  ```go
  import "github.com/thanhbvha/go-common/queue"

  q := queue.New(queue.Config{
      RedisAddr: os.Getenv("REDIS_ADDR"),
  })
  ```
- [x] Expose interface trong `pkg/messaging/messaging.go`:
  > **Note:** File này không cần thiết. `thanhbvha/go-common/redis` đã cung cấp sẵn `Publish` và `Subscribe`, và `thanhbvha/go-common/queue` cung cấp `Enqueue` (sử dụng Redis Streams với Consumer Groups, hỗ trợ dead-letter queue, retry limit, v.v.). Chúng ta sẽ sử dụng trực tiếp các method này để đảm bảo tính native và sức mạnh của thư viện.

### 2.4 WebSocket

- [x] Init WebSocket hub qua `go-common/websocket`:
  > **Note:** Đã loại bỏ việc tự viết `pkg/ws/hub.go`. Thay vào đó, sử dụng trực tiếp `commonWSFiber.NewHandler` (Adapter cho Fiber) và `commonWSCore.GetGlobalManager()` từ thư viện `thanhbvha/go-common`.
  > Chúng ta gắn trực tiếp các route WebSocket (`/ws`, `/api/ws/stats`, `/api/ws/shard`) vào `cmd/api/main.go` để tận dụng hệ thống Rate Limiting, Logging, và Authentication dùng chung ở Sprint 1. Đoạn code khởi chạy Manager được bọc bằng `utils.SafeGo`.

---

## 3. Storage Package — `pkg/storage/`

- [x] `pkg/storage/minio.go`:
  ```go
  func NewMinioClient(endpoint, accessKey, secretKey string, useSSL bool) (*minio.Client, error)

  type StorageClient interface {
      Upload(ctx context.Context, bucket, objectName string, reader io.Reader, size int64) error
      Download(ctx context.Context, bucket, objectName string) (io.ReadCloser, error)
      Delete(ctx context.Context, bucket, objectName string) error
      GetURL(ctx context.Context, bucket, objectName string, expires time.Duration) (string, error)
  }
  ```

---

## 4. Shared Utilities & Middleware

### Utilities — `pkg/utils/`

- [x] `pkg/utils/safego.go`:
  ```go
  // SafeGo chạy goroutine an toàn, tự động bắt panic và log qua Async Logger để không sập app.
  func SafeGo(fn func())
  ```

### Middlewares — `pkg/middleware/`

- [x] `pkg/middleware/recover.go`: Chặn panic, log Async ra console, và trả về mã lỗi 500 chuẩn API.
- [x] `pkg/middleware/logger.go`: Ghi nhận `method`, `path`, `status`, `duration_ms` cho từng request, tự động bỏ qua endpoint `/health`.


### Error Types — `pkg/errors/`

- [x] `pkg/errors/errors.go`:
  ```go
  type AppError struct {
      Code    string `json:"code"`
      Message string `json:"message"`
      Status  int    `json:"-"`
  }

  var (
      ErrNotFound     = &AppError{Code: "NOT_FOUND", Status: 404}
      ErrValidation   = &AppError{Code: "VALIDATION_ERROR", Status: 422}
      ErrUnauthorized = &AppError{Code: "UNAUTHORIZED", Status: 401}
      ErrForbidden    = &AppError{Code: "FORBIDDEN", Status: 403}
      ErrConflict     = &AppError{Code: "CONFLICT", Status: 409}
      ErrInternal     = &AppError{Code: "INTERNAL_ERROR", Status: 500}
  )
  ```

### Response Wrapper — `pkg/response/`

- [x] `pkg/response/response.go`:
  ```go
  type Response struct {
      Success bool        `json:"success"`
      Data    any         `json:"data,omitempty"`
      Error   *ErrorInfo  `json:"error,omitempty"`
      Meta    *Meta       `json:"meta,omitempty"`
  }

  type Meta struct {
      Page  int `json:"page"`
      Limit int `json:"limit"`
      Total int `json:"total"`
  }

  func OK(c *fiber.Ctx, data any) error
  func OKWithMeta(c *fiber.Ctx, data any, meta *Meta) error
  func Fail(c *fiber.Ctx, err *AppError) error
  ```

---

## 5. Fiber App Bootstrap — `cmd/api/main.go`

- [x] Khởi tạo app theo thứ tự:
  ```
  1. Load env (godotenv)
  2. Init logger (go-common/logger) -> Đã cấu hình dùng 100% InfoAsync/ErrorAsync
  3. Init DB (pgxpool, mongo)
  4. Init Redis (go-common/redis)
  5. Init Queue (go-common/queue)
  6. Init WebSocket Manager (go-common/websocket/adapter/fiber)
  7. Init Storage (minio)
  8. Bootstrap Fiber app
  9. Register middleware (Recover, RequestLogger)
  10. Register routes (REST + WebSocket)
  11. Start server (được bọc trong utils.SafeGo)
  ```
- [x] Graceful shutdown: bắt `SIGINT`, `SIGTERM` → drain connections

---

## Definition of Done (Step 4)

- [x] `go build ./cmd/api/` không lỗi
- [x] `go run ./cmd/api/` khởi động Fiber thành công, log xuất ra JSON
- [x] Logger có PII masking (không log phone/cccd/email dạng plaintext)
- [x] Redis ping thành công qua `go-common/redis`
- [x] Queue publish/subscribe cơ bản test được (manual hoặc unit test)
- [x] WebSocket hub khởi động không panic
- [x] `Response.OK()` và `Response.Fail()` trả đúng format JSON
