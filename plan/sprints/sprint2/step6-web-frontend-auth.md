# Sprint 2 — Step 6: Web Frontend — OTP Login & Register Flow

> **Mục tiêu:** Thay thế LoginPage stub bằng OTP flow thực, implement RegisterPage, hoàn thiện apiClient với cookie refresh, và xây dựng OTP Input component.
> **Phụ thuộc:** Step 3 backend API ready (`/auth/otp/send`, `/auth/otp/verify`, `/auth/register`).
> **Output:** Bệnh nhân đăng nhập và đăng ký thành công qua OTP trên Web, auto-refresh token hoạt động.

---

## Các thành quả đã hoàn thành (Từ Step 1 đến Step 5)

- **Step 1 (Platform Infrastructure):** Thiết lập Platform Core (Fiber, Go-Common logger, Queue, WebSocket) và kết nối hạ tầng (PostgreSQL, MongoDB, Redis, MinIO).
- **Step 2 (Desktop Backend Auth):** Triển khai luồng đăng nhập Challenge-Response (`/auth/login/init`, `/auth/login/complete`), mã hoá payload bằng JWT + AES-GCM và MFA (TOTP).
- **Step 3 (Web Patient Auth):** Triển khai luồng xác thực bằng SĐT + OTP (SMS/Zalo) và đăng ký tài khoản bệnh nhân (`/auth/otp/send`, `/auth/otp/verify`, `/auth/register`).
- **Step 4 (RBAC & Admin API):** Hoàn tất Middleware kiểm soát truy cập (`JWTAuth`, `RequireRole`, `RequirePermission`, `RequestSignature`) cùng hệ thống API quản trị nhân viên và phòng ban.
- **Step 5 (Desktop Frontend Auth):** Tích hợp Native Hardware Keystore (Windows CNG TPM & macOS CGO Keychain) với Wails. Gắn chữ ký điện tử tự động vào Request Interceptor, hoàn thiện các luồng giao diện Đăng nhập, xác thực MFA và thiết lập mã QR Code trên ứng dụng Desktop.

---

## Nền tảng Sprint 1 sử dụng

| File | Trạng thái | Công việc cần làm |
|------|-----------|------------------|
| `web/src/lib/apiClient.ts` | Stub — Bearer + clearAuth on 401 | **Hoàn thiện:** cookie refresh logic |
| `web/src/store/authStore.ts` | `token`, `patient`, `setAuth`, `clearAuth` | Thêm `setToken` |
| `web/src/pages/LoginPage.tsx` | Mock login stub | **Replace** bằng OTP flow |
| `web/src/pages/RegisterPage.tsx` | "Coming soon" | **Implement** |
| `web/src/components/ProtectedRoute.tsx` | Redirect `/login` | **Mở rộng:** returnUrl + attempt refresh |
| `web/src/i18n/vi.json`, `en.json` | Keys cơ bản | **Bổ sung** keys auth |

---

## 1. Hoàn thiện `apiClient.ts` — Cookie Refresh

- [x] Cập nhật `web/src/lib/apiClient.ts`:
  - `withCredentials: true` — browser tự đính kèm HttpOnly cookie
  - Response 401 interceptor: gọi `POST /auth/refresh` (cookie tự đính)
  - Nếu refresh OK → `setToken(newToken)` → retry request gốc
  - Nếu refresh fail → `clearAuth()` → redirect `/login`
  - Xử lý concurrent: queue pending requests trong khi đang refresh
- [x] Cập nhật `authStore.ts` — thêm `setToken: (token: string) => void`

---

## 2. shadcn/ui Components bổ sung

- [x] `npx shadcn@latest add input label form card badge`
- [x] Verify `react-hook-form` + `zod` + `@hookform/resolvers` đã có trong `node_modules`

---

## 3. OTP Input Component — `src/components/OTPInput.tsx`

```typescript
interface OTPInputProps {
  length?: number;        // default 6
  onChange: (otp: string) => void;
  onComplete?: (otp: string) => void;
  disabled?: boolean;
  error?: boolean;
}
```

- [x] 6 ô input riêng biệt, chỉ nhận ký tự số
- [x] Auto-focus next ô sau khi điền 1 ký tự
- [x] Backspace: xoá ô hiện tại → focus ô trước
- [x] Paste: paste "123456" → điền toàn bộ 6 ô ngay lập tức
- [x] Error state: viền đỏ toàn bộ khi `error=true`

---

## 4. Login Page — `src/pages/LoginPage.tsx` (Replace hoàn toàn)

State machine: `"enter_phone" | "enter_otp" | "loading"`

**Phase 1 — Nhập SĐT:**
- [x] Input SĐT + nút "Gửi mã OTP"
- [x] Validation: 10 số, bắt đầu `0`
- [x] `POST /auth/otp/send` → chuyển Phase 2
- [x] Error 429: "Đã gửi OTP quá nhiều lần. Thử lại sau X phút."

**Phase 2 — Nhập OTP:**
- [x] `<OTPInput onComplete={...} error={hasError} />`
- [x] Countdown 60s + nút "Gửi lại" (disabled trong countdown)
- [x] `POST /auth/otp/verify`:
  - `needs_register: true` → `navigate("/register", { state: { phone } })`
  - success → `setAuth(token, patient)` → `navigate("/my-appointments")`
  - `401` → "Mã OTP không đúng", error state OTPInput
  - `410` → "Mã OTP hết hạn", bật nút Gửi lại ngay
  - `429` → "Quá nhiều lần sai, yêu cầu mã mới"

---

## 5. Register Page — `src/pages/RegisterPage.tsx` (Implement)

- [x] Guard: không có `phone` từ location state → redirect `/login`
- [x] Form với `react-hook-form` + `zod`:
  ```typescript
  z.object({
    full_name: z.string().min(2),
    dob: z.string().regex(/^\d{4}-\d{2}-\d{2}$/),
    gender: z.enum(["male", "female", "other"]),
    email: z.string().email().optional().or(z.literal("")),
  })
  ```
- [x] `POST /auth/register` → `setAuth(token, patient)` → `navigate("/my-appointments")`
- [x] Error 409: "SĐT đã đăng ký. Vui lòng đăng nhập."

---

## 6. Enhanced ProtectedRoute

- [x] Lưu `returnUrl = location.pathname` vào navigate state khi redirect `/login`
- [x] Sau login/register thành công → check `location.state?.returnUrl` → redirect về đó

---

## 7. i18n Keys bổ sung

- [x] `vi.json` + `en.json`: thêm namespace `"auth"` với keys:
  - `loginTitle`, `enterPhone`, `sendOTP`, `enterOTP`, `resend`, `resendIn`
  - `registerTitle`, `fullName`, `dob`, `gender`, `createAccount`
  - `errors.*` cho mọi error case

---

## Definition of Done (Step 6)

- [x] OTP 6 ô: paste, auto-focus, backspace đều đúng
- [x] Countdown 60s chạy, nút Gửi lại bật khi hết giờ
- [x] `needs_register: true` → chuyển `/register` với phone state
- [x] RegisterPage: validation, tạo account, redirect thành công
- [x] ProtectedRoute: lưu và khôi phục returnUrl
- [x] Auto-refresh cookie: token hết hạn → retry silently
- [x] `npm run build` không lỗi
