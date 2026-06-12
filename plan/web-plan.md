# HIS WEB PLAN — React + Vite (Patient Web App)

> **Cập nhật:** 2026-06-11
> **Stack:** React 18 · Vite · TypeScript · shadcn/ui · TanStack Query · Zustand
> **Target:** Bệnh nhân / Khách — đặt lịch, xem kết quả, quản lý hồ sơ

**Tài liệu liên quan:**
- 🔧 Backend API: [backend-plan.md](./backend-plan.md)
- 📱 Desktop client: [desktop-plan.md](./desktop-plan.md)

---

## 1. TỔNG QUAN

Web App dành cho **bệnh nhân** truy cập qua trình duyệt (HTTPS). Scope chức năng hẹp hơn Desktop — chỉ xem và tương tác với dữ liệu của chính bệnh nhân đó.

```
WEB APP (React+Vite)
      │
      │ HTTPS
      ▼
Go Fiber API (/api/v1)
 → Auth scope: PATIENT role
 → Chỉ xem data của chính mình
```

> ⚠️ **NOTE:** Web app KHÔNG có RBAC phức tạp như Desktop. Backend phải enforce
> rằng patient chỉ xem được data của chính họ (patient_id match với JWT claims).

---

## 2. TECH STACK

```
React 18 + TypeScript
Vite 5
TanStack Query v5      — data fetching, cache
Zustand                — auth state, UI state
React Hook Form + Zod  — form validation
shadcn/ui              — UI components
Tailwind CSS           — styling
React Router v6        — routing
i18next                — VI/EN internationalization
```

---

## 3. PAGES & ROUTES

| Route | Tên | Auth Required |
|-------|-----|:---:|
| `/` | Landing page | ❌ |
| `/login` | Đăng nhập | ❌ |
| `/register` | Đăng ký tài khoản | ❌ |
| `/book` | Đặt lịch khám | ✅ |
| `/my-appointments` | Lịch hẹn của tôi | ✅ |
| `/results` | Kết quả xét nghiệm | ✅ |
| `/account` | Hồ sơ cá nhân | ✅ |

---

## 4. CHI TIẾT TỪNG PAGE

### 4.1 Landing Page (`/`)

**Nội dung:**
- Giới thiệu phòng khám (tên, địa chỉ, chuyên khoa)
- Danh sách bác sĩ + chuyên môn
- Dịch vụ & bảng giá cơ bản
- CTA: "Đặt lịch ngay"
- Thông tin liên hệ, bản đồ

**API gọi (public, không cần auth):**
```
GET /api/v1/public/clinic-info        → thông tin phòng khám
GET /api/v1/public/doctors            → danh sách bác sĩ (public profile)
GET /api/v1/public/services           → danh sách dịch vụ + giá
```

> ⚠️ **NOTE:** Backend cần thêm `/api/v1/public/*` endpoints (không cần auth).
> Đây là endpoints mà backend-plan cần bổ sung — check với [backend-plan.md](./backend-plan.md).

---

### 4.2 Đăng Ký (`/register`)

**Flow:**
```
Nhập SĐT
  └──► POST /api/v1/auth/otp/send
          └──► Nhập OTP (6 số, TTL 5 phút)
                  └──► POST /api/v1/auth/otp/verify
                          └──► Nhập thông tin cá nhân
                                  └──► POST /api/v1/auth/register
                                          └──► Nhận access_token → redirect /book
```

**Form thông tin cá nhân:**
- Họ tên, Ngày sinh, Giới tính
- SĐT (đã verify qua OTP)
- Email (optional)
- CCCD (optional, dùng cho BHYT sau)

> 🔗 **Backend API:** Sprint 2 — auth/otp/send, auth/otp/verify, auth/register
> ⚠️ **NOTE:** Backend phải rate-limit OTP gửi: max 3 lần/SĐT/giờ.
> Frontend phải hiển thị countdown timer 60s trước khi cho phép resend OTP.

