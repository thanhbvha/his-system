# Sprint 1 — Step 7: Desktop App (Wails + React)

> **Mục tiêu:** Khởi tạo ứng dụng desktop dùng Wails + React TypeScript, setup design system và routing cơ bản.
> **Phụ thuộc:** Step 5 (Backend `/health` sẵn sàng để test kết nối).
> **Output:** Desktop app build được, hiển thị layout cơ bản theo role, i18n VI/EN.

---

## 1. Khởi tạo Wails Project

- [x] Cài Wails CLI:
  ```bash
  go install github.com/wailsapp/wails/v2/cmd/wails@latest
  ```

- [x] Init project:
  ```bash
  wails init -n desktop -t react-ts
  ```
  Thư mục output: `desktop/`

- [x] Cài React dependencies:
  ```bash
  cd desktop/frontend
  npm install axios @tanstack/react-query zustand antd react-router-dom zod i18next react-i18next i18next-browser-languagedetector
  npm install -D @types/node
  ```

- [x] Setup path alias `@/` → `src/`:
  ```ts
  // vite.config.ts
  resolve: {
    alias: { "@": path.resolve(__dirname, "src") }
  }
  ```

---

## 2. Design System (Ant Design 5.x)

- [x] Ant Design theme config trong `src/main.tsx`:
  ```tsx
  import { ConfigProvider } from "antd";

  const theme = {
    token: {
      colorPrimary: "#1677ff",    // màu y tế (xanh dương)
      colorSuccess: "#52c41a",
      colorWarning: "#faad14",
      colorError:   "#ff4d4f",
      borderRadius: 6,
      fontFamily:   "Inter, Roboto, sans-serif",
    },
  };
  ```

- [x] Global CSS reset + custom variables: `src/styles/global.css`
  ```css
  :root {
    --color-primary: #1677ff;
    --color-bg: #f5f7fa;
    --sidebar-width: 240px;
  }
  * { box-sizing: border-box; }
  ```

- [x] Typography scale: `src/styles/typography.css`
  ```css
  h1 { font-size: 2rem; font-weight: 700; }
  h2 { font-size: 1.5rem; font-weight: 600; }
  h3 { font-size: 1.25rem; font-weight: 600; }
  h4 { font-size: 1rem; font-weight: 600; }
  ```

- [x] Import Google Fonts Inter trong `index.html`:
  ```html
  <link rel="preconnect" href="https://fonts.googleapis.com">
  <link href="https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700&display=swap" rel="stylesheet">
  ```

---

## 3. Core Setup

### API Client — `src/lib/apiClient.ts`

```ts
import axios from "axios";

const apiClient = axios.create({
  baseURL: "http://localhost:8080",  // from Wails env config
  timeout: 10_000,
});

// Attach Bearer token
apiClient.interceptors.request.use((config) => {
  const token = useAuthStore.getState().token;
  if (token) config.headers.Authorization = `Bearer ${token}`;
  return config;
});

// Auto-refresh interceptor (placeholder — implement Sprint 2)
apiClient.interceptors.response.use(
  (res) => res,
  async (err) => {
    // TODO: Sprint 2 — refresh token logic
    return Promise.reject(err);
  }
);

export default apiClient;
```

- [x] `src/lib/queryClient.ts`:
  ```ts
  import { QueryClient } from "@tanstack/react-query";

  export const queryClient = new QueryClient({
    defaultOptions: {
      queries: {
        staleTime: 5 * 60 * 1000,  // 5 phút
        retry: 2,
        refetchOnWindowFocus: false,
      },
    },
  });
  ```

- [x] `src/lib/websocket.ts` — WS client skeleton:
  ```ts
  class WSClient {
    private ws: WebSocket | null = null;
    connect(url: string): void { ... }
    disconnect(): void { ... }
    on(event: string, handler: (data: any) => void): void { ... }
    send(event: string, data: any): void { ... }
  }
  export const wsClient = new WSClient();
  ```

### State (Zustand)

- [x] `src/store/authStore.ts`:
  ```ts
  interface AuthState {
    token: string | null;
    user: User | null;
    role: "admin" | "doctor" | "nurse" | "receptionist" | "pharmacist" | null;
    setAuth: (token: string, user: User) => void;
    clearAuth: () => void;
  }
  ```

- [x] `src/store/uiStore.ts`:
  ```ts
  interface UIState {
    sidebarOpen: boolean;
    toggleSidebar: () => void;
    setSidebarOpen: (open: boolean) => void;
  }
  ```

---

## 4. Routing & Layout

- [x] React Router v6 setup trong `src/App.tsx`
- [x] `src/layouts/RoleLayout.tsx`:
  - Sidebar + Header + Content wrapper
  - Sidebar items render theo `role` từ `authStore`
  - Sidebar items per role:
    | Role | Menu items |
    |------|-----------|
    | `admin` | Tổng quan, Nhân sự, Cài đặt |
    | `doctor` | Bệnh nhân, Lịch khám, EMR |
    | `nurse` | Bệnh nhân, Chuẩn bị khám |
    | `receptionist` | Lịch hẹn, Tiếp nhận |
    | `pharmacist` | Kê đơn, Kho thuốc |

- [x] Route guard:
  ```tsx
  // src/components/ProtectedRoute.tsx
  const ProtectedRoute = () => {
    const token = useAuthStore((s) => s.token);
    return token ? <Outlet /> : <Navigate to="/login" replace />;
  };
  ```

- [x] Placeholder pages hiển thị "Coming soon" cho từng role

---

## 5. i18n

- [x] i18next config trong `src/i18n/index.ts`
- [x] `src/i18n/vi.json`:
  ```json
  {
    "common": {
      "loading": "Đang tải...",
      "error": "Đã xảy ra lỗi",
      "comingSoon": "Tính năng đang phát triển"
    },
    "nav": {
      "dashboard": "Tổng quan",
      "patients": "Bệnh nhân",
      "appointments": "Lịch hẹn"
    }
  }
  ```
- [x] `src/i18n/en.json` — bản tiếng Anh tương ứng
- [x] Language switcher component trong Header

---

## Definition of Done (Step 7)

- [x] `wails dev` khởi động không lỗi
- [x] Layout hiển thị đúng sidebar + header + content area
- [x] Sidebar items thay đổi theo role (test với mock role trong store)
- [x] Route guard redirect `/login` khi chưa có token
- [x] i18n: toggle VI/EN thay đổi text trên UI
- [x] Ant Design theme áp dụng đúng màu primary `#1677ff`
- [x] Fonts Inter load từ Google Fonts
