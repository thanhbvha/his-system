# Sprint 1 — Step 5: Observability & API Foundation

> **Mục tiêu:** Setup monitoring/tracing và các API endpoint nền tảng (health, swagger, middleware).
> **Phụ thuộc:** Step 4 (Fiber app bootstrap xong).
> **Output:** `/health`, `/ready` OK; Swagger UI hoạt động; Jaeger nhận traces.

---

## 1. Observability

### OpenTelemetry + Jaeger

- [x] Cài packages:
  ```bash
  go get go.opentelemetry.io/otel
  go get go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc
  go get go.opentelemetry.io/otel/sdk/trace
  ```
- [x] `pkg/telemetry/tracer.go`:
  ```go
  func InitTracer(serviceName, jaegerEndpoint string) (func(context.Context) error, error)
  // Trả về shutdown func để defer trong main
  ```
- [x] Fiber middleware tự động tạo span cho mỗi request:
  ```go
  app.Use(otelfiber.Middleware())
  ```

### Prometheus Metrics

- [x] Cài:
  ```bash
  go get github.com/ansrivas/fiberprometheus/v2
  ```
- [x] Register middleware:
  ```go
  prometheus := fiberprometheus.New("his-system")
  prometheus.RegisterAt(app, "/metrics")
  app.Use(prometheus.Middleware)
  ```
- [x] Metrics thu thập:
  - `http_requests_total` (method, route, status_code)
  - `http_request_duration_seconds` (latency histogram)

### Logging

- [x] `go-common/logger` output JSON → stdout
- [x] Log format mẫu (đã tích hợp tự động qua `pkg/middleware/logger.go` và `InfoAsync`):
  ```json
  {
    "level": "info",
    "time": "2025-01-01T00:00:00Z",
    "request_id": "uuid",
    "method": "POST",
    "path": "/api/v1/patients",
    "status": 201,
    "latency_ms": 45
  }
  ```
- [x] PII fields tự động masked (phone → `***`, cccd → `***`)

---

## 2. API Foundation

### Health & Readiness Endpoints

- [x] `GET /health` — liveness check:
  ```json
  { "status": "ok", "version": "1.0.0" }
  ```

- [x] `GET /ready` — readiness check (kiểm tra DB, Redis):
  ```json
  {
    "status": "ok",
    "checks": {
      "postgres": "ok",
      "mongodb": "ok",
      "redis": "ok"
    }
  }
  ```
  - Trả `503` nếu bất kỳ dependency nào fail

### API Documentation (Swagger & ReDoc)

- [x] Cài đặt `swaggo/swag` và generate spec json.
- [x] Thêm annotations vào `cmd/api/main.go`
- [x] Khởi tạo custom HTML template phục vụ 2 engine UI qua tệp nội bộ:
  - Tải CSS, JS, Bundle của Swagger và ReDoc lưu vào thư mục `public/docs/`.
  - Config **ReDoc** load Font nội bộ từ `https://fonts.googleapis.com/css?family=Montserrat:300,400,700|Roboto:300,400,700`.
- [x] Route: 
  - `GET /docs` → ReDoc UI
  - `GET /docs/tool` → Swagger UI
- [x] Lệnh: `swag init -g ./cmd/api/main.go --output ./docs`

---

## 3. Middleware Stack

Thứ tự đăng ký middleware (quan trọng):

```go
// 1. Request ID — gắn UUID cho header và context
app.Use(requestid.New())

// 2. Logger middleware — sử dụng pkg/middleware/logger.go tự viết
app.Use(middleware.RequestLogger(appLogger))

// 3. OpenTelemetry Tracer
app.Use(otelfiber.Middleware())

// 4. CORS
app.Use(cors.New(cors.Config{
    AllowOrigins: "http://localhost:5173",
    AllowHeaders: "Origin, Content-Type, Accept, Authorization",
    AllowMethods: "GET, POST, PUT, PATCH, DELETE",
}))

// 5. Rate limiter (global)
app.Use(limiter.New(limiter.Config{
    Max:        100,           // 100 requests
    Expiration: 1 * time.Minute,
    KeyGenerator: func(c *fiber.Ctx) string {
        return c.IP()
    },
    LimitReached: func(c *fiber.Ctx) error {
        return response.Fail(c, &appErrors.AppError{Code: "TOO_MANY_REQUESTS", Status: 429})
    },
}))

// 6. Recover — bắt panic không làm sập server
app.Use(middleware.Recover(appLogger))
```

- [x] CORS config: whitelist `localhost:5173` (web Vite)
- [x] Rate limit: 100 req/min per IP, giới hạn trả 429 qua custom JSON error.
- [x] Request ID: UUID v4, attach vào response header và log system.

---

## 4. Route Structure

```
/health         GET  (public)
/ready          GET  (public)
/metrics        GET  (public, Prometheus)
/docs/tool      GET  (public, Swagger UI)
/docs           GET  (public, ReDoc UI)

/api/v1/
  auth/         (Sprint 2)
  patients/     (Sprint 3)
  appointments/ (Sprint 3)
  ...
```

---

## Definition of Done (Step 5)

- [x] `GET /health` → `200 OK` với JSON response
- [x] `GET /ready` → `200 OK` khi PG + Mongo + Redis connected; `503` khi 1 service down
- [x] `GET /docs/tool` và `GET /docs` → Tải UI thành công qua internal JS/CSS.
- [x] `GET /metrics` → Prometheus metrics format
- [x] Jaeger UI (`http://localhost:16686`) nhận traces từ API calls (đã init OTLP)
- [x] Request ID xuất hiện trong log và response header
- [x] Rate limiter hoạt động (curl 101 lần → request thứ 101 trả `429 TOO_MANY_REQUESTS` dạng JSON chuẩn)
- [x] CORS header đúng cho origin `localhost:5173`