---

### 4.3 Đăng Nhập (`/login`)

**Flow:**
```
Nhập SĐT
  └──► POST /api/v1/auth/otp/send
          └──► Nhập OTP
                  └──► POST /api/v1/auth/otp/verify
                          └──► Nhận access_token, refresh_token
                                  └──► Redirect trang trước đó hoặc /my-appointments
```

> ⚠️ **NOTE:** Web app dùng **OTP via SĐT** để login, KHÔNG dùng password.
> Khác với Desktop dùng username/password + TOTP MFA.
> Backend phải có 2 flow login riêng biệt.

**Token storage:**
- `access_token`: memory (Zustand)
- `refresh_token`: HttpOnly Cookie (backend set via `Set-Cookie`)

> ⚠️ **NOTE:** Refresh token phải được backend set qua HttpOnly Cookie (không qua JSON body)
> để bảo vệ khỏi XSS. Cần CORS config đúng để cookie hoạt động cross-origin.

---

### 4.4 Đặt Lịch (`/book`)

**Multi-step flow:**

**Step 1: Chọn dịch vụ**
```
GET /api/v1/public/services
→ Hiển thị grid dịch vụ (Khám tổng quát, Tim mạch, ...)
```

**Step 2: Chọn bác sĩ**
```
GET /api/v1/public/doctors?service_id=...
→ Danh sách bác sĩ theo dịch vụ
```

**Step 3: Chọn ngày & giờ**
```
GET /api/v1/appointments/slots?doctor_id=...&date=...
→ Danh sách slot còn trống
→ Calendar picker (highlight ngày có slot)
```

> ⚠️ **NOTE:** Slots phải refresh khi chọn ngày khác. Slot có thể bị book bởi người khác
> giữa lúc user đang chọn. Backend phải check lại khi submit — trả lỗi `409 Conflict`
> nếu slot đã đầy. Frontend xử lý: hiển thị "Slot vừa được đặt, vui lòng chọn lại".

**Step 4: Xác nhận**
```
Hiển thị tóm tắt: bác sĩ, dịch vụ, ngày giờ, phí dịch vụ
POST /api/v1/appointments
  body: { doctor_id, service_id, slot_id, note }
→ Nhận appointment_id
→ Redirect /my-appointments
```

> 🔗 **Backend API:** Sprint 3 — Appointment CRUD + Slots API
> ⚠️ **NOTE:** Sau khi đặt lịch, backend publish `AppointmentScheduled` event → notification
> worker gửi SMS xác nhận cho bệnh nhân. Web chỉ cần hiển thị success message.

---

### 4.5 Lịch Hẹn Của Tôi (`/my-appointments`)

**Tabs:**
- **Upcoming:** lịch hẹn sắp tới (có thể hủy nếu còn > 24h)
- **History:** lịch khám đã qua

```
GET /api/v1/appointments?patient_id=me&status=upcoming
GET /api/v1/appointments?patient_id=me&status=past
```

**Hủy lịch:**
```
DELETE /api/v1/appointments/:id
→ Chỉ cho phép hủy nếu còn > 24h trước giờ hẹn
→ Backend enforce rule này, frontend hiển thị điều kiện
```

**Card mỗi lịch hẹn:**
- Tên bác sĩ, dịch vụ, ngày giờ, trạng thái
- Badge màu theo status: Chờ xác nhận / Đã xác nhận / Đã hoàn thành / Đã hủy

> 🔗 **Backend API:** Sprint 3 — Appointment list + cancel
> ⚠️ **NOTE:** Backend cần filter `patient_id = me` từ JWT claims.
> Không cho phép patient xem appointment của người khác.

---

### 4.6 Kết Quả Xét Nghiệm (`/results`)

**Hiển thị:**
- Danh sách lần xét nghiệm (theo visit/ngày)
- Chỉ hiển thị kết quả đã được **Lab verify** (`status = verified`)

