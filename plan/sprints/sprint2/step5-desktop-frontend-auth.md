# Sprint 2 — Step 5: Desktop Frontend — Hardware Key & Login Flow

> **Mục tiêu:** Hoàn thiện Desktop frontend với TPM/Keychain hardware key integration, request signature interceptor, và toàn bộ auth UI screens (Login, MFA, MFA Setup).
> **Phụ thuộc:** Step 2 backend API ready (`/auth/login/init`, `/auth/login/complete`, `/auth/mfa/*`).
> **Output:** Nhân viên đăng nhập Desktop thành công qua challenge-response + MFA; mọi API request được ký tự động.

---

## Các thành quả đã hoàn thành (Từ Step 1 đến Step 4)

- **Step 1 (Platform Infrastructure):** Thiết lập Platform Core (Fiber, Go-Common logger, Queue, WebSocket) và kết nối hạ tầng (PostgreSQL, MongoDB, Redis, MinIO).
- **Step 2 (Desktop Backend Auth):** Triển khai luồng đăng nhập Challenge-Response (`/auth/login/init`, `/auth/login/complete`), mã hoá payload bằng JWT + AES-GCM và MFA (TOTP).
- **Step 3 (Web Patient Auth):** Triển khai luồng xác thực bằng SĐT + OTP (SMS/Zalo) và đăng ký tài khoản bệnh nhân (`/auth/otp/send`, `/auth/otp/verify`, `/auth/register`).
- **Step 4 (RBAC & Admin API):** Hoàn tất Middleware kiểm soát truy cập (`JWTAuth`, `RequireRole`, `RequirePermission`, `RequestSignature`) cùng hệ thống API quản trị nhân viên và phòng ban.

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

- [x] `desktop/app.go` — Mở rộng struct App:
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

- [x] `desktop/internal/keystore/keystore.go`:
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

- [x] Implementations:
  - [x] `desktop/internal/keystore/windows_cng.go` — CNG API via `golang.org/x/sys/windows`, tạo key ECDSA-P256
  - [x] `desktop/internal/keystore/macos_keychain_cgo.go` — CGO macOS stub
  - [x] `desktop/internal/keystore/software_fallback.go` — lưu file `~/.his/device_key.pem`
  - [x] Build tag để chọn implementation: `//go:build windows`, `//go:build darwin`, `//go:build linux`

> ⚠️ **NOTE:** Private Key KHÔNG bao giờ rời khỏi hardware/file. `SignData` thực hiện ký trong hardware, chỉ trả về signature.

---

## 2. Token & Signature Interceptor — `apiClient.ts`

> ⚠️ **KHÔNG implement feature nào khác** cho đến khi interceptor này hoạt động đúng.

- [x] Hoàn thiện `desktop/frontend/src/lib/apiClient.ts`:

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

- [x] Cập nhật `authStore.ts` — thêm `refreshToken`:
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

- [x] Form: `username` + `password` (Ant Design Form, validation required)
- [x] Submit flow:
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
- [x] Error states:
  - `401` → "Tên đăng nhập hoặc mật khẩu không đúng"
  - `429` → "Quá nhiều lần thử. Vui lòng đợi X phút."
  - `423` → "Tài khoản bị khóa. Liên hệ quản trị viên."
- [x] Loading state trong khi ký (spinner + disable button)

---

## 4. MFA Screen — `src/pages/MFAPage.tsx`

- [x] Nhận `{ challenge_string, username }` từ navigation state
- [x] 6-ô OTP input (dùng OTP component, tương tự Web nhưng Ant Design style)
- [x] Submit flow:
  ```
  1. POST /auth/mfa/verify → { mfa_token }
  2. GetPublicKey() → publicKeyPem
  3. SignData(challenge_string) → signature
  4. POST /auth/login/complete (kèm mfa_token) → { access_token, refresh_token }
  5. setAuth() → navigate theo role
  ```
- [x] Nút "Dùng backup code" → input text field thay thế 6-ô

---

## 5. MFA Setup Screen — `src/pages/MFASetupPage.tsx`

> Hiển thị sau login lần đầu của Doctor/Admin nếu `user.mfa_enabled = false`.

- [x] `POST /auth/mfa/setup` → nhận `{ qr_uri, backup_codes }`
- [x] Hiển thị QR code (dùng `npm install qrcode.react`)
- [x] Hướng dẫn từng bước:
  - Bước 1: Cài app Google Authenticator / Authy
  - Bước 2: Quét QR code
  - Bước 3: Nhập mã 6 số để xác nhận kích hoạt
- [x] Sau xác nhận → hiển thị 8 backup codes, nút "Download" + "Copy"
- [x] Warning: "Lưu backup codes ở nơi an toàn. Chúng sẽ không hiển thị lại."

---

## 6. Role-based Redirect

- [x] Sau login thành công, đọc `claims.roles[0]` và navigate:
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
- [x] Tạo placeholder pages cho từng route (hiển thị role name + "Coming in Sprint 3+")

---

## 7. i18n keys — Desktop

- [x] Bổ sung keys vào `vi.json` và `en.json`:
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

- [x] `GetPublicKey()` và `SignData()` expose qua Wails và gọi được từ React
- [x] Windows: key pair sinh từ CNG (hoặc software fallback nếu không có TPM)
- [x] Login form submit → challenge-response hoàn chỉnh → JWT nhận được
- [x] `mfa_required: true` → redirect đúng MFA screen, flow tiếp tục thành công
- [x] MFA Setup screen hiển thị QR, sau verify → backup codes hiện ra
- [x] Auto-refresh: khi token hết hạn, request tự retry sau khi refresh
- [x] Concurrent requests trong lúc refresh: queue đúng, không call refresh 2 lần
- [x] Role-based redirect: login với `doctor` → `/doctor/worklist`
- [x] `wails dev` không có lỗi console
