# HIS DESKTOP PLAN — Wails + React (Staff App)

> **Cập nhật:** 2026-06-11
> **Stack:** Wails v2 · React 18 · TypeScript · Ant Design · TanStack Query · Zustand
> **Target:** Nhân viên nội bộ: Doctor, Nurse, Lab Tech, Pharmacist, Receptionist, Admin

**Tài liệu liên quan:**
- 🔧 Backend API: [backend-plan.md](./backend-plan.md)
- 🌐 Web client: [web-plan.md](./web-plan.md)

---

## 1. TỔNG QUAN

Ứng dụng Desktop chạy trên máy tính nội bộ phòng khám, kết nối đến backend qua **localhost HTTP** (hoặc LAN). Mỗi role sẽ thấy một giao diện khác nhau dựa trên RBAC.

```
DESKTOP APP (Wails v2)
      │
      │ localhost HTTP / LAN
      ▼
Go Fiber API (/api/v1)
```

> ⚠️ **NOTE:** Desktop giao tiếp qua **HTTP (không phải HTTPS)** khi dùng localhost.
> Nếu deploy trên LAN cần cân nhắc mTLS hoặc VPN tunnel.

---

## 2. TECH STACK

```
Wails v2 + React 18 + TypeScript
TanStack Query v5   — server state, cache, refetch
Zustand             — global UI state (auth, sidebar)
Ant Design 5.x      — UI components (phù hợp nghiệp vụ phức tạp)
Cornerstone.js      — DICOM viewer (Phase 3)
Recharts            — biểu đồ dashboard
React-to-print      — in ấn hóa đơn, phiếu kết quả
i18next             — VI/EN
React Router v6     — routing
Zod                 — schema validation
```

---

## 3. KIẾN TRÚC CLIENT

### Auth Flow
```
Login screen
  └──► POST /api/v1/auth/login
          └──► Nhận access_token (AES-GCM encrypted payload)
                  └──► Lưu vào memory (Zustand) — KHÔNG lưu localStorage
                  └──► Refresh token lưu HttpOnly Cookie (nếu Wails hỗ trợ)
                  └──► Nếu MFA enabled → MFA verify screen
                          └──► POST /api/v1/auth/mfa/verify
```

> ⚠️ **NOTE:** JWT payload được encrypt bằng AES-GCM ở server — client KHÔNG decode
> payload. Chỉ forward token trong Authorization header. Server tự decrypt.

### API Client Pattern
```typescript
// src/lib/apiClient.ts
// Axios instance với interceptor:
// - Tự động attach Authorization: Bearer {access_token}
// - Auto refresh khi nhận 401 (dùng refresh token)
// - Retry 1 lần sau refresh
```

> ⚠️ **NOTE:** Phải implement token refresh interceptor trước khi build bất kỳ feature nào.
> Đây là prerequisite quan trọng.

### Folder Structure
```
desktop/
├── src/
│   ├── lib/
│   │   ├── apiClient.ts       # Axios + interceptors
│   │   ├── queryClient.ts     # TanStack Query config
│   │   └── websocket.ts       # WS client (queue realtime)
│   ├── store/
│   │   ├── authStore.ts       # access_token, user, role
│   │   └── uiStore.ts         # sidebar, modals
│   ├── hooks/
│   │   ├── useAuth.ts
│   │   ├── useQueue.ts        # WebSocket queue hook
│   │   └── usePatient.ts
│   ├── modules/
│   │   ├── auth/              # Login, MFA screens
│   │   ├── admin/             # User mgmt, config, audit
│   │   ├── receptionist/      # Queue, check-in, billing
│   │   ├── doctor/            # Worklist, visit, EMR
│   │   ├── labtech/           # Sample tracking, results
│   │   ├── pharmacist/        # Prescription queue, dispense
│   │   └── shared/            # Patient search, common modals
│   └── layouts/
│       └── RoleLayout.tsx     # Sidebar theo role
└── wails.json
```

---

## 4. MÀN HÌNH THEO ROLE

### 4.1 Auth (tất cả roles)

