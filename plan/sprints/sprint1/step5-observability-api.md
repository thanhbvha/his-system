# Sprint 1 — Step 5: Observability & API Foundation

> **Mục tiêu:** Setup monitoring/tracing và các API endpoint nền tảng (health, swagger, middleware).
> **Phụ thuộc:** Step 4 (Fiber app bootstrap xong).
> **Output:** `/health`, `/ready` OK; Swagger UI hoạt động; Jaeger nhận traces.

---

## 1. Observability

### OpenTelemetry + Jaeger

- [ ] Cài packages:
  ```bash
  go get go.opentelemetry.io/otel
  go get go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc
  go get go.opentelemetry.io/otel/sdk/trace
  ```
- [ ] `pkg/telemetry/tracer.go`:
  ```go
  func InitTracer(serviceName, jaegerEndpoint string) (func(), error)
  // Trả về shutdown func để defer trong main
  ```
- [ ] Fiber middleware tự động tạo span cho mỗi request:
  ```go
  app.Use(otelfiber.Middleware())
  ```

### Prometheus Metrics

- [ ] Cài:
  ```bash
  go get github.com/ansrivas/fiberprometheus/v2
  ```
- [ ] Register middleware:
  ```go
  prometheus := fiberprometheus.New("his-system")
  prometheus.RegisterAt(app, "/metrics")
  app.Use(prometheus.Middleware)
  ```
- [ ] Metrics thu thập:
  - `http_requests_total` (method, route, status_code)
  - `http_request_duration_seconds` (latency histogram)

### Logging

- [ ] `go-common/logger` output JSON → stdout
- [ ] Log format mẫu:
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
- [ ] PII fields tự động masked (phone → `***`, cccd → `***`)

---

## 2. API Foundation

### Health & Readiness Endpoints

- [ ] `GET /health` — liveness check:
  ```json
  { "status": "ok", "version": "1.0.0" }
  ```

- [ ] `GET /ready` — readiness check (kiểm tra DB, Redis):
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

### Swagger Documentation

- [ ] Cài swaggo:
  ```bash
  go install github.com/swaggo/swag/cmd/swag@latest
  go get github.com/swaggo/fiber-swagger
  go get github.com/swaggo/swag
  ```
- [ ] Thêm annotations vào `cmd/api/main.go`:
  ```go
  // @title HIS System API
  // @version 1.0
  // @description Hospital Information System API
  // @host localhost:8080
  // @BasePath /api/v1
  // @securityDefinitions.apikey BearerAuth
  // @in header
  // @name Authorization
  ```
- [ ] Route: `GET /swagger/*` → Swagger UI
- [ ] `make swag` → generate docs vào `docs/`

---

## 3. Middleware Stack

Thứ tự đăng ký middleware (quan trọng):

```go
// 1. Request ID — phải đầu tiên để gắn vào log
app.Use(requestid.New())

// 2. Logger middleware — log sau khi có request_id
app.Use(logger.New())

// 3. CORS
app.Use(cors.New(cors.Config{
    AllowOrigins: os.Getenv("CORS_ORIGINS"), // "http://localhost:5173,http://localhost:3000"
    AllowHeaders: "Origin, Content-Type, Accept, Authorization",
    AllowMethods: "GET, POST, PUT, PATCH, DELETE",
}))

// 4. Rate limiter (global)
app.Use(limiter.New(limiter.Config{
    Max:        100,           // 100 requests
    Expiration: 1 * time.Minute,
    KeyGenerator: func(c *fiber.Ctx) string {
        return c.IP()
    },
}))

// 5. Recover — phải cuối cùng để bắt panic
app.Use(recover.New())
```

- [ ] CORS config: whitelist `localhost:5173` (web Vite), `localhost:3000` (optional)
- [ ] Rate limit: 100 req/min per IP (global), các route nhạy cảm có limit riêng ở Sprint 2
- [ ] Request ID: UUID v4, attach vào response header `X-Request-ID`

---

## 4. Route Structure

```
/health         GET  (public)
/ready          GET  (public)
/metrics        GET  (public, Prometheus)
/swagger/*      GET  (public, dev only)

/api/v1/
  auth/         (Sprint 2)
  patients/     (Sprint 3)
  appointments/ (Sprint 3)
  ...
```

---

## Definition of Done (Step 5)

- [ ] `GET /health` → `200 OK` với JSON response
- [ ] `GET /ready` → `200 OK` khi PG + Mongo + Redis connected; `503` khi 1 service down
- [ ] `GET /swagger/` → Swagger UI load được trong browser
- [ ] `GET /metrics` → Prometheus metrics format
- [ ] Jaeger UI (`http://localhost:16686`) nhận traces từ API calls
- [ ] Request ID xuất hiện trong log và response header
- [ ] Rate limiter hoạt động (curl 101 lần → request thứ 101 trả `429`)
- [ ] CORS header đúng cho origin `localhost:5173`
