# HIS System (Hospital Information System)

Hệ thống quản lý bệnh viện toàn diện, bao gồm Backend (Go), ứng dụng Desktop cho bác sĩ/nhân viên (Wails + React), và Web App cho bệnh nhân (React + Vite).

---

## 🛠 Yêu cầu hệ thống
- **Docker & Docker Compose** (Để chạy PostgreSQL, MongoDB, Redis, Grafana, Jaeger)
- **Go 1.22+**
- **Node.js 20.x+** & **npm** (để chạy Web và Frontend của Desktop)
- **Wails CLI** (`go install github.com/wailsapp/wails/v2/cmd/wails@latest`)

---

## 🚀 Hướng dẫn khởi chạy

### 1. Khởi động Cơ sở dữ liệu (Database & Observability)
Hệ thống yêu cầu các database (Postgres, Mongo, Redis) phải chạy trước khi bật API.

```bash
cd backend
docker-compose up -d
```
*Ghi chú: Đợi khoảng 10-15s để các database khởi động hoàn tất. Các công cụ monitoring cũng sẽ chạy ở cổng 3000 (Grafana) và 16686 (Jaeger).*

---

### 2. Khởi chạy Backend API (Go)
Backend Fiber cung cấp các API endpoint và kết nối trực tiếp với Database.

Mở một terminal mới:
```bash
cd backend
go run ./cmd/api/main.go
```
- API sẽ chạy tại: `http://localhost:8080`
- Swagger UI (Docs): `http://localhost:8080/docs/tool`
- ReDoc (Docs): `http://localhost:8080/docs`
- Health Check: `http://localhost:8080/health`

---

### 3. Khởi chạy Ứng dụng Desktop (Bác sĩ & Nhân viên)
Ứng dụng Wails bọc giao diện React, gọi trực tiếp API `localhost:8080`. Dùng cho nội bộ bệnh viện.

Mở một terminal mới:
```bash
cd desktop
wails dev
```
*Ghi chú: Lần đầu chạy có thể tốn chút thời gian để Wails cài đặt go module và biên dịch React frontend. Khi hoàn tất, một cửa sổ Desktop native sẽ tự động bật lên.*

---

### 4. Khởi chạy Web App (Bệnh nhân)
Web app sử dụng Vite và Tailwind CSS, dành cho bệnh nhân tra cứu thông tin và đặt lịch.

Mở một terminal mới:
```bash
cd web
npm run dev
```
- Ứng dụng Web sẽ chạy tại: `http://localhost:5173`
- Mọi request từ Web `/api` sẽ tự động được proxy sang `localhost:8080` để giải quyết lỗi CORS trong lúc code.

---

## 📂 Cấu trúc dự án
- `/backend`: Mã nguồn Go, DB Migrations, scripts và docker-compose.
- `/desktop`: Mã nguồn Wails (Go wrapper) + Frontend React (nhân viên y tế).
- `/web`: Mã nguồn Vite + React + Tailwind (bệnh nhân).
- `/plan`: Lộ trình phát triển và các tài liệu thiết kế hệ thống.