**Login Screen**
- Form: username + password
- Gọi: `POST /api/v1/auth/login`
- Nếu `mfa_required: true` → redirect MFA screen

**MFA Screen**
- Nhập TOTP code (6 chữ số)
- Gọi: `POST /api/v1/auth/mfa/verify`

> 🔗 **Backend API:** Sprint 2 — Identity module
> ⚠️ **NOTE:** Lưu token trong Zustand memory, không persist sang localStorage để tránh XSS.

---

### 4.2 Lễ Tân (RECEPTIONIST)

**Queue Dashboard** (màn hình chính)
- Hiển thị realtime số thứ tự theo từng phòng/dịch vụ
- Kết nối: `WS /api/v1/queue/ws`
- Các action: Gọi số, Bỏ qua, Reset queue

> 🔗 **Backend API:** Sprint 4 — WebSocket queue endpoint
> ⚠️ **NOTE:** WebSocket phải implement reconnect logic với exponential backoff.
> Server push event khi có thay đổi queue — client không poll.

**Check-in Walk-in**
- Tìm bệnh nhân: search theo SĐT hoặc CCCD (gọi `GET /api/v1/patients?q=...`)
- Nếu chưa có → form đăng ký nhanh (`POST /api/v1/patients`)
- Assign queue number: `POST /api/v1/queue/checkin`

> 🔗 **Backend API:** Sprint 3 (Patient) + Sprint 4 (Queue)
> ⚠️ **NOTE:** Search patient phải debounce 300ms để tránh spam API.

**Appointment Management**
- Xem lịch theo ngày/tuần (calendar view)
- Đặt lịch thủ công: `POST /api/v1/appointments`
- Hủy/reschedule: `PUT/DELETE /api/v1/appointments/:id`
- Gọi: `GET /api/v1/appointments?date=...&doctor_id=...`

**Billing (Thu ngân)**
- Xem hóa đơn: `GET /api/v1/billing/invoices?visit_id=...`
- Thu tiền: `POST /api/v1/billing/invoices/:id/pay`
- In biên lai: `GET /api/v1/billing/invoices/:id/pdf` → React-to-print

> 🔗 **Backend API:** Sprint 7 — Billing module
> ⚠️ **NOTE:** Invoice tự động tạo sau `VisitClosed` — lễ tân chỉ cần confirm thanh toán.

---

### 4.3 Bác Sĩ (DOCTOR)

**Worklist**
- Danh sách bệnh nhân chờ khám trong ca hiện tại
- Gọi: `GET /api/v1/visits?status=waiting&doctor_id=me&date=today`
- Realtime update qua WebSocket (khi bệnh nhân check-in)

> 🔗 **Backend API:** Sprint 4 — Visit + Queue WebSocket

**Visit Screen (Khám bệnh)**

*Step 1: Vitals*
- Nhập: huyết áp, mạch, nhiệt độ, SpO2, cân nặng, chiều cao
- Gọi: `POST /api/v1/visits/:id/vitals`

*Step 2: SOAP Editor*
- Structured form: Subjective / Objective / Assessment / Plan
- Rich text hoặc structured fields
- Gọi: `POST /api/v1/emr` (create/update với versioning)

> ⚠️ **NOTE:** SOAP editor cần auto-save mỗi 30 giây để tránh mất dữ liệu.
> Implement optimistic update + conflict detection khi 2 user cùng sửa.

*Step 3: Diagnosis*
- ICD-10 search (autocomplete)
- Gọi: `GET /api/v1/icd10/search?q=...`
- Gắn chẩn đoán vào visit: `POST /api/v1/visits/:id/diagnoses`

*Step 4: Orders*
- Tạo lab order: `POST /api/v1/visits/:id/orders` (type: lab)
- Tạo radiology order: `POST /api/v1/visits/:id/orders` (type: radiology)
- Kê đơn thuốc: `POST /api/v1/pharmacy/prescriptions`

> 🔗 **Backend API:** Sprint 4 (Visit) + Sprint 5 (EMR) + Sprint 6 (Pharmacy)

