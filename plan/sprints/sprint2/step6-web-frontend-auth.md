# Sprint 2 — Step 6: Web Frontend — OTP Login & Register Flow

> **Mục tiêu:** Thay thế LoginPage stub bằng OTP flow thực, implement RegisterPage, hoàn thiện apiClient với cookie refresh, và xây dựng OTP Input component.
> **Phụ thuộc:** Step 3 backend API ready (`/auth/otp/send`, `/auth/otp/verify`, `/auth/register`).
> **Output:** Bệnh nhân đăng nhập và đăng ký thành công qua OTP trên Web, auto-refresh token hoạt động.

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

- [ ] Cập nhật `web/src/lib/apiClient.ts`:
  - `withCredentials: true` — browser tự đính kèm HttpOnly cookie
  - Response 401 interceptor: gọi `POST /auth/refresh` (cookie tự đính)
  - Nếu refresh OK → `setToken(newToken)` → retry request gốc
  - Nếu refresh fail → `clearAuth()` → redirect `/login`
  - Xử lý concurrent: queue pending requests trong khi đang refresh
- [ ] Cập nhật `authStore.ts` — thêm `setToken: (token: string) => void`

---

## 2. shadcn/ui Components bổ sung

- [ ] `npx shadcn@latest add input label form card badge`
- [ ] Verify `react-hook-form` + `zod` + `@hookform/resolvers` đã có trong `node_modules`

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

- [ ] 6 ô input riêng biệt, chỉ nhận ký tự số
- [ ] Auto-focus next ô sau khi điền 1 ký tự
- [ ] Backspace: xoá ô hiện tại → focus ô trước
- [ ] Paste: paste "123456" → điền toàn bộ 6 ô ngay lập tức
- [ ] Error state: viền đỏ toàn bộ khi `error=true`

---

## 4. Login Page — `src/pages/LoginPage.tsx` (Replace hoàn toàn)

State machine: `"enter_phone" | "enter_otp" | "loading"`

**Phase 1 — Nhập SĐT:**
- [ ] Input SĐT + nút "Gửi mã OTP"
- [ ] Validation: 10 số, bắt đầu `0`
- [ ] `POST /auth/otp/send` → chuyển Phase 2
- [ ] Error 429: "Đã gửi OTP quá nhiều lần. Thử lại sau X phút."

**Phase 2 — Nhập OTP:**
- [ ] `<OTPInput onComplete={...} error={hasError} />`
- [ ] Countdown 60s + nút "Gửi lại" (disabled trong countdown)
- [ ] `POST /auth/otp/verify`:
  - `needs_register: true` → `navigate("/register", { state: { phone } })`
  - success → `setAuth(token, patient)` → `navigate("/my-appointments")`
  - `401` → "Mã OTP không đúng", error state OTPInput
  - `410` → "Mã OTP hết hạn", bật nút Gửi lại ngay
  - `429` → "Quá nhiều lần sai, yêu cầu mã mới"

---

## 5. Register Page — `src/pages/RegisterPage.tsx` (Implement)

- [ ] Guard: không có `phone` từ location state → redirect `/login`
- [ ] Form với `react-hook-form` + `zod`:
  ```typescript
  z.object({
    full_name: z.string().min(2),
    dob: z.string().regex(/^\d{4}-\d{2}-\d{2}$/),
    gender: z.enum(["male", "female", "other"]),
    email: z.string().email().optional().or(z.literal("")),
  })
  ```
- [ ] `POST /auth/register` → `setAuth(token, patient)` → `navigate("/my-appointments")`
- [ ] Error 409: "SĐT đã đăng ký. Vui lòng đăng nhập."

---

## 6. Enhanced ProtectedRoute

- [ ] Lưu `returnUrl = location.pathname` vào navigate state khi redirect `/login`
- [ ] Sau login/register thành công → check `location.state?.returnUrl` → redirect về đó

---

## 7. i18n Keys bổ sung

- [ ] `vi.json` + `en.json`: thêm namespace `"auth"` với keys:
  - `loginTitle`, `enterPhone`, `sendOTP`, `enterOTP`, `resend`, `resendIn`
  - `registerTitle`, `fullName`, `dob`, `gender`, `createAccount`
  - `errors.*` cho mọi error case

---

## Definition of Done (Step 6)

- [ ] OTP 6 ô: paste, auto-focus, backspace đều đúng
- [ ] Countdown 60s chạy, nút Gửi lại bật khi hết giờ
- [ ] `needs_register: true` → chuyển `/register` với phone state
- [ ] RegisterPage: validation, tạo account, redirect thành công
- [ ] ProtectedRoute: lưu và khôi phục returnUrl
- [ ] Auto-refresh cookie: token hết hạn → retry silently
- [ ] `npm run build` không lỗi