```
GET /api/v1/patients/me/lab-results
  → list các order đã verify, sort by date desc
```

**Chi tiết từng kết quả:**
```
GET /api/v1/lab/results/:id
  → { items: [{ name, value, unit, reference_range, flag }] }
```

**Hiển thị:**
- Bảng kết quả với highlight bất thường (H = cao, L = thấp)
- Reference range bên cạnh
- Ngày lấy mẫu, ngày có kết quả
- Tên bác sĩ chỉ định

> 🔗 **Backend API:** Sprint 5 (LIS result) + Sprint 5 (EMR)
> ⚠️ **CRITICAL:** Web chỉ được xem kết quả sau khi Lab Tech **verify**.
> Đây là điểm kết nối với [desktop-plan.md](./desktop-plan.md) — Lab Tech verify
> (`PUT /lab/orders/:id/verify`) → trigger `LabResultReady` event → Web có thể xem.
> Backend phải filter `status = verified` trong query cho patient endpoint.

**Notification khi có kết quả mới:**
- SMS từ notification worker (backend)
- Web có thể poll hoặc dùng SSE (Server-Sent Events) để push notification

> ⚠️ **NOTE:** Phase 1 dùng polling mỗi 5 phút là đủ. SSE/WebSocket cho Phase 2.

---

### 4.7 Hồ Sơ Cá Nhân (`/account`)

**Tabs:**
- **Thông tin cá nhân:** cập nhật họ tên, ngày sinh, giới tính, email
- **Bảo hiểm (BHYT):** số thẻ, hạn thẻ, mức hưởng
- **Lịch sử khám:** tóm tắt các lần khám

```
GET  /api/v1/patients/me              → thông tin hiện tại (PII masked)
PUT  /api/v1/patients/me              → cập nhật thông tin
GET  /api/v1/patients/me/insurance    → thông tin BHYT
PUT  /api/v1/patients/me/insurance    → cập nhật BHYT
GET  /api/v1/patients/me/visits       → lịch sử khám (summary)
```

> ⚠️ **NOTE:** Khi hiển thị SĐT và CCCD, chỉ hiển thị masked version (09x***xxx).
> Chỉ ADMIN/RECEPTIONIST mới xem được full value qua Desktop.
> Backend phải trả masked value cho patient endpoint.

**Đổi SĐT:**
- Yêu cầu verify OTP SĐT mới trước khi update
- Flow: nhập SĐT mới → OTP verify → cập nhật

> ⚠️ **NOTE:** Đây là sensitive operation — backend phải re-encrypt AES-GCM với số mới
> và cập nhật `phone_hmac`.

---

## 5. KIẾN TRÚC CLIENT

### Auth Flow (Web)

```
Access protected route
  └──► Kiểm tra access_token trong Zustand
          │
          ├── Có token (còn hạn) → proceed
          │
          └── Không có / hết hạn
                  └──► POST /api/v1/auth/refresh (dùng HttpOnly Cookie)
                          ├── 200 OK → cập nhật access_token → proceed
                          └── 401 → redirect /login
```

### API Client Pattern

```typescript
// src/lib/apiClient.ts
// Axios instance:
// - baseURL: import.meta.env.VITE_API_URL
// - withCredentials: true (để gửi HttpOnly Cookie cho refresh)
// - Interceptor: auto refresh khi 401
// - Interceptor: attach Bearer token từ Zustand
```

### Folder Structure

