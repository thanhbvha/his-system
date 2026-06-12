# Sprint 1 — Step 8: Web App (React + Vite)

> **Mục tiêu:** Khởi tạo patient-facing web app bằng React + Vite, setup design system shadcn/ui và routing cơ bản.
> **Phụ thuộc:** Step 5 (Backend `/health` để test proxy).
> **Output:** Vite dev server chạy, landing page placeholder hiển thị, i18n VI mặc định.

---

## 1. Khởi tạo Vite Project

- [x] Tạo project:
  ```bash
  npm create vite@latest web -- --template react-ts
  cd web
  npm install
  ```
  Thư mục output: `web/`

- [x] Cài dependencies:
  ```bash
  npm install \
    axios \
    @tanstack/react-query \
    zustand \
    react-hook-form \
    zod \
    @hookform/resolvers \
    react-router-dom \
    i18next \
    react-i18next \
    i18next-browser-languagedetector
  ```

- [x] Cài shadcn/ui + Tailwind CSS:
  ```bash
  npm install tailwindcss @tailwindcss/vite
  npx shadcn@latest init
  ```

---

## 2. Design System (shadcn/ui + Tailwind)

- [x] `tailwind.config.ts` — màu primary y tế:
  ```ts
  extend: {
    colors: {
      primary: {
        50:  "#eff6ff",
        500: "#3b82f6",  // blue-500
        600: "#2563eb",  // blue-600 (màu chính)
        700: "#1d4ed8",
      },
    },
    fontFamily: {
      sans: ["Inter", "sans-serif"],
    },
  }
  ```

- [x] shadcn/ui `components.json` config:
  ```json
  {
    "style": "default",
    "baseColor": "blue",
    "cssVariables": true
  }
  ```

- [x] `src/styles/globals.css`:
  ```css
  @import url('https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700&display=swap');
  @tailwind base;
  @tailwind components;
  @tailwind utilities;

  :root {
    --color-primary: #2563eb;
    --color-primary-hover: #1d4ed8;
  }
  ```

- [x] Responsive breakpoints sử dụng Tailwind defaults (sm, md, lg, xl)

---

## 3. Core Setup

### API Client — `src/lib/apiClient.ts`

```ts
import axios from "axios";

const apiClient = axios.create({
  baseURL: import.meta.env.VITE_API_URL,
  withCredentials: true,  // cho cookie-based refresh token
  timeout: 10_000,
});

// Attach Bearer token từ store
apiClient.interceptors.request.use((config) => {
  const token = useAuthStore.getState().token;
  if (token) config.headers.Authorization = `Bearer ${token}`;
  return config;
});

// Auto-refresh interceptor (placeholder — implement Sprint 2)
apiClient.interceptors.response.use(
  (res) => res,
  async (err) => {
    // TODO: Sprint 2 — cookie refresh token logic
    return Promise.reject(err);
  }
);

export default apiClient;
```

- [x] `src/lib/queryClient.ts` — TanStack Query config
- [x] `src/store/authStore.ts`:
  ```ts
  interface AuthState {
    token: string | null;
    patient: Patient | null;
    setAuth: (token: string, patient: Patient) => void;
    clearAuth: () => void;
  }
  ```
- [x] `src/store/bookingStore.ts` — multi-step booking state:
  ```ts
  interface BookingState {
    step: 1 | 2 | 3 | 4;
    selectedDepartment: string | null;
    selectedDoctor: string | null;
    selectedSlot: Slot | null;
    patientInfo: PatientInfo | null;
    setStep: (step: number) => void;
    reset: () => void;
  }
  ```

---

## 4. Routing & Layout

- [x] React Router v6 setup trong `src/App.tsx`:
  ```tsx
  <Routes>
    {/* Public routes */}
    <Route element={<PublicLayout />}>
      <Route path="/"           element={<LandingPage />} />
      <Route path="/login"      element={<LoginPage />} />
      <Route path="/register"   element={<RegisterPage />} />
    </Route>

    {/* Protected routes */}
    <Route element={<ProtectedRoute />}>
      <Route element={<AuthLayout />}>
        <Route path="/book"             element={<BookingPage />} />
        <Route path="/my-appointments"  element={<MyAppointmentsPage />} />
        <Route path="/results"          element={<ResultsPage />} />
        <Route path="/account"          element={<AccountPage />} />
      </Route>
    </Route>
  </Routes>
  ```

- [x] `src/layouts/PublicLayout.tsx`:
  - Header: Logo + Nav links + Login button
  - Footer: Copyright, links

- [x] `src/layouts/AuthLayout.tsx`:
  - Header: Logo + Nav links + User avatar dropdown (logout)
  - Footer

- [x] Protected route wrapper:
  ```tsx
  const ProtectedRoute = () => {
    const token = useAuthStore((s) => s.token);
    return token ? <Outlet /> : <Navigate to="/login" replace />;
  };
  ```

- [x] Placeholder pages (hiển thị content cơ bản + "Coming soon"):
  | Route | Page | Nội dung |
  |-------|------|---------|
  | `/` | LandingPage | Hero section, giới thiệu dịch vụ |
  | `/login` | LoginPage | Form đăng nhập (UI, chưa có API) |
  | `/register` | RegisterPage | Form đăng ký bệnh nhân |
  | `/book` | BookingPage | Stepper booking 4 bước |
  | `/my-appointments` | MyAppointmentsPage | Danh sách lịch hẹn |
  | `/results` | ResultsPage | Kết quả xét nghiệm |
  | `/account` | AccountPage | Thông tin cá nhân |

---

## 5. i18n

- [x] `src/i18n/index.ts`:
  ```ts
  i18n
    .use(LanguageDetector)
    .use(initReactI18next)
    .init({
      lng: "vi",        // mặc định VI
      fallbackLng: "vi",
      resources: {
        vi: { translation: viTranslation },
        en: { translation: enTranslation },
      },
    });
  ```
- [x] `src/i18n/vi.json` — nội dung VI đầy đủ cho các trang landing
- [x] `src/i18n/en.json` — bản EN tương ứng

---

## 6. Build & Env Config

- [x] `.env.development`:
  ```env
  VITE_API_URL=http://localhost:8080
  ```

- [x] `.env.production`:
  ```env
  VITE_API_URL=https://api.his-system.vn
  ```

- [x] `vite.config.ts` — proxy cho dev mode:
  ```ts
  server: {
    proxy: {
      "/api": {
        target: "http://localhost:8080",
        changeOrigin: true,
      },
    },
  }
  ```

---

## Definition of Done (Step 8)

- [x] `npm run dev` khởi động Vite dev server tại `localhost:5173`
- [x] Landing page `/` load không lỗi, hiển thị hero section
- [x] Route `/login` render form (UI only, chưa cần API)
- [x] Route guard: `/book` redirect về `/login` khi chưa auth
- [x] shadcn/ui theme áp dụng đúng màu primary blue-600
- [x] Font Inter load từ Google Fonts
- [x] i18n mặc định VI, text hiển thị đúng tiếng Việt
- [x] Vite proxy `/api` → `localhost:8080` hoạt động (test với `/api/health`)
