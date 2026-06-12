# Sprint 1 — Step 1: Khởi tạo Project & Infrastructure

> **Mục tiêu:** Dựng skeleton Go project và toàn bộ infra local bằng Docker Compose.
> **Phụ thuộc:** Không có — đây là bước đầu tiên.
> **Output:** `go mod init` xong, `docker-compose up` chạy tất cả service healthy.

---

## 1. Khởi tạo Go Project

- [ ] `go mod init his-system`
- [ ] Cài các dependencies core:
  ```bash
  go get github.com/gofiber/fiber/v2
  go get github.com/jackc/pgx/v5/pgxpool
  go get go.mongodb.org/mongo-driver/mongo
  go get github.com/thanhbvha/go-common   # logger, redis, queue, websocket
  ```
- [ ] Tạo cấu trúc thư mục theo plan:
  ```
  his-system/
  ├── cmd/
  │   ├── api/        # HTTP server entry point
  │   ├── worker/     # Redis Stream consumer worker
  │   └── migrate/    # DB migration runner
  ├── internal/
  │   ├── identity/
  │   ├── patient/
  │   └── ...
  └── pkg/
      ├── database/
      ├── storage/
      ├── errors/
      └── ...
  ```
- [ ] `Makefile` với các targets:
  ```makefile
  dev       # air hot-reload
  migrate   # chạy migration
  test      # go test ./...
  lint      # golangci-lint run
  swag      # generate swagger docs
  build     # go build -o bin/api cmd/api/main.go
  ```

---

## 2. Docker Compose Infrastructure

File: `docker-compose.yml`

- [ ] Các service cần thiết:
  | Service | Image | Port |
  |---------|-------|------|
  | `postgres` | `postgres:15-alpine` | `5432` |
  | `mongodb` | `mongo:7` | `27017` |
  | `redis` | `redis:7-alpine` | `6379` |
  | `minio` | `minio/minio` | `9000`, `9001` |
  | `jaeger` | `jaegertracing/all-in-one` | `16686`, `4317` |
  | `prometheus` | `prom/prometheus` | `9090` |
  | `grafana` | `grafana/grafana` | `3000` |

  > ⚠️ **NOTE:** Không cần service `nats` riêng — Redis Stream dùng chung container `redis`.

- [ ] Volume config:
  ```yaml
  volumes:
    postgres_data:
    mongodb_data:
    redis_data:
    minio_data:
  ```

- [ ] Network config:
  ```yaml
  networks:
    his-net:
      driver: bridge
  ```

- [ ] Health check cho từng service:
  ```yaml
  # Ví dụ postgres
  healthcheck:
    test: ["CMD-SHELL", "pg_isready -U postgres"]
    interval: 10s
    timeout: 5s
    retries: 5
  ```

- [ ] `.env` file mẫu:
  ```env
  POSTGRES_DSN=postgres://postgres:postgres@localhost:5432/his_db?sslmode=disable
  MONGO_URI=mongodb://localhost:27017/his_db
  REDIS_ADDR=localhost:6379
  MINIO_ENDPOINT=localhost:9000
  MINIO_ACCESS_KEY=minioadmin
  MINIO_SECRET_KEY=minioadmin
  FIELD_ENCRYPTION_KEY=<256-bit-base64-key>
  ```

---

## Definition of Done (Step 1)

- [ ] `go mod tidy` không lỗi
- [ ] `docker-compose up -d` → tất cả service status `healthy`
- [ ] Có thể kết nối PostgreSQL: `psql -h localhost -U postgres`
- [ ] Có thể kết nối Redis: `redis-cli ping` → `PONG`
- [ ] MinIO UI truy cập được tại `http://localhost:9001`
