# Sprint 1 — Step 6: CI/CD (GitHub Actions)

> **Mục tiêu:** Thiết lập pipeline CI tự động chạy lint, test và build check khi push/PR.
> **Phụ thuộc:** Step 3 (Crypto tests xong), Step 4 (go build OK).
> **Output:** GitHub Actions pipeline pass trên mọi PR vào `main`/`develop`.

---

## 1. Cấu trúc CI Pipeline

File: `.github/workflows/ci.yml`

```yaml
name: CI

on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main, develop]

jobs:
  ci:
    name: Lint, Test & Build
    runs-on: ubuntu-latest

    services:
      postgres:
        image: postgres:15-alpine
        env:
          POSTGRES_PASSWORD: postgres
          POSTGRES_DB: his_test
        ports: ["5432:5432"]
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5

      redis:
        image: redis:7-alpine
        ports: ["6379:6379"]
        options: >-
          --health-cmd "redis-cli ping"
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5

    steps:
      - uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: "1.22"
          cache: true

      - name: Install dependencies
        run: go mod download

      - name: Lint
        uses: golangci/golangci-lint-action@v4
        with:
          version: v1.57.2
          args: --timeout=5m

      - name: Test
        env:
          POSTGRES_DSN: postgres://postgres:postgres@localhost:5432/his_test?sslmode=disable
          REDIS_ADDR: localhost:6379
          FIELD_ENCRYPTION_KEY: dGVzdC1rZXktMzItYnl0ZXMtZm9yLXRlc3Rpbmch
          FIELD_HMAC_KEY: dGVzdC1obWFjLWtleS0zMi1ieXRlcy1mb3ItdGVzdA==
        run: go test ./... -v -race -coverprofile=coverage.out

      - name: Upload coverage
        uses: codecov/codecov-action@v4
        with:
          file: coverage.out

      - name: Build check
        run: go build -o /dev/null ./cmd/api/
```

---

## 2. Lint Config

File: `.golangci.yml`

- [ ] Tạo config với các linter bật:
  ```yaml
  linters:
    enable:
      - errcheck      # kiểm tra error không bị bỏ qua
      - gosimple      # code đơn giản hơn
      - govet         # go vet standard
      - ineffassign   # gán biến không dùng
      - staticcheck   # static analysis
      - unused        # code không dùng
      - gofmt         # format check
      - goimports     # import order
      - misspell      # typo checker
      - bodyclose     # http response body close check
      - noctx         # http request without context

  linters-settings:
    errcheck:
      check-type-assertions: true
    govet:
      enable-all: true
  ```

---

## 3. Pre-commit Hooks (tùy chọn, khuyến nghị)

- [ ] Cài `pre-commit`:
  ```bash
  pip install pre-commit
  ```
- [ ] File `.pre-commit-config.yaml`:
  ```yaml
  repos:
    - repo: local
      hooks:
        - id: go-fmt
          name: go fmt
          entry: gofmt -l -w
          language: system
          types: [go]
        - id: go-vet
          name: go vet
          entry: go vet ./...
          language: system
          pass_filenames: false
  ```

---

## 4. Makefile Targets CI-related

- [ ] Thêm vào `Makefile`:
  ```makefile
  lint:
      golangci-lint run --timeout=5m

  test:
      go test ./... -v -race -cover

  test-crypto:
      go test ./pkg/crypto/... -v -cover

  ci: lint test build
  ```

---

## Definition of Done (Step 6)

- [ ] Push lên GitHub → CI workflow trigger tự động
- [ ] `Lint` step: golangci-lint không có warning nào
- [ ] `Test` step: tất cả test pass, coverage xuất hiện trong report
- [ ] `Build` step: `go build ./cmd/api/` thành công
- [ ] CI badge hiển thị `passing` trên README