**EMR Viewer**
- Xem toàn bộ lịch sử bệnh án theo patient_id
- Gọi: `GET /api/v1/emr/:patient_id/history`
- Version comparison view

**Kết quả LIS/RIS inline**
- Xem kết quả xét nghiệm ngay trong màn hình khám
- Gọi: `GET /api/v1/lab/results/:id`
- Gọi: `GET /api/v1/radiology/reports/:id`

> 🔗 **Backend API:** Sprint 5 (LIS result) — Phase 3 (RIS full)

**Drug Interaction Check**
- Khi kê đơn, check realtime: `GET /api/v1/drugs/:id/interactions?with=drug_id1,drug_id2`
- Hiển thị warning modal nếu có tương tác nguy hiểm

> ⚠️ **NOTE:** Drug interaction check phải block submit nếu severity = CONTRAINDICATED.

---

### 4.4 Kỹ Thuật Viên XN (LAB_TECH)

**Sample Tracking Worklist**
- Gọi: `GET /api/v1/lab/orders?status=pending`
- Màn hình kanban: Pending → Received → Processing → Completed

**Nhận mẫu**
- Scan barcode hoặc nhập manually
- Gọi: `PUT /api/v1/lab/orders/:id/sample`

> ⚠️ **NOTE:** Wails cần access serial port cho barcode scanner. Implement Wails binding
> cho Go serial library.

**Nhập kết quả**
- Form nhập từng test item với reference range hiển thị
- Highlight bất thường (H/L)
- Gọi: `POST /api/v1/lab/orders/:id/results`

**Verify & Approve**
- Review kết quả, ký duyệt
- Gọi: `PUT /api/v1/lab/orders/:id/verify`
- Sau verify → trigger `LabResultReady` event → bệnh nhân nhận notification

> ⚠️ **NOTE:** Sau verify, kết quả mới visible trên Web app của bệnh nhân.
> Đây là điểm kết nối quan trọng với [web-plan.md](./web-plan.md) — Web poll hoặc nhận
> notification khi `LabResultReady`.

**In phiếu kết quả**
- React-to-print → in phiếu PDF
- Gọi: `GET /api/v1/lab/results/:id` → render template → print

---

### 4.5 Dược Sĩ (PHARMACIST)

**Prescription Queue**
- Danh sách đơn thuốc theo status: Pending / Verified / Dispensed
- Gọi: `GET /api/v1/pharmacy/prescriptions?status=pending`
- Realtime update khi Doctor kê đơn mới (WebSocket hoặc polling 10s)

**Drug Interaction Review**
- Hiển thị chi tiết tương tác thuốc trước khi duyệt
- Gọi: `GET /api/v1/drugs/:id/interactions`

**Dispensing**
- Xác nhận xuất thuốc: `PUT /api/v1/pharmacy/prescriptions/:id/dispense`
- In nhãn thuốc: React-to-print
- Stock tự động giảm ở backend

> 🔗 **Backend API:** Sprint 6 — Pharmacy + Inventory
> ⚠️ **NOTE:** Dispensing phải confirm tồn kho trước khi submit. Backend trả lỗi
> nếu stock không đủ — frontend phải xử lý error case này rõ ràng.

**Inventory View**
- Xem tồn kho thuốc: `GET /api/v1/inventory/items`
- Filter theo warehouse, low-stock alert

---

### 4.6 Admin

**User Management**
- CRUD users: `GET/POST/PUT /api/v1/users`
- Assign roles: `PUT /api/v1/users/:id/roles`
- Deactivate: `PUT /api/v1/users/:id/deactivate`

**Role & Permission Matrix**
- Xem và cập nhật permission per role
- Gọi: `GET /api/v1/roles`, `PUT /api/v1/roles/:id/permissions`

**System Configuration**
- Giờ làm việc, dịch vụ, bảng giá
- Department management

**Dashboard & Reporting**
- Revenue chart (Recharts): `GET /api/v1/reports/revenue?from=...&to=...`
- Patient count, top services
- Date range picker

