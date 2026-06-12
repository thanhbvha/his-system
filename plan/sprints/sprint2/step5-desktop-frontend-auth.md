# Sprint 2 — Step 5: Desktop Frontend — Hardware Key & Login Flow

> **Mục tiêu:** Hoàn thiện Desktop frontend với TPM/Keychain hardware key integration, request signature interceptor, và toàn bộ auth UI screens (Login, MFA, MFA Setup).
> **Phụ thuộc:** Step 2 backend API ready (`/auth/login/init`, `/auth/login/complete`, `/auth/mfa/*`).
> **Output:** Nhân viên đăng nhập Desktop thành công qua challenge-response + MFA; mọi API request được ký tự động.

---

## Nền tảng Sprint 1 sử dụng

| File | Trạng thái | Công việc cần làm |
|------|-----------|------------------|
| `desktop/frontend/src/lib/apiClient.ts` | Stub — Bearer interceptor | **Hoàn thiện:** thêm signature + refresh logic |
| `desktop/frontend/src/store/authStore.ts` | Có `token`, `user`, `role` | **Mở rộng:** thêm `publicKeyHash` |
| `desktop/frontend/src/layouts/RoleLayout.tsx` | Dynamic sidebar theo role | Dùng ngay sau login thành công |
| `desktop/frontend/src/components/ProtectedRoute.tsx` | Redirect `/login` | Dùng ngay |
| `desktop/frontend/src/i18n/vi.json`, `en.json` | Keys cơ bản | **Bổ sung** keys cho auth screens |
| `desktop/app.go` | Wails app struct | **Thêm** `GetPublicKey()`, `SignData()` |

---

## 1. Wails Go Backend — Hardware Key (TPM/Keychain)

- [ ] `desktop/app.go` — Mở rộng struct App:
  ```go
  type App struct {
      ctx context.Context
  }

  // GetPublicKey đọc hoặc generate key pair từ OS hardware keystore.
  // Windows: Windows CNG + TPM
  // macOS:   Keychain (Secure Enclave nếu có)
  // Linux:   TPM 2.0 hoặc software fallback (file-based, warn user)
  // Trả về PEM của Public Key.
  func (a *App) GetPublicKey() (string, error)

  // SignData ký `data` bằng Private Key trong hardware store.
  // Trả về base64(signature).
  func (a *App) SignData(data string) (string, error)
  ```

- [ ] `desktop/internal/keystore/keystore.go`:
  ```go
  type KeyStore interface {
      GetOrCreate() (*KeyPair, error)
      Sign(data []byte) ([]byte, error)
  }

  type KeyPair struct {
      PublicKeyPEM string
      // Algorithm luôn là ECDSA-P256 (secp256r1) — không có field Algorithm vì:
      // Windows CNG/TPM và macOS Secure Enclave chỉ hỗ trợ ECDSA-P256 native.
      // Chọn thuật toán khác sẽ buộc phải dùng software fallback,
      // vô hiệu hóa toàn bộ mục tiêu hardware binding.
  }
  ```

- [ ] Implementations:
  - [ ] `desktop/internal/keystore/windows_tpm.go` — CNG API via `golang.org/x/sys/windows`, tạo key ECDSA-P256 trọn TPM
  - [ ] `desktop/internal/keystore/macos_keychain.go` — Secure Enclave/Keychain via cgo, ECDSA-P256 (thuật toán duy nhất Secure Enclave hỗ trợ)
  - [ ] `desktop/internal/keystore/software_fallback.go` — tạo ECDSA-P256 key pair, lưu file `~/.his/key.pem` (warn user: không có hardware protection)
  - [ ] Build tag để chọn implementation: `//go:build windows`, `//go:build darwin`, `//go:build linux`

> ⚠️ **NOTE:** Private Key KHÔNG bao giờ rời khỏi hardware/file. `SignData` thực hiện ký trong hardware, chỉ trả về signature.

---

## 2. Token & Signature Interceptor — `apiClient.ts`

