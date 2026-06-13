# Báo Cáo Tổng Kết Sprint 2: Identity & Authentication

**Thời gian thực hiện:** Tuần 3 - Tuần 4
**Mục tiêu chính:** Hoàn thiện hệ thống xác thực bảo mật cao (Identity & Auth) bao gồm JWT mã hoá AES-GCM, RBAC (Role-Based Access Control), phần cứng (Hardware Key Binding), TOTP MFA cho Desktop và OTP SĐT cho Web.

---

## I. Tổng Quan Thành Quả Sprint 2

Trong Sprint 2, dự án đã thiết lập thành công toàn bộ kiến trúc phân quyền và xác thực ở cả 3 layer: **Backend (Go/Fiber)**, **Desktop Client (Wails/React)** và **Web Client (React)**. Hệ thống đã đáp ứng tiêu chuẩn bảo mật khắt khe cấp y tế, chống lại các rủi ro đánh cắp token (Token Theft) bằng Hardware Key Binding, cũng như đảm bảo tính linh hoạt trong quản trị người dùng.

---

## II. Chi Tiết Công Việc Đã Hoàn Thành Theo Từng Step

### Step 1: Identity Domain & JWT Infrastructure
- **Domain Models:** Khởi tạo các entity cốt lõi (`User`, `Role`, `Permission`, `Device`, `MFA`, `Department`, `Patient`) với thiết kế Clean Architecture.
- **Security:** Triển khai module `pkg/auth/jwt.go` hỗ trợ mã hoá payload JWT bằng AES-GCM. 
- **Database:** Tạo repository interfaces và thực thi Postgres repositories (PGX) cho các entities. Thực hiện migrations database hoàn chỉnh.

### Step 2: Desktop Auth API (Backend)
- Xây dựng luồng đăng nhập Challenge-Response bảo mật cao cho Desktop.
- Cấu hình API `/auth/login/init` (trả về challenge string).
- Cấu hình API `/auth/login/complete` (verify chữ ký ECDSA và sinh Access/Refresh Token).
- Xây dựng API cấu hình và xác thực TOTP MFA (`/auth/mfa/setup`, `/auth/mfa/verify`).
- Cơ chế Refresh Token Rotation lưu trên Redis.

### Step 3: Web Auth API (Backend)
- Thiết lập luồng đăng nhập cho bệnh nhân sử dụng Số Điện Thoại.
- Tích hợp mô phỏng Gửi OTP (`/auth/otp/send`) qua Redis Queue, hỗ trợ fallback logic (Zalo ZNS -> SMS).
- Hoàn thiện API `/auth/otp/verify` cấp Access Token (trả qua payload) và Refresh Token (trả qua HttpOnly Cookie chống XSS).
- API Đăng ký tài khoản bệnh nhân `/auth/register`.

### Step 4: RBAC Middleware & Admin APIs
- **Middlewares:** Viết các middleware kiểm soát truy cập mạnh mẽ:
  - `JWTAuth`: Giải mã và verify JWT.
  - `RequireRole` / `RequirePermission`: Phân quyền dựa trên Role và Permission nạp từ Redis.
  - `RequestSignature`: Bắt buộc ký mọi request API từ Desktop bằng Private Key.
- **Admin APIs:** Hoàn tất CRUD cho User, Role, Permission và Department. Quản lý trạng thái vô hiệu hoá nhân sự và gán quyền chủ động.

### Step 5: Desktop Frontend Auth (Wails + React)
- **Native Integration:** Gọi CGO (macOS Keychain) / Windows CNG (TPM) từ Wails để sinh và lưu trữ ECDSA Key Pair.
- **Interceptor:** Tích hợp `apiClient.ts` tự động ký điện tử bằng Hardware Key vào header `X-Signature` và `X-Timestamp`.
- **UI/UX:** Hoàn thiện luồng giao diện Login, chuyển hướng qua màn hình MFA Setup (quét QR code), màn hình nhập MFA, và cơ chế tự động làm mới token ngầm (Silent Refresh).