> 🔗 **Backend API:** Sprint 8 — Reporting APIs

**Audit Log Viewer**
- Xem audit trail: `GET /api/v1/audit/logs?entity=...&from=...`
- Filter theo user, action, entity type
- Highlight sensitive operations

> ⚠️ **NOTE:** Audit log chứa PII (user_id, entity_id). Chỉ ADMIN mới có quyền xem.
> Backend phải kiểm tra RBAC nghiêm ngặt cho endpoint này.

---

## 5. SHARED COMPONENTS

### Patient Search Modal
- Dùng chung cho: Receptionist check-in, Doctor worklist, Lab tech
- Gọi: `GET /api/v1/patients?q={input}` (debounce 300ms)
- Hiển thị: tên, SĐT (masked), ngày sinh, mã BN

> ⚠️ **NOTE:** Hiển thị SĐT dưới dạng masked (09x***xxx) để bảo vệ PII.
> Backend chỉ trả về masked version trong list response.

### Print Engine
- Wrapper React-to-print dùng chung
- Templates: hóa đơn, biên lai, phiếu kết quả XN, nhãn thuốc

### Notification Toast
- Hiển thị khi nhận WebSocket event liên quan đến role hiện tại
- Ví dụ: Doctor nhận toast khi lab result ready

---

## 6. WEBSOCKET MANAGEMENT

```typescript
// src/lib/websocket.ts
class WSClient {
  connect(url: string): void
  reconnect(): void  // exponential backoff: 1s, 2s, 4s, 8s, max 30s
  on(event: string, handler: Function): void
  disconnect(): void
}
```

**Events nhận từ server:**
- `queue.updated` → cập nhật Queue Dashboard
- `lab_result.ready` → Doctor nhận notification
- `prescription.created` → Pharmacist nhận notification

> ⚠️ **NOTE:** WebSocket token authentication: gửi token qua query param
> `?token=...` hoặc header khi connect. Backend phải validate.

---

## 7. ROADMAP DESKTOP — PHASE 1 (Tuần 1–16)

### Sprint 1 (Tuần 1–2): Foundation
- [ ] Init Wails v2 + React 18 + TypeScript
- [ ] Design system: color palette, typography (Ant Design theme)
- [ ] Base layout: Sidebar (role-based), Header, Content
- [ ] Axios API client + TanStack Query setup
- [ ] Zustand stores: authStore, uiStore
- [ ] Token refresh interceptor
- [ ] WebSocket client base class

> ⚠️ **NOTE:** KHÔNG build feature nào cho đến khi token refresh interceptor hoạt động đúng.

### Sprint 2 (Tuần 3–4): Auth
- [ ] Login screen (username + password form)
- [ ] MFA screen (TOTP 6-digit input)
- [ ] Role-based routing (redirect sau login theo role)
- [ ] User management UI (Admin)
- [ ] Role & Permission matrix UI

> 🔗 **Backend prerequisite:** Sprint 2 Identity APIs phải sẵn sàng

### Sprint 3 (Tuần 5–6): Patient & Appointment
- [ ] Patient search modal (shared component)
- [ ] Patient registration form (Receptionist)
- [ ] Patient detail view
- [ ] Appointment calendar view (doctor schedule)
- [ ] Manual appointment booking (Receptionist)

> 🔗 **Backend prerequisite:** Sprint 3 Patient + Appointment APIs

### Sprint 4 (Tuần 7–8): Queue & Visit
- [ ] Queue Dashboard (WebSocket realtime)
- [ ] Check-in flow (search → assign queue number)
- [ ] Doctor worklist
- [ ] Visit screen: Vitals form

> 🔗 **Backend prerequisite:** Sprint 4 Reception + Visit + WebSocket

### Sprint 5 (Tuần 9–10): EMR
- [ ] SOAP editor (auto-save mỗi 30s)
- [ ] ICD-10 diagnosis picker (autocomplete)
- [ ] Order creation UI (lab/radiology/prescription)
- [ ] Patient history timeline
- [ ] Lab result inline viewer