> ⚠️ **KHÔNG implement feature nào khác** cho đến khi interceptor này hoạt động đúng.

- [ ] Hoàn thiện `desktop/frontend/src/lib/apiClient.ts`:

```typescript
import axios from "axios";
import { useAuthStore } from "@/store/authStore";
import { GetPublicKey, SignData } from "../wailsjs/go/main/App";

const api = axios.create({ baseURL: "/api/v1" });

// === Request Interceptor: Attach Bearer + Hardware Signature ===
api.interceptors.request.use(async (config) => {
  const token = useAuthStore.getState().token;
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }

  // Ký request cho các endpoint yêu cầu signature
  if (config.headers["X-Require-Signature"] !== "false") {
    const timestamp = Date.now().toString();
    const body = config.data ? JSON.stringify(config.data) : "";
    const message = `${config.method?.toUpperCase()}${config.url}${timestamp}${body}`;
    try {
      const signature = await SignData(message);
      config.headers["X-Timestamp"] = timestamp;
      config.headers["X-Signature"] = signature;
    } catch {
      // Hardware không khả dụng — request vẫn tiếp tục nhưng sẽ bị server reject
    }
  }
  return config;
});

// === Response Interceptor: Auto-refresh Token ===
let isRefreshing = false;
let pendingQueue: Array<{ resolve: (token: string) => void; reject: (err: unknown) => void }> = [];

api.interceptors.response.use(
  (res) => res,
  async (err) => {
    const originalRequest = err.config;
    if (err.response?.status === 401 && !originalRequest._retry) {
      if (isRefreshing) {
        // Queue request pending trong khi đang refresh
        return new Promise((resolve, reject) => {
          pendingQueue.push({ resolve, reject });
        }).then((token) => {
          originalRequest.headers.Authorization = `Bearer ${token}`;
          return api(originalRequest);
        });
      }
      originalRequest._retry = true;
      isRefreshing = true;
      try {
        const publicKeyPem = await GetPublicKey();
        const refreshToken = useAuthStore.getState().refreshToken;
        const signature = await SignData(refreshToken);
        const res = await api.post("/auth/refresh", { refresh_token: refreshToken, signature, public_key_pem: publicKeyPem }, { headers: { "X-Require-Signature": "false" } });
        const newToken = res.data.data.access_token;
        useAuthStore.getState().setToken(newToken);
        pendingQueue.forEach(p => p.resolve(newToken));
        pendingQueue = [];
        originalRequest.headers.Authorization = `Bearer ${newToken}`;
        return api(originalRequest);
      } catch {
        pendingQueue.forEach(p => p.reject(new Error("Session expired")));
        pendingQueue = [];
        useAuthStore.getState().clearAuth();
        window.location.hash = "/login";
      } finally {
        isRefreshing = false;
      }
    }
    return Promise.reject(err);
  }
);

export default api;
```

- [ ] Cập nhật `authStore.ts` — thêm `refreshToken`:
  ```typescript
  interface AuthState {
    token: string | null;
    refreshToken: string | null;  // ← Thêm mới
    user: User | null;
    role: string | null;
    setAuth: (token: string, refreshToken: string, user: User) => void;
    setToken: (token: string) => void;  // ← Thêm mới (dùng sau refresh)
    clearAuth: () => void;
  }
  ```

---

## 3. Login Screen — `src/pages/LoginPage.tsx`

- [ ] Form: `username` + `password` (Ant Design Form, validation required)
- [ ] Submit flow:
  ```
  1. POST /auth/login/init → { challenge_string, mfa_required }
  2a. mfa_required = true → navigate("/mfa", { state: { challenge_string, username } })
  2b. mfa_required = false:
      - GetPublicKey() → publicKeyPem
      - SignData(challenge_string) → signature
      - POST /auth/login/complete → { access_token, refresh_token }
      - setAuth(access_token, refresh_token, user)
      - navigate theo role
  ```