### Step 6: Web Frontend Auth (React)
- Triển khai giao diện cho bệnh nhân: Đăng nhập bằng SĐT, trang nhập OTP (với component đếm ngược gửi lại mã), trang Đăng ký.
- Cấu hình interceptor quản lý Cookie Refresh Token tự động.
- Quản lý state đăng nhập toàn cục sử dụng Zustand.

### Step 7: Admin UI Desktop
- Xây dựng giao diện Quản trị viên (Admin Dashboard) trên ứng dụng Desktop.
- **Quản lý Nhân viên:** Bảng danh sách nhân sự (tìm kiếm, phân trang), form tạo mới nhân viên (tự tạo random password), tính năng vô hiệu hóa (Deactivate) và phân quyền (Assign Role).
- **Phân quyền (RBAC):** Bảng ma trận Role-Permission cho phép chỉnh sửa quyền theo từng chức năng (Menu, Action). 
- **Phòng ban:** Quản lý danh mục Khoa / Phòng ban.
- **Tối ưu SQL:** Tối ưu hóa API fetch vai trò hàng loạt (Bulk Fetch) để giải quyết lỗi N+1 Query.

---

## III. Các Lỗi Nghiệp Vụ & Kỹ Thuật Đã Xử Lý (Bug Fixes & Stabilizations)

1. **Bug Tự Động Logout sau 15 phút (Desktop):** 
   - Đã xử lý triệt để lỗi người dùng bị đăng xuất do thiếu cơ chế lưu lại Refresh Token sau khi Refresh Rotation. Interceptor hiện tại cập nhật đồng bộ cả Access và Refresh Token mới vào Zustand persist storage.
2. **Lỗi Lưu Phân Quyền 500 (Admin UI):** 
   - Refactor repository `role_repository_pg` dùng SQL `INSERT INTO ... SELECT` để tra cứu động UUID của Resource/Action từ tên string do Frontend gửi lên.
3. **Mất Đồng Bộ Schema Department:** 
   - Bổ sung trường `code` (Mã khoa) vào Model, Repository, Frontend và viết Database Migration cập nhật hồi tố cho dữ liệu cũ, đảm bảo tính nguyên vẹn dữ liệu (Full-stack data parity).
4. **Lỗi Brute-Force Nhận Diện Sai (MFA Flow):** 
   - Sửa lỗi thư viện Redis Wrapper gây lỗi double-prefix khi gọi hàm `Delete`, dẫn đến việc không reset được bộ đếm "login_attempts" khi đăng nhập đúng mật khẩu. Đã chuyển sang dùng `Native().Del()` cho tính toàn vẹn tuyệt đối.
   - Sửa lỗi nuốt Exception `401 Unauthorized` của luồng Refresh Token trong interceptor, giúp UI hiển thị đúng cảnh báo "Sai mật khẩu" thay vì "Lỗi hệ thống", giúp người dùng không bị nhầm lẫn và tránh spam click.

---

## IV. Đánh Giá Chung & Bước Tiếp Theo

**Kết luận:** Sprint 2 đã hoàn thành 100% mục tiêu đề ra. Nền tảng xác thực đa lớp hiện tại cực kỳ vững chắc, không chỉ đáp ứng chức năng cơ bản mà còn chống lại các hình thức tấn công bảo mật hiện đại.

**Hướng tiếp cận Sprint 3 (Medical Core & Workflow):**
- Với hệ thống User, Role, Department đã sẵn sàng, Sprint 3 sẽ tập trung vào xây dựng Core Y Tế: Tiếp đón bệnh nhân, Khám bệnh, Chỉ định cận lâm sàng và Quản lý Dược.
- Kế thừa hệ thống `apiClient` có chữ ký phần cứng và RBAC để bảo vệ dữ liệu sức khoẻ (PHI) của bệnh nhân.
