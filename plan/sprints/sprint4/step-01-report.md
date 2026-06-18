# Báo Cáo Hoàn Thành - Step 1: Reception Domain & WebSocket API (Sprint 4)

**Ngày hoàn thành:** [Ngày hiện tại]

## 1. Tổng quan
Trong Step 1, chúng ta đã hoàn thành việc xây dựng nền tảng cốt lõi cho module Reception (Tiếp đón bệnh nhân). Trọng tâm của bước này là xây dựng hệ thống **Quản lý hàng đợi (Queue Management)** và cơ sở hạ tầng giao tiếp thời gian thực bằng **WebSocket**.

Đặc biệt, hệ thống đã được thiết kế lại để WebSocket hoạt động dưới dạng một Microservice độc lập (`cmd/websocket/main.go`), giúp tối ưu hóa hiệu năng, tách biệt tài nguyên và dễ dàng scale server realtime qua Redis PubSub.

## 2. Các hạng mục đã hoàn thành

### 2.1. Cơ sở dữ liệu (Database Migration)
- Đã khởi tạo schema cho bảng `queue_entries` qua file migration `000011_reception_queue_schema.up.sql`.
- Đã thiết lập các trường dữ liệu quan trọng như: `patient_id`, `service_type` (GENERAL, LAB, RADIOLOGY,...), `queue_number` và các trạng thái của hàng đợi (`WAITING`, `CALLED`, `IN_PROGRESS`, `DONE`, `SKIPPED`).
- Áp dụng Partial Indexing theo thời gian (`created_at`) để tối ưu hóa truy vấn hàng đợi trong ngày.

### 2.2. Domain & Application Layers
- **Domain Layer:** Xây dựng `QueueEntry` entity và interface `QueueRepository`.
- **Application Layer:** Đã triển khai kiến trúc CQRS với các commands xử lý logic:
  - `CheckInCommand`: Đăng ký số thứ tự (tự động gen prefix như `KB001`, `XN002`).
  - `CallQueueCommand`: Gọi bệnh nhân tiếp theo.
  - `SkipQueueCommand`: Bỏ qua lượt.
  - `CompleteQueueCommand`: Hoàn thành khám/tiếp đón.
  - `GetCurrentQueueQuery`: Lấy danh sách chờ.
  - `GetQueueStatsQuery`: Lấy thống kê số lượng và thời gian chờ trung bình.
- Mỗi khi có sự thay đổi trạng thái (checkin, call, complete...), hệ thống sẽ tự động broadcast event JSON qua WebSocket.

### 2.3. WebSocket & Redis PubSub
- Sử dụng thư viện hiệu năng cao `github.com/thanhbvha/go-common/websocket`.
- **Custom Fiber Adapter**: Đã custom lại wrapper adapter để hỗ trợ hook `OnConnect`. Hook này đóng vai trò quan trọng, giúp tự động gửi toàn bộ danh sách Queue hiện tại xuống Client ngay khi handshake thành công.
- Tách tiến trình WebSocket ra khỏi API service (`cmd/websocket/main.go`). Service này lắng nghe trên port mặc định `8081`, giao tiếp với API service qua cơ chế PubSub của Redis.

### 2.4. Infrastructure & HTTP Handlers
- Đã triển khai Repository bằng `pgxpool` để tương tác trực tiếp với PostgreSQL.
- Định tuyến các API HTTP cơ bản tại `cmd/api/main.go` dưới middleware `JWTAuth`.
- Cấu hình Makefile chuẩn với các mục tiêu `build-ws`, `dev-ws`.
- Khắc phục các vấn đề liên quan đến UTF-16 Encoding trong file `.env`.

## 3. Kết quả đánh giá
- Toàn bộ source code biên dịch thành công (`go build ./...`).
- Các service HTTP và WebSocket đã có thể khởi chạy song song không xung đột port (8080 và 8081).
- Chức năng Heartbeat của WebSocket và PubSub chạy ổn định.

## 4. Bước tiếp theo
Từ báo cáo này, chúng ta đã sẵn sàng chuyển sang:
- **Step 2**: Xây dựng **Visit Domain, Vitals & Orders** (Khám bệnh, Dấu hiệu sinh tồn, Chỉ định dịch vụ).
- Hoặc có thể tiến hành tích hợp Frontend (React/Zustand) để bắt kết nối WebSocket và hiển thị Dashboard hàng đợi trực quan.