```
web/
├── src/
│   ├── lib/
│   │   ├── apiClient.ts       # Axios + interceptors
│   │   └── queryClient.ts     # TanStack Query config
│   ├── store/
│   │   ├── authStore.ts       # access_token, patient profile
│   │   └── bookingStore.ts    # booking flow state (multi-step)
│   ├── hooks/
│   │   ├── useAuth.ts
│   │   ├── useAppointments.ts
│   │   ├── useLabResults.ts
│   │   └── usePatientProfile.ts
│   ├── pages/
│   │   ├── Landing/
│   │   ├── Login/
│   │   ├── Register/
│   │   ├── Book/              # Multi-step booking
│   │   ├── MyAppointments/
│   │   ├── Results/
│   │   └── Account/
│   ├── components/
│   │   ├── ui/                # shadcn/ui components
│   │   ├── AppointmentCard/
│   │   ├── LabResultTable/
│   │   ├── OTPInput/
│   │   └── StepWizard/        # Multi-step booking
│   └── layouts/
│       ├── PublicLayout.tsx
│       └── AuthLayout.tsx     # Header + Nav cho logged-in pages
└── vite.config.ts
```

---

## 6. SECURITY & UX NOTES

| # | Vấn đề | Chi tiết |
|---|--------|---------|
| 1 | Refresh token HttpOnly Cookie | Không lưu refresh token trong JS scope |
| 2 | Access token memory only | Zustand store, xóa khi tab close |
| 3 | PII masking | SĐT, CCCD luôn masked trong UI |
| 4 | Lab result gating | Chỉ hiển thị sau Lab verify — check status = verified |
| 5 | Slot conflict | Xử lý 409 Conflict khi submit booking |
| 6 | OTP rate limit | Countdown 60s + max 3 lần/giờ message |
| 7 | Patient scope | Backend enforce patient chỉ xem data của mình |
| 8 | CORS config | `withCredentials: true` cần CORS `Allow-Credentials: true` |
| 9 | Public endpoints | `/public/*` không cần auth, có thể cache CDN |
| 10 | Cancel appointment | Chỉ cho phép nếu còn > 24h — backend enforce, frontend hiển thị |

---

## 7. ROADMAP WEB — PHASE 1 (Tuần 1–16)

### Sprint 1 (Tuần 1–2): Foundation
- [ ] Init React 18 + Vite + TypeScript
- [ ] shadcn/ui setup + Tailwind CSS config
- [ ] Design system: màu sắc y tế (xanh lam/trắng), typography
- [ ] React Router v6 setup (public + protected routes)
- [ ] Axios API client + TanStack Query setup
- [ ] Zustand: authStore + bookingStore
- [ ] Token refresh interceptor
- [ ] i18next setup (VI mặc định)
- [ ] PublicLayout + AuthLayout

### Sprint 2 (Tuần 3–4): Auth
- [ ] Login page (OTP via SĐT)
- [ ] Register page (OTP → form thông tin)
- [ ] OTP input component (6 ô, auto-focus, paste support)
- [ ] Countdown timer component (resend OTP)
- [ ] Protected route HOC/wrapper
- [ ] Auth state persistence khi reload (dùng refresh token cookie)

> 🔗 **Backend prerequisite:** Sprint 2 — OTP send/verify, register APIs
> ⚠️ **NOTE:** Cần test kỹ flow "reload page → auto refresh token → không bị logout".

### Sprint 3 (Tuần 5–6): Landing + Booking
- [ ] Landing page: clinic info, doctors list, services
- [ ] Booking flow (4 steps: service → doctor → slot → confirm)
- [ ] Calendar slot picker component
- [ ] Booking confirmation + success screen

> 🔗 **Backend prerequisite:** Sprint 3 — Appointment + Slots + Public endpoints

### Sprint 4 (Tuần 7–8): My Appointments
- [ ] My appointments page (Upcoming tab + History tab)
- [ ] Appointment card component
- [ ] Cancel appointment (confirm dialog + check 24h rule)
- [ ] Empty state (khi chưa có lịch)

> 🔗 **Backend prerequisite:** Sprint 3 — Appointment list + cancel

