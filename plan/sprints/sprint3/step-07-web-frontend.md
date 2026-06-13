# Sprint 3 — Step 7: Web App — Patient Portal & Booking Flow (Frontend)

## Mục tiêu
Xây dựng trang Web cho bệnh nhân: Landing page, Luồng đặt lịch 4 bước, và trang quản lý lịch hẹn cá nhân.

## Files cần tạo
```
web/frontend/src/
├── stores/bookingStore.ts
├── pages/LandingPage.tsx
├── pages/BookingPage.tsx        -- 4-step flow
├── pages/MyAppointmentsPage.tsx
└── components/booking/
    ├── StepService.tsx
    ├── StepDoctor.tsx
    ├── StepSlot.tsx
    └── StepConfirm.tsx
```

## Nhiệm vụ

### 1. Landing Page (`/`)
- Section Giới thiệu: `GET /public/clinic-info`.
- Section Bác sĩ: `GET /public/doctors` (card: ảnh, tên, chuyên khoa).
- Section Dịch vụ: `GET /public/services` (giá, thời gian).
- CTA "Đặt lịch ngay" → `/book` (redirect `/login?returnUrl=/book` nếu chưa login).

### 2. Booking Flow (`/book`) — 4 Bước
Dùng Zustand `bookingStore` persist state qua các bước:

| Bước | UI | API |
|------|----|-----|
| 1 Chọn dịch vụ | Grid cards | `GET /public/services` |
| 2 Chọn bác sĩ | Card bác sĩ | `GET /public/doctors?service_id=` |
| 3 Chọn ngày & slot | Calendar + time grid | `GET /appointments/slots?doctor_id=&date=` |
| 4 Xác nhận | Summary + textarea ghi chú | `POST /appointments` |

**⚠️ Xử lý 409 Conflict ở bước 4:**
- Toast lỗi "Slot vừa được đặt bởi người khác".
- Auto redirect về bước 3 để chọn lại.
- Refetch slots tự động để hiển thị trạng thái mới nhất.

### 3. My Appointments (`/my-appointments`)
- Tab **Upcoming**: `GET /appointments?patient_id=me&status=upcoming`.
- Tab **History**: `GET /appointments?patient_id=me&status=past`.
- Card appointment: bác sĩ, dịch vụ, ngày giờ, status badge.
- Nút **Hủy** chỉ hiện nếu `scheduled_at > now + 24h`.
- Confirm dialog trước khi hủy → `DELETE /appointments/:id`.
- Empty state + CTA "Đặt lịch ngay".
