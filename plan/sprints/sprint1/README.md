# Sprint 1 — Foundation: Index

> **Mục tiêu sprint:** Dựng toàn bộ nền tảng dự án — backend skeleton, infra local, crypto package, và khởi tạo 2 app client.
> **Thời gian:** Tuần 1–2
> **Kết thúc sprint:** Hệ thống chạy được local, health check OK, 2 app client hiển thị layout cơ bản.

---

## Danh sách Steps

| Step | File | Mô tả | Phụ thuộc |
|------|------|-------|-----------|
| **Step 1** | [step1-project-init-infra.md](./step1-project-init-infra.md) | Khởi tạo Go project & Docker Compose infra | — |
| **Step 2** | [step2-database-migration.md](./step2-database-migration.md) | Database schema & migration (PG + Mongo) | Step 1 |
| **Step 3** | [step3-security-crypto.md](./step3-security-crypto.md) | Security package: AES-GCM + HMAC | Step 1 |
| **Step 4** | [step4-core-framework.md](./step4-core-framework.md) | Core framework: go-common (logger, redis, queue, ws) | Step 1, 3 |
| **Step 5** | [step5-observability-api.md](./step5-observability-api.md) | Observability (OTel, Prometheus) + API Foundation | Step 4 |
| **Step 6** | [step6-cicd.md](./step6-cicd.md) | GitHub Actions CI/CD pipeline | Step 3, 4 |
| **Step 7** | [step7-desktop-app.md](./step7-desktop-app.md) | Desktop app: Wails + React + Ant Design | Step 5 |
| **Step 8** | [step8-web-app.md](./step8-web-app.md) | Web app: React + Vite + shadcn/ui | Step 5 |

---

## Thứ tự thực hiện

```
Step 1 (Init & Infra)
    ├── Step 2 (Database)
    ├── Step 3 (Crypto) ← CRITICAL, xong trước step 4
    │       └── Step 4 (Core Framework / go-common)
    │               ├── Step 5 (Observability & API)
    │               │       ├── Step 7 (Desktop App)
    │               │       └── Step 8 (Web App)
    │               └── Step 6 (CI/CD)
```

---

## Definition of Done (Sprint 1)

- [ ] `docker-compose up` → tất cả service healthy
- [ ] `GET /health` trả `200 OK`
- [ ] `GET /ready` trả `200 OK` (PG + Mongo + Redis connected)
- [ ] `pkg/crypto` unit test pass 100%
- [ ] Migration 001–003 chạy thành công
- [ ] `go-common` logger, redis, queue, websocket tích hợp và khởi động OK
- [ ] Desktop app build và hiển thị layout cơ bản
- [ ] Web app chạy Vite dev server, hiển thị landing placeholder
- [ ] CI GitHub Actions pass trên mọi PR

---

## Thư viện chính sử dụng trong Sprint 1

| Package | Version | Mục đích |
|---------|---------|---------|
| `github.com/gofiber/fiber/v2` | v2.x | HTTP framework |
| `github.com/thanhbvha/go-common` | latest | Logger, Redis, Queue, WebSocket |
| `github.com/jackc/pgx/v5` | v5.x | PostgreSQL driver |
| `go.mongodb.org/mongo-driver` | v1.x | MongoDB driver |
| `github.com/golang-migrate/migrate/v4` | v4.x | DB migration |
| `go.opentelemetry.io/otel` | v1.x | Distributed tracing |
| `github.com/swaggo/fiber-swagger` | latest | Swagger UI |
| `github.com/minio/minio-go/v7` | v7.x | Object storage |