### Sprint 5 (Tuần 9–10): Lab Results
- [ ] Lab results list (per visit, sort by date)
- [ ] Lab result detail table (value, reference range, flag H/L)
- [ ] Abnormal highlight (màu đỏ/cam cho H/L)
- [ ] Empty state (khi chưa có kết quả)
- [ ] Polling mỗi 5 phút cho kết quả mới

> 🔗 **Backend prerequisite:** Sprint 5 — LIS result verified API
> ⚠️ **NOTE:** Kết quả chỉ hiện sau khi Lab Tech verify trên Desktop app.
> Đây là cross-client dependency quan trọng.

### Sprint 6–7 (Tuần 11–14): Account
- [ ] Account page: thông tin cá nhân + edit form
- [ ] BHYT: xem + cập nhật
- [ ] Lịch sử khám (summary list)
- [ ] Đổi SĐT flow (OTP verify)
- [ ] Invoice view (`GET /billing/invoices/:id`)

> 🔗 **Backend prerequisite:** Sprint 3 (Patient) + Sprint 7 (Billing)

### Sprint 8 (Tuần 15–16): Polish
- [ ] Responsive design (mobile-first)
- [ ] Loading states, skeleton UI
- [ ] Error boundaries
- [ ] 404 / 403 pages
- [ ] SEO: meta tags, Open Graph
- [ ] Performance: lazy loading routes
- [ ] E2E test cơ bản

---

## 8. API CHECKLIST (Cross-reference với Backend)

> Dùng checklist này để đảm bảo không missing API khi develop:

**Public (không cần auth):**
- [ ] `GET /public/clinic-info` — Sprint 1 backend
- [ ] `GET /public/doctors` — Sprint 1 backend
- [ ] `GET /public/services` — Sprint 1 backend

**Auth:**
- [ ] `POST /auth/otp/send` — Sprint 2
- [ ] `POST /auth/otp/verify` — Sprint 2
- [ ] `POST /auth/register` — Sprint 2
- [ ] `POST /auth/refresh` — Sprint 2 (cookie-based)
- [ ] `POST /auth/logout` — Sprint 2

**Patient (scope: me):**
- [ ] `GET /patients/me` — Sprint 3
- [ ] `PUT /patients/me` — Sprint 3
- [ ] `GET /patients/me/insurance` — Sprint 3
- [ ] `PUT /patients/me/insurance` — Sprint 3
- [ ] `GET /patients/me/visits` — Sprint 4
- [ ] `GET /patients/me/lab-results` — Sprint 5

**Appointment:**
- [ ] `GET /appointments/slots?doctor_id=&date=` — Sprint 3
- [ ] `GET /doctors?service_id=` — Sprint 3
- [ ] `POST /appointments` — Sprint 3
- [ ] `GET /appointments?patient_id=me` — Sprint 3
- [ ] `DELETE /appointments/:id` — Sprint 3

**Lab Results:**
- [ ] `GET /lab/results/:id` — Sprint 5

**Billing:**
- [ ] `GET /billing/invoices/:id` — Sprint 7

---

## 9. CROSS-CLIENT DEPENDENCIES

> Những điểm mà Web và Desktop phải phối hợp — không được develop độc lập:

| # | Sự kiện | Desktop action | Web effect |
|---|---------|---------------|------------|
| 1 | Lab result verify | Lab Tech click "Verify" trên Desktop | Kết quả hiện trên `/results` của bệnh nhân |
| 2 | Appointment confirmed | Receptionist confirm trên Desktop | Badge "Đã xác nhận" trên `/my-appointments` |
| 3 | Visit closed | Doctor close visit trên Desktop | Invoice tạo tự động, bệnh nhân xem được |
| 4 | Appointment cancelled (by staff) | Receptionist hủy lịch trên Desktop | Bệnh nhân nhận SMS + thấy status "Đã hủy" |

> ⚠️ **NOTE:** Các dependency này đều thông qua backend events (Redis Stream) + notification worker.
> Web không cần real-time connection — SMS/email notification là đủ cho Phase 1.
