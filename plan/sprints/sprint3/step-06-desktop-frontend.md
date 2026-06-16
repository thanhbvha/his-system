# Sprint 3 — Step 6: Desktop App — Receptionist Features (Frontend)

## Mục tiêu
Tích hợp giao diện Patient Management và Appointment Calendar vào Desktop App. Tận dụng `apiClient.ts` với Hardware Key signing đã có.

## Files cần tạo
```
desktop/frontend/src/
├── stores/patientStore.ts
├── stores/appointmentStore.ts
├── components/patient/PatientSearchModal.tsx
├── components/patient/PatientRegForm.tsx
├── components/patient/PatientDetailView.tsx
├── components/appointment/AppointmentCalendar.tsx
├── components/appointment/BookingModal.tsx
├── pages/PatientsPage.tsx
└── pages/AppointmentsPage.tsx
```

## Nhiệm vụ

### 1. Zustand Stores
- `patientStore`: search, selectedPatient, create actions.
- `appointmentStore`: fetchByDate, fetchSlots, book, confirm, cancel.

### 2. Patient Search Modal
- Debounce 300ms → `GET /patients?q={input}`.
- Hiển thị: Tên, SĐT masked `091***678`, Ngày sinh.
- ⚠️ Không hiển thị SĐT đầy đủ trong dropdown.

### 3. Patient Registration Form
- Ant Design Form: Họ tên, Ngày sinh, Giới tính, SĐT*, CCCD, Email, Địa chỉ.
- Validate client-side: SĐT 10 số, CCCD 12 số.
- Submit → `POST /patients` (apiClient tự ký Hardware Key).

### 4. Appointment Calendar
- Calendar view theo ngày → `GET /appointments?date={date}`.
- Status badge màu: 🟡 PENDING | 🔵 CONFIRMED | 🟢 CHECKED_IN | ⚫ COMPLETED | 🔴 CANCELLED.

### 5. Manual Booking Modal
- Chọn bác sĩ → dịch vụ → ngày → slot grid.
- Xử lý 409: toast + refetch slots tự động.

### 6. Navigation
Thêm sidebar menu: Tìm kiếm bệnh nhân, Đăng ký, Lịch hẹn hôm nay.

### 7. Bugfixes & Cải tiến hệ thống (Đã hoàn thành)
- **Sửa lỗi Panic Backend**: Khắc phục lỗi ép kiểu `interface conversion` trong `AppointmentHandler` và `PatientHandler`.
- **Sửa lỗi White Screen**: Sửa logic mapping data phân trang từ `res.data.data.items` trong `patientStore` và `appointmentStore`.
- **Sửa lỗi Auto Logout**: Dùng Regex thay vì `JSON.parse` để parse chuỗi payload AES-GCM khi check token expiration, giúp app không tự văng khi hot-reload.
- **Cấu hình Ngôn ngữ (Preferred Language)**: 
  - Thêm cột `preferred_language` vào table `users` (DB Migration).
  - Tích hợp API `PUT /api/v1/auth/me/language` (tuân thủ `x-signature` và `jwtAuth`).
  - Giao diện góc trên phải tự động đồng bộ và lưu thiết lập tiếng Anh/Việt trên hệ thống i18n và DB.

### 8. Ngày mai (Next Step): User Profile Page
- **Mục tiêu**: Xây dựng một trang Profile Page cho Desktop App.
- **Tính năng**: 
  - Ngay sau khi Login thành công, hệ thống sẽ tự động điều hướng (auto redirect) bay thẳng vào màn hình Profile này (thay vì vào Dashboard như cũ).
  - Hiển thị đầy đủ thông tin chi tiết: Thông tin User, Chức vụ/Phân quyền, Thông tin Phòng ban làm việc, Địa chỉ IP đăng nhập hiện tại, Ngày giờ đăng nhập, v.v.
