# Báo cáo Tổng kết Giai đoạn 3 & 4 (Sprint 4: Desktop App & Real-time Khám Bệnh)

## 1. Tổng quan các hạng mục đã hoàn thành

Trong quá trình thực hiện Step 3 (Quản lý hàng đợi tiếp đón) và Step 4 (Quản lý ca khám của bác sĩ), chúng ta đã hoàn thiện toàn bộ luồng quy trình nghiệp vụ cốt lõi tại phòng khám từ lúc bệnh nhân bước vào cửa cho đến lúc kết thúc khám. Các hệ thống real-time và hệ thống quản trị danh mục đều đã đi vào hoạt động trơn tru.

### 1.1. Step 3: Quản lý Hàng Đợi (Queue Management - Lễ tân)
- **Giao diện Hàng Đợi (QueuePage):** Xây dựng giao diện bảng điều khiển hàng đợi trực quan bằng Ant Design cho Lễ tân, cho phép theo dõi danh sách bệnh nhân đang chờ, đang gọi, hoặc đã bỏ qua.
- **Tương tác Thời gian thực (Websocket):** Tích hợp thành công Wails Websocket Client với Go Backend (`pkg/ws/adapter.go`). Giao diện tự động cập nhật ngay lập tức khi có sự kiện thay đổi trạng thái (check-in, called, skipped, completed) thông qua kiến trúc Pub/Sub trên Redis.
- **Domain Reception:** Hoàn thiện các Command & Query (Check-in, CallQueue, SkipQueue) đảm bảo logic chuyển trạng thái được xác thực chặt chẽ tại Backend.
- **Quốc tế hóa (i18n):** Tích hợp đa ngôn ngữ toàn diện, bao gồm cả các thông báo Toast và nhãn trạng thái hàng đợi.

### 1.2. Step 4: Quản lý Ca Khám (Visit Management - Bác sĩ)
- **Worklist Bác sĩ:** Xây dựng danh sách bệnh nhân theo lịch phân công, chia theo phòng làm việc (Room-based Sharding). Tự động push data realtime khi bệnh nhân được chuyển từ hàng đợi vào phòng bác sĩ.
- **Màn hình Khám Bệnh (Visit Screen):**
  - **Sinh hiệu (Vitals):** Xây dựng form nhập Sinh hiệu kèm cảnh báo màu đỏ trực quan khi các chỉ số (Huyết áp, Nhịp tim, Nhiệt độ...) vượt ngưỡng an toàn. Lưu lịch sử sinh hiệu.
  - **Chỉ định (Orders):** Tích hợp module tìm kiếm mã bệnh ICD-10 (tối ưu hiệu năng bằng cơ chế Debounce 300ms) để thêm các chỉ định xét nghiệm/chẩn đoán hình ảnh.
  - **Lịch sử khám:** Tab theo dõi các thông tin y tế trước đây của bệnh nhân.
- **Đóng ca khám (Close Visit):** Xử lý luồng kết thúc ca khám, cập nhật trạng thái `COMPLETED` song song cho cả `Queue` và `Visit`.
- **Tối ưu UX / Cache:** 
  - Khắc phục triệt để lỗi lưu cache trên bộ nhớ (React Query) khi chuyển tài khoản Lễ tân -> Admin.
  - Fix lỗi giật chớp ("bóng ma" UI) khi click chuyển đổi giữa các bệnh nhân bằng cách chủ động clear state trong Zustand `visitStore`.
  - Fix lỗi hiển thị sai Role "ADMIN" khi đăng nhập qua luồng bảo mật 2 lớp (MFA).

## 2. Các vấn đề Kỹ thuật đã xử lý & Best Practices áp dụng

1. **Bảo mật tối đa với Zero-Trust & Encryption:**
   - Hoàn thiện luồng đăng nhập yêu cầu chữ ký phần cứng (Hardware Signature) qua Wails + Golang.
   - Các API nhạy cảm (Tìm kiếm user) được thiết kế đặc biệt với `email_hmac` để có thể tìm kiếm dữ liệu mã hóa mà không làm lộ dữ liệu thật trên RAM Database.
   - Quản lý mã lỗi 429 (Too Many Requests), 423 (Locked) minh bạch và chính xác trên cả luồng Login và MFA.

2. **Quản trị State chuẩn xác (Zustand + React Query):**
   - Kết hợp mượt mà giữa trạng thái API (React Query) và trạng thái ứng dụng (Zustand). Xử lý triệt để việc dọn dẹp cache `queryClient.clear()` khi `clearAuth()`.

3. **Room-based WebSocket Sharding:**
   - Thay vì Broadcast toàn bộ thông điệp đến mọi Client, hệ thống phân mảnh dữ liệu (Shard) theo `room_id`. Bác sĩ ở phòng nào chỉ nhận được event hàng đợi và thông tin của phòng đó, tối ưu băng thông cho hệ thống quy mô lớn.

## 3. Đánh giá & Bước tiếp theo

Hệ thống Desktop App hiện tại đã rất vững chắc cả về thiết kế kiến trúc lẫn trải nghiệm UI/UX. Chức năng chính cho 2 vai trò cốt cán (Lễ tân và Bác sĩ) đã đáp ứng được 90% các Use Case thực tế tại phòng khám.

**Next Steps (Chuẩn bị cho Step 5 & Sprint 5):**
- Phát triển phân hệ Xét nghiệm (Lab) & Dược (Pharmacy) để tiếp nhận các chỉ định từ Bác sĩ.
- Tinh chỉnh cơ chế Caching/Offline mode (nếu cần) cho Wails.
- Chuẩn bị dữ liệu mẫu (Seeding) phong phú hơn để test hiệu năng khi có hàng nghìn bệnh nhân cùng giao dịch.
