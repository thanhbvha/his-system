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
