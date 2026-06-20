# Báo Cáo Hoàn Thành - Step 2: Visit Domain + Vitals + Orders + ICD-10 API (Sprint 4)

**Ngày hoàn thành:** [Ngày hiện tại]

## 1. Tổng quan
Trong Step 2, chúng ta đã hoàn thành việc xây dựng module Khám Bệnh (`internal/visit`) phục vụ cho bác sĩ và điều dưỡng. Trọng tâm của bước này là quản lý vòng đời của một lượt khám (Visit), lưu trữ thông tin sinh tồn (Vitals), đặt các chỉ định cận lâm sàng (Orders), và cung cấp API Full-Text Search cho mã bệnh ICD-10.
Ngoài ra, kiến trúc Worker chạy ngầm độc lập đã được thiết lập thành công.

## 2. Các hạng mục đã hoàn thành

### 2.1. Cơ sở dữ liệu (Database Migration)
- Đã khắc phục lỗi migration cũ, file `000012_visit_schema.up.sql` đã chạy hoàn chỉnh, sinh ra 4 bảng: `visits`, `visit_vitals`, `visit_orders`, và `icd10_codes`.
- Đã thiết lập chỉ mục Full-Text Search (`idx_icd10_fts`) cho bảng `icd10_codes`.
- Đã áp dụng 22 mẫu dữ liệu mã ICD-10 seed vào DB để test search.

### 2.2. Domain & Application Layers
- **Domain Layer:** Xây dựng `Visit`, `VisitVital`, `VisitOrder`, và `ICD10Code`. Cài đặt State Machine cho Visit giúp ràng buộc quy trình cập nhật trạng thái (`REGISTERED` -> `WAITING` -> `IN_PROGRESS` -> `ORDERED` -> `COMPLETED`).
- **Application Layer:** 
  - `CreateVisitCommand`: Tạo thông tin khám.
  - `RecordVitalsCommand`: Ghi nhận huyết áp, mạch, nhiệt độ...
  - `CreateVisitOrderCommand`: Đặt chỉ định dịch vụ y tế.
  - `CloseVisitCommand`: Kết thúc phiên khám.
  - `GetDoctorWorklistQuery`: Query danh sách bệnh nhân dựa theo ngày và status (JOIN tables).
  - `SearchICD10Query`: Tra cứu ICD-10 sử dụng hàm `to_tsquery`.

### 2.3. Hệ Thống Worker Độc Lập
- Đã xây dựng ứng dụng Worker tách biệt ở `cmd/worker/main.go`, tận dụng thư viện `github.com/thanhbvha/go-common/queue`.
- Worker service kết nối Redis và lắng nghe 3 loại background job: `HIS.VISIT.VisitStarted`, `HIS.VISIT.LabOrderCreated`, và `HIS.VISIT.VisitClosed`.
- API service (Producer) tạo job và không bị block, trong khi Worker service (Consumer) âm thầm xử lý, tách rời hoàn toàn quá trình xử lý phụ tải.

### 2.4. Infrastructure & HTTP Handlers
- Các repository (`pgxpool`) được liên kết với Database phục vụ việc ghi và tra cứu.
- Định tuyến các API HTTP REST ở `visit_handler.go` bao gồm 9 endpoints hỗ trợ mọi tác vụ cho quy trình thăm khám.
- Đã cập nhật `Makefile` để hỗ trợ chạy 3 server riêng biệt:
  - `make dev`: Chạy API chính (8080)
  - `make dev-ws`: Chạy WebSocket (8081)
  - `make dev-worker`: Chạy Background Worker 

## 3. Kết quả đánh giá
- Database đã đồng bộ đầy đủ toàn bộ bảng.
- Source code biên dịch hoàn hảo (`go build ./...` Exit code 0).
- Hệ thống sẵn sàng để scale theo kiến trúc Microservices cơ bản.

## 4. Bước tiếp theo
Từ báo cáo này, chúng ta đã kết thúc quá trình làm việc trên backend cho Sprint 4, bước tiếp theo sẽ là tích hợp các API và dữ liệu realtime (WebSocket) này lên giao diện Web (Frontend/Web Portal).
