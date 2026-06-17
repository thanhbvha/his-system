# Báo Cáo Tổng Kết Sprint 3: Patient & Appointment Management

**Thời gian thực hiện**: Tuần 5–6  
**Mục tiêu**: Xây dựng phân hệ Quản lý Bệnh nhân (Mã hoá PII) và Đặt lịch khám, hoàn thiện từ hệ thống Backend (Database, Domain, API) đến Frontend (Desktop App cho Nhân viên, Web App cho Bệnh nhân).

Sprint 3 đã kế thừa toàn bộ nền tảng xác thực (JWT, MFA, Hardware Key, RBAC) và hệ thống phân quyền từ Sprint 2, triển khai thành công 7 bước cốt lõi.

---

## Các Hạng Mục Đã Hoàn Thành

### 1. Database & Infrastructure (Step 1)
- **Migrations**: Tạo bảng `patients`, `patient_insurances`, `appointments`, `appointment_slots` với cấu trúc chặt chẽ.
- **Bảo mật dữ liệu (PII)**: Thiết kế các cột dữ liệu nhạy cảm (SĐT, CCCD, Email, Địa chỉ) dưới dạng `_encrypted` (AES-GCM) và `_hmac` (SHA-256) phục vụ cho Exact Search mà không làm rò rỉ dữ liệu gốc.
- Thêm cơ chế tìm kiếm Full-text search (tsvector) trên `full_name`.

### 2. Patient Domain & API (Step 2, 3, 4)
- **Domain Layer**: Khởi tạo Entity `Patient`, `PatientInsurance` kết hợp các Value Objects (`PhoneNumber`, `CCCD`) có khả năng tự động encrypt/decrypt và sinh mã HMAC.
- **Data Access**: `PatientRepositoryPG` tự động hoá quá trình mã hoá khi `Save` và giải mã khi `Get`. Đảm bảo không có plaintext PII lọt xuống DB.
- **APIs**:
  - `GET /patients`: Dành cho nhân viên, tự động che (mask) số điện thoại. Hỗ trợ tìm kiếm theo tên, SĐT (qua HMAC) và CCCD.
  - `POST /patients`, `GET /patients/:id`: Tạo mới và xem chi tiết (đầy đủ thông tin, không che) chỉ dành cho Staff (Receptionist, Doctor).
  - Tích hợp Middleware Auth & RBAC cho từng endpoint.

### 3. Appointment Backend & Redis Events (Step 5)
- **Luồng Booking**: Xây dựng API đặt lịch có kiểm soát xung đột (Concurrency Control) bằng kỹ thuật `SELECT FOR UPDATE`.
  - Khi có hai người cùng đặt một slot, database sẽ lock row và ngăn chặn double-booking, trả về lỗi `409 Conflict`.
- **Public API**: Cung cấp `/public/clinic-info`, `/public/doctors`, `/public/services` cho Landing Page của bệnh nhân.
- **Redis Stream**: Phát thành công các sự kiện như `AppointmentScheduled`, `PatientRegistered` để chuẩn bị cho Notification Worker (gửi SMS/Email) trong tương lai.

### 4. Desktop Frontend - Dành cho Nhân viên (Step 6)
- **Quản lý Bệnh nhân**: Form Đăng ký bệnh nhân mới có validate khắt khe (SĐT 10 số, CCCD 12 số) và Component tìm kiếm bệnh nhân (Debounce + Masked PII).
- **Quản lý Lịch hẹn**: Xem lịch khám dạng Calendar, cho phép Lễ tân Xác nhận (Confirm), Check-in hoặc Hủy lịch hẹn của bệnh nhân. Modal đặt lịch thủ công có tích hợp refetch tự động khi dính lỗi 409.
- **Đồng bộ Đa ngôn ngữ (i18n)**: Rà soát và đồng bộ lại cấu trúc i18n toàn bộ app (Login, MFA, Profile, Dashboard, Patients, Appointments, Admin Config), loại bỏ hoàn toàn tình trạng hardcode "nửa Anh nửa Việt".
- **User Profile Page**: Tính năng xem thông tin tài khoản hiện tại, lịch sử đăng nhập, thiết lập ngôn ngữ và kiểm tra MFA.

### 5. Web Frontend - Cổng thông tin Bệnh nhân (Step 7)
- **Landing Page**: Thiết kế hiện đại bằng Tailwind CSS + shadcn/ui. Tự động fetch thông tin phòng khám, dịch vụ và bác sĩ từ Public API.
- **Booking Flow (4 Bước)**: 
  - Quy trình đặt lịch mượt mà bằng Zustand (`bookingStore`): Chọn Dịch vụ ➔ Chọn Bác sĩ ➔ Chọn Khung giờ (Slots) ➔ Xác nhận.
  - Xử lý mượt lỗi trùng lịch (409): Cảnh báo UI và tự động đưa người dùng lùi về bước chọn giờ mà không mất data cũ.
- **My Appointments**: Bảng điều khiển quản lý lịch khám cá nhân chia theo tab (Sắp tới / Lịch sử), hỗ trợ huỷ lịch trực tiếp.

---

## Definition of Done (Đánh giá hoàn thành)

- [x] Patient CRUD hoạt động (PII encrypt/decrypt và mask đúng chuẩn bảo mật).
- [x] Search patient theo SĐT/CCCD (qua HMAC) / Tên (qua tsvector) hoạt động.
- [x] Double booking được prevent thành công bằng DB Row Lock.
- [x] Desktop: tìm kiếm + đăng ký bệnh nhân thành công.
- [x] Desktop: xem lịch hẹn theo ngày, đặt lịch thủ công, UI hoàn thiện i18n.
- [x] Web: Landing page hiển thị đúng thông tin động từ Backend.
- [x] Web: Booking flow 4 bước hoạt động end-to-end, bắt lỗi 409 đúng.
- [x] Web: My-appointments hiển thị và hủy được.
- [x] Redis Stream events được publish đúng kênh.
- [x] RBAC hoạt động chuẩn xác: PATIENT không truy cập được API của RECEPTIONIST.

**Trạng thái Sprint 3:** HOÀN THÀNH ✅  
**Sẵn sàng cho Sprint tiếp theo (Clinical & Pharmacy / Billing)**.
