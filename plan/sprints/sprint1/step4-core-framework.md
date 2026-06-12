# Sprint 1 — Step 4: Core Framework (go-common Integration)

> **Mục tiêu:** Tích hợp `github.com/thanhbvha/go-common` và xây dựng các shared package cốt lõi.
> **Phụ thuộc:** Step 1 (Go project init), Step 3 (Crypto package xong).
> **Output:** Tất cả shared package trong `pkg/` hoạt động, `cmd/api/main.go` bootstrap Fiber thành công.

---

## 1. Database Package — `pkg/database/`

### PostgreSQL (pgxpool)

- [ ] `pkg/database/postgres.go`:
  ```go
  func NewPostgresPool(dsn string) (*pgxpool.Pool, error)
  func MustNewPostgresPool(dsn string) *pgxpool.Pool // panic nếu không kết nối được
  ```
  - Cấu hình pool: `MaxConns`, `MinConns`, `MaxConnLifetime`
  - Health check: `pool.Ping(ctx)`

### MongoDB

- [ ] `pkg/database/mongo.go`:
  ```go
  func NewMongoClient(uri string) (*mongo.Client, error)
  func MustNewMongoClient(uri string) *mongo.Client
  ```

---

## 2. Tích hợp `github.com/thanhbvha/go-common`

### 2.1 Logger

- [ ] Init logger trong `cmd/api/main.go`:
  ```go
  import "github.com/thanhbvha/go-common/logger"

  log := logger.New(logger.Config{
      Level:  "info",
      Format: "json",       // structured JSON cho production
  })
  ```
- [ ] Thêm **PII masking hook** — tự động mask các field nhạy cảm khi log:
  ```go
  // Tạo wrapper trong pkg/logger/pii.go
  // Mask các field: phone, cccd, email, password khi xuất hiện trong log entry
  ```
- [ ] Attach logger vào Fiber context để dùng xuyên suốt request lifecycle

### 2.2 Redis (Cache)

- [ ] Init Redis client qua `go-common/redis`:
  ```go
  import "github.com/thanhbvha/go-common/redis"

  rdb := redis.New(redis.Config{
      Addr:     os.Getenv("REDIS_ADDR"),
      Password: os.Getenv("REDIS_PASSWORD"),
      DB:       0,
  })
  ```
- [ ] Expose qua `pkg/cache/cache.go` wrapper nếu cần thêm logic:
  - `Set(ctx, key, value, ttl)`
  - `Get(ctx, key)`
  - `Del(ctx, keys...)`
  - `SetEX(ctx, key, value, expiration)`

### 2.3 Queue (Redis Stream)

- [ ] Init queue qua `go-common/queue`:
  ```go
  import "github.com/thanhbvha/go-common/queue"

  q := queue.New(queue.Config{
      RedisAddr: os.Getenv("REDIS_ADDR"),
  })
  ```
- [ ] Expose interface trong `pkg/messaging/messaging.go`:
  ```go
  type EventPublisher interface {
      Publish(ctx context.Context, stream string, payload any) error
  }

  type EventSubscriber interface {
      Subscribe(ctx context.Context, stream, group, consumer string,
                handler func(msg queue.Message) error) error
  }
  ```
  > Interface hoá để có thể swap implementation (Redis → Kafka) ở Phase 4 mà không đổi business logic.

### 2.4 WebSocket

- [ ] Init WebSocket hub qua `go-common/websocket`:
  ```go
  import "github.com/thanhbvha/go-common/websocket"

  hub := websocket.NewHub()
  go hub.Run()
  ```
- [ ] Expose trong `pkg/ws/hub.go`:
  - `Register(client)`
  - `Broadcast(message)`
  - `BroadcastToRoom(roomID, message)`

---

## 3. Storage Package — `pkg/storage/`

- [ ] `pkg/storage/minio.go`:
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

## 4. Shared Utilities

### Error Types — `pkg/errors/`

- [ ] `pkg/errors/errors.go`:
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

- [ ] `pkg/response/response.go`:
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

- [ ] Khởi tạo app theo thứ tự:
  ```
  1. Load env (godotenv)
  2. Init logger (go-common/logger)
  3. Init DB (pgxpool, mongo)
  4. Init Redis (go-common/redis)
  5. Init Queue (go-common/queue)
  6. Init WebSocket hub (go-common/websocket)
  7. Init Storage (minio)
  8. Bootstrap Fiber app
  9. Register middleware
  10. Register routes
  11. Start server
  ```
- [ ] Graceful shutdown: bắt `SIGINT`, `SIGTERM` → drain connections

---

## Definition of Done (Step 4)

- [ ] `go build ./cmd/api/` không lỗi
- [ ] `go run ./cmd/api/` khởi động Fiber thành công, log xuất ra JSON
- [ ] Logger có PII masking (không log phone/cccd/email dạng plaintext)
- [ ] Redis ping thành công qua `go-common/redis`
- [ ] Queue publish/subscribe cơ bản test được (manual hoặc unit test)
- [ ] WebSocket hub khởi động không panic
- [ ] `Response.OK()` và `Response.Fail()` trả đúng format JSON
