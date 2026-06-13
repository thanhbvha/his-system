# Sprint 2 — Step 7: Admin UI (Desktop) — User & Role Management

> **Mục tiêu:** Xây dựng giao diện quản trị trên Desktop cho Admin: quản lý user (tạo/xem/deactivate/phân quyền) và ma trận Role-Permission.
> **Phụ thuộc:** Step 4 Admin APIs ready; Step 5 (apiClient với signature + RBAC hoạt động).
> **Output:** Admin đăng nhập Desktop và quản lý user, roles, permissions hoàn chỉnh.

---

## Các thành quả đã hoàn thành (Từ Step 1 đến Step 6)

- **Step 1 (Platform Infrastructure):** Thiết lập Platform Core (Fiber, Go-Common logger, Queue, WebSocket) và kết nối hạ tầng (PostgreSQL, MongoDB, Redis, MinIO).
- **Step 2 (Desktop Backend Auth):** Triển khai luồng đăng nhập Challenge-Response (`/auth/login/init`, `/auth/login/complete`), mã hoá payload bằng JWT + AES-GCM và MFA (TOTP).
- **Step 3 (Web Patient Auth):** Triển khai luồng xác thực bằng SĐT + OTP (SMS/Zalo) và đăng ký tài khoản bệnh nhân (`/auth/otp/send`, `/auth/otp/verify`, `/auth/register`).
- **Step 4 (RBAC & Admin API):** Hoàn tất Middleware kiểm soát truy cập (`JWTAuth`, `RequireRole`, `RequirePermission`, `RequestSignature`) cùng hệ thống API quản trị nhân viên và phòng ban.
- **Step 5 (Desktop Frontend Auth):** Tích hợp Native Hardware Keystore (Windows CNG TPM & macOS CGO Keychain) với Wails. Gắn chữ ký điện tử tự động vào Request Interceptor, hoàn thiện các luồng giao diện Đăng nhập, xác thực MFA và thiết lập mã QR Code trên ứng dụng Desktop.
- **Step 6 (Web Frontend Auth):** Triển khai luồng đăng nhập 2 giai đoạn (SĐT + OTP) với UI/UX mượt mà, hỗ trợ tự động điền và countdown. Cấu hình HTTPOnly Cookie cho Refresh Token an toàn chống XSS.

---

## Nền tảng Sprint 1 + Step 5 sử dụng

| Thành phần | Trạng thái |
|------------|-----------|
| `desktop/frontend/src/layouts/RoleLayout.tsx` | Sidebar dynamic theo role — **thêm Admin menu items** |
| `desktop/frontend/src/lib/apiClient.ts` | Đã hoàn thiện ở Step 5 (signature + refresh) — dùng ngay |
| Ant Design (`antd`) | Table, Modal, Form, Button, Badge, Checkbox — dùng ngay |
| `@tanstack/react-query` | `useQuery`, `useMutation`, `invalidateQueries` |

---

## 1. Admin Route Setup

- [x] Thêm Admin routes vào `desktop/frontend/src/App.tsx`:
  ```tsx
  <Route path="/admin" element={<ProtectedRoute roles={["admin"]} />}>
    <Route element={<RoleLayout />}>
      <Route path="dashboard"   element={<AdminDashboardPage />} />
      <Route path="users"       element={<UserListPage />} />
      <Route path="users/new"   element={<UserCreatePage />} />
      <Route path="roles"       element={<RolePermissionPage />} />
      <Route path="departments" element={<DepartmentPage />} />
    </Route>
  </Route>
  ```

- [x] Cập nhật `RoleLayout.tsx` — thêm Admin menu items:
  ```typescript
  admin: [
    { key: "/admin/dashboard",   label: "Dashboard",      icon: <DashboardOutlined /> },
    { key: "/admin/users",       label: "Quản lý User",   icon: <TeamOutlined /> },
    { key: "/admin/roles",       label: "Phân quyền",     icon: <SafetyOutlined /> },
    { key: "/admin/departments", label: "Khoa/Phòng ban", icon: <BankOutlined /> },
  ]
  ```

---

## 2. ProtectedRoute nâng cao (Desktop)

- [x] Cập nhật `ProtectedRoute.tsx` — hỗ trợ `roles` prop:
  ```typescript
  interface ProtectedRouteProps {
    roles?: string[];  // nếu truyền, check role trong claims
  }
  const ProtectedRoute = ({ roles }: ProtectedRouteProps) => {
    const { token, role } = useAuthStore();
    if (!token) return <Navigate to="/login" replace />;
    if (roles && !roles.includes(role ?? "")) return <Navigate to="/403" replace />;
    return <Outlet />;
  };
  ```

---

## 3. User List Page — `src/pages/admin/UserListPage.tsx`

- [x] `useQuery` gọi `GET /api/v1/users?page=1&limit=20&role=&search=`
- [x] Ant Design `<Table>` với columns:
  - Avatar (Ant Design `<Avatar>` với initials)
  - Username
  - Email (decrypt từ backend)
  - Roles (tags `<Badge>`)
  - Department
  - Status: `<Tag color="green">Hoạt động</Tag>` / `<Tag color="red">Đã khóa</Tag>`
  - Actions: Xem, Sửa role, Khóa tài khoản