- [ ] Error states:
  - `401` → "Tên đăng nhập hoặc mật khẩu không đúng"
  - `429` → "Quá nhiều lần thử. Vui lòng đợi X phút."
  - `423` → "Tài khoản bị khóa. Liên hệ quản trị viên."
- [ ] Loading state trong khi ký (spinner + disable button)

---

## 4. MFA Screen — `src/pages/MFAPage.tsx`

- [ ] Nhận `{ challenge_string, username }` từ navigation state
- [ ] 6-ô OTP input (dùng OTP component, tương tự Web nhưng Ant Design style)
- [ ] Submit flow:
  ```
  1. POST /auth/mfa/verify → { mfa_token }
  2. GetPublicKey() → publicKeyPem
  3. SignData(challenge_string) → signature
  4. POST /auth/login/complete (kèm mfa_token) → { access_token, refresh_token }
  5. setAuth() → navigate theo role
  ```
- [ ] Nút "Dùng backup code" → input text field thay thế 6-ô

---

## 5. MFA Setup Screen — `src/pages/MFASetupPage.tsx`

> Hiển thị sau login lần đầu của Doctor/Admin nếu `user.mfa_enabled = false`.

- [ ] `POST /auth/mfa/setup` → nhận `{ qr_uri, backup_codes }`
- [ ] Hiển thị QR code (dùng `npm install qrcode.react`)
- [ ] Hướng dẫn từng bước:
  - Bước 1: Cài app Google Authenticator / Authy
  - Bước 2: Quét QR code
  - Bước 3: Nhập mã 6 số để xác nhận kích hoạt
- [ ] Sau xác nhận → hiển thị 8 backup codes, nút "Download" + "Copy"
- [ ] Warning: "Lưu backup codes ở nơi an toàn. Chúng sẽ không hiển thị lại."

---

## 6. Role-based Redirect

- [ ] Sau login thành công, đọc `claims.roles[0]` và navigate:
  ```typescript
  const roleRouteMap: Record<string, string> = {
    receptionist: "/receptionist/queue",
    doctor:       "/doctor/worklist",
    lab_tech:     "/lab/worklist",
    pharmacist:   "/pharmacy/prescriptions",
    admin:        "/admin/dashboard",
  };
  navigate(roleRouteMap[role] ?? "/dashboard");
  ```
- [ ] Tạo placeholder pages cho từng route (hiển thị role name + "Coming in Sprint 3+")

---

## 7. i18n keys — Desktop

- [ ] Bổ sung keys vào `vi.json` và `en.json`:
  ```json
  {
    "auth": {
      "loginTitle": "Đăng nhập hệ thống",
      "username": "Tên đăng nhập",
      "password": "Mật khẩu",
      "loginBtn": "Đăng nhập",
      "mfaTitle": "Xác thực 2 bước",
      "mfaSetupTitle": "Thiết lập xác thực 2 bước",
      "backupCodes": "Mã dự phòng",
      "errors": {
        "invalidCredentials": "Tên đăng nhập hoặc mật khẩu không đúng",
        "tooManyAttempts": "Quá nhiều lần thử. Vui lòng đợi {{minutes}} phút.",
        "accountLocked": "Tài khoản bị khóa. Liên hệ quản trị viên."
      }
    }
  }
  ```

---

## Definition of Done (Step 5)

- [ ] `GetPublicKey()` và `SignData()` expose qua Wails và gọi được từ React
- [ ] Windows: key pair sinh từ CNG (hoặc software fallback nếu không có TPM)
- [ ] Login form submit → challenge-response hoàn chỉnh → JWT nhận được
- [ ] `mfa_required: true` → redirect đúng MFA screen, flow tiếp tục thành công
- [ ] MFA Setup screen hiển thị QR, sau verify → backup codes hiện ra
- [ ] Auto-refresh: khi token hết hạn, request tự retry sau khi refresh
- [ ] Concurrent requests trong lúc refresh: queue đúng, không call refresh 2 lần
- [ ] Role-based redirect: login với `doctor` → `/doctor/worklist`
- [ ] `wails dev` không có lỗi console