> 🔗 **Backend prerequisite:** Sprint 5 EMR APIs

### Sprint 6 (Tuần 11–12): Pharmacy & Inventory
- [ ] Drug search + Prescription editor
- [ ] Drug interaction warning modal (block on CONTRAINDICATED)
- [ ] Pharmacist dispensing queue (polling/WS)
- [ ] Inventory stock view
- [ ] Print nhãn thuốc

> 🔗 **Backend prerequisite:** Sprint 6 Pharmacy + Inventory APIs

### Sprint 7 (Tuần 13–14): Billing
- [ ] Billing screen (hóa đơn từng visit)
- [ ] Payment confirm form
- [ ] Print hóa đơn / biên lai (React-to-print)
- [ ] Payment history list

> 🔗 **Backend prerequisite:** Sprint 7 Billing APIs

### Sprint 8 (Tuần 15–16): Dashboard & Polish
- [ ] Admin dashboard (Recharts: revenue, patient stats)
- [ ] Audit log viewer
- [ ] System config UI
- [ ] E2E test cơ bản
- [ ] Wails production build + code signing

> 🔗 **Backend prerequisite:** Sprint 8 Reporting + Audit APIs

---

## 8. LƯU Ý QUAN TRỌNG

| # | Vấn đề | Chi tiết |
|---|--------|---------|
| 1 | JWT không decode ở client | Server encrypt AES-GCM, client chỉ forward token |
| 2 | PII masking trong list | SĐT, CCCD hiển thị masked trong danh sách |
| 3 | WebSocket reconnect | Phải implement exponential backoff |
| 4 | SOAP auto-save | Mỗi 30s, tránh mất dữ liệu khi mất điện |
| 5 | Drug interaction block | CONTRAINDICATED phải block submit |
| 6 | Stock check trước dispense | Xử lý error case stock không đủ |
| 7 | Barcode scanner | Cần Wails binding Go serial port |
| 8 | Token trong memory | Không persist localStorage/sessionStorage |
| 9 | localhost vs LAN | HTTP cho localhost, cân nhắc mTLS cho LAN |
| 10 | Lab result → Web sync | Sau Lab verify, Web mới được xem kết quả |

---

## 9. API CHECKLIST (Cross-reference với Backend)

> Dùng checklist này để đảm bảo không missing API khi develop:

- [ ] `POST /auth/login` — Sprint 2
- [ ] `POST /auth/mfa/verify` — Sprint 2
- [ ] `POST /auth/refresh` — Sprint 2
- [ ] `GET/POST /patients` — Sprint 3
- [ ] `GET/POST/PUT/DELETE /appointments` — Sprint 3
- [ ] `GET /appointments/slots` — Sprint 3
- [ ] `POST /queue/checkin` — Sprint 4
- [ ] `WS /queue/ws` — Sprint 4
- [ ] `GET/POST /visits` — Sprint 4
- [ ] `POST /visits/:id/vitals` — Sprint 4
- [ ] `POST /visits/:id/orders` — Sprint 4
- [ ] `POST /emr` — Sprint 5
- [ ] `GET /emr/:patient_id/history` — Sprint 5
- [ ] `GET /lab/orders` — Sprint 6 (LIS)
- [ ] `PUT /lab/orders/:id/sample` — Sprint 6
- [ ] `POST /lab/orders/:id/results` — Sprint 6
- [ ] `PUT /lab/orders/:id/verify` — Sprint 6
- [ ] `GET /pharmacy/prescriptions` — Sprint 6
- [ ] `PUT /pharmacy/prescriptions/:id/dispense` — Sprint 6
- [ ] `GET /drugs/search` — Sprint 6
- [ ] `GET /billing/invoices` — Sprint 7
- [ ] `POST /billing/invoices/:id/pay` — Sprint 7
- [ ] `GET /billing/invoices/:id/pdf` — Sprint 7
- [ ] `GET /reports/revenue` — Sprint 8
- [ ] `GET /audit/logs` — Sprint 8