- [x] Search bar: debounce 300ms → refetch với `search` param
- [x] Filter by Role: `<Select>` dropdown
- [x] Pagination: Ant Design `<Table pagination>` built-in

---

## 4. User Create Modal — `src/components/admin/UserCreateModal.tsx`

- [x] Ant Design `<Modal>` trigger từ nút "+ Tạo user mới"
- [x] `<Form>` fields:
  - Username (required, unique)
  - Email (required, email format)
  - Role (required, multi-select từ danh sách roles)
  - Department (select từ `GET /api/v1/departments`)
- [x] Submit: `useMutation` → `POST /api/v1/users` → `invalidateQueries(["users"])`
- [x] Success: thông báo "Đã tạo user. Email thông tin đăng nhập đã được gửi."
- [x] Error 409: "Username hoặc Email đã tồn tại."

---

## 5. Deactivate User

- [x] Nút "Khóa tài khoản" → Ant Design `<Popconfirm>`:
  ```
  "Bạn có chắc chắn muốn khóa tài khoản này không? User sẽ không thể đăng nhập."
  → Xác nhận: PUT /api/v1/users/:id/deactivate
  → invalidateQueries(["users"])
  ```

---

## 6. Assign Roles Modal — `src/components/admin/AssignRolesModal.tsx`

- [x] Ant Design `<Modal>` + `<Checkbox.Group>`:
  - Load danh sách roles từ `GET /api/v1/roles`
  - Pre-check roles hiện tại của user
- [x] Submit: `PUT /api/v1/users/:id/roles` với `{ role_ids: [...] }`

---

## 7. Role & Permission Matrix — `src/pages/admin/RolePermissionPage.tsx`

- [x] Load: `GET /api/v1/roles` (kèm permissions)
- [x] Ma trận bảng:
  - **Rows** = permissions (format: `resource:action`, ví dụ "patient:read")
  - **Columns** = roles (admin, doctor, nurse, receptionist, pharmacist)
  - **Cell** = `<Checkbox>` — checked nếu role có permission đó
- [x] Khi check/uncheck → lưu state local
- [x] Nút "Lưu thay đổi" → gọi `PUT /api/v1/roles/:id/permissions` cho từng role đã thay đổi
- [x] Debounce hoặc batch: không gọi API mỗi lần click, chỉ gọi khi nhấn Lưu

---

## 8. Department Page — `src/pages/admin/DepartmentPage.tsx`

- [x] `GET /api/v1/departments` → Table: Tên, Mã khoa, Số nhân viên
- [x] Modal tạo department mới: Tên (required), Mã khoa (required, unique)
- [x] `POST /api/v1/departments` → `invalidateQueries(["departments"])`

---

## 9. Admin Dashboard — `src/pages/admin/AdminDashboardPage.tsx`

- [x] Placeholder với stats cards (dùng Ant Design `<Card>` + `<Statistic>`):
  - Tổng số user
  - User đang hoạt động / đã khóa
  - Số khoa/phòng ban
- [x] Data từ `GET /api/v1/users?limit=1` (lấy `total` từ Meta)

---

## Definition of Done (Step 7)

- [x] Admin login → redirect `/admin/dashboard`
- [x] Non-admin login → không thấy admin menu, route `/admin/*` trả 403
- [x] `UserListPage` hiển thị danh sách, search, filter by role hoạt động
- [x] Tạo user mới thành công → xuất hiện trong danh sách
- [x] Deactivate user → status chuyển "Đã khóa", user không login được
- [x] Assign roles → cập nhật đúng, user nhận quyền mới sau khi đăng nhập lại
- [x] RolePermissionPage: check/uncheck checkbox → Lưu thay đổi → reload lại đúng
- [x] `wails dev` không console error trên Admin pages

---

## Kết thúc Sprint 2 — Tổng hợp Definition of Done

- [x] **Backend:** JWT AES-GCM issue/verify pass unit test ≥90% coverage
- [x] **Backend:** Challenge-response login flow hoạt động end-to-end
- [x] **Backend:** TOTP MFA setup + verify với Google Authenticator
- [x] **Backend:** OTP flow: gửi Zalo ZNS + SMS fallback, rate limit đúng
- [x] **Backend:** RBAC: route không có permission → 403
- [x] **Backend:** Signature middleware: request giả mạo → 401
- [x] **Backend:** Admin APIs CRUD user/role/department
- [x] **Desktop:** TPM/Keychain key pair sinh và ký request thành công
- [x] **Desktop:** Login challenge-response + MFA hoàn chỉnh
- [x] **Desktop:** Auto-refresh token với signature rotation
- [x] **Desktop:** Admin quản lý user, roles, permissions
- [x] **Web:** OTP login + register patient thành công
- [x] **Web:** Cookie refresh: token expire → silently refresh
- [x] **Web:** ProtectedRoute returnUrl hoạt động
