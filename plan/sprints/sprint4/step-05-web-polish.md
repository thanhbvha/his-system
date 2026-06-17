# Sprint 4 — Step 5: Web App Polish — UX, Responsive, SEO, Error Handling

## Mục tiêu
Hoàn thiện Web App bệnh nhân lên trạng thái production-ready: skeleton loading, error boundary, responsive mobile, SEO metadata, và thay thế tất cả `window.alert` bằng toast notifications.

## Prerequisite
- Sprint 3 ✅: Booking Flow 4 bước, My Appointments, Landing Page đã hoạt động cơ bản.
- `sonner` hoặc `shadcn/ui` Toast đã có trong project (kiểm tra `package.json`).

## Files cần tạo / cập nhật

```
web/src/
├── components/
│   ├── ui/
│   │   ├── skeleton.tsx          -- Skeleton component (nếu chưa có)
│   │   └── toast.tsx             -- Toast (kiểm tra đã có chưa)
│   ├── layout/
│   │   └── ErrorBoundary.tsx     -- React Error Boundary wrapper
│   └── shared/
│       ├── PageSkeleton.tsx      -- Skeleton layout dùng chung
│       └── EmptyState.tsx        -- Empty state component dùng chung
├── pages/
│   ├── LandingPage.tsx           -- Thêm skeleton + SEO meta
│   ├── BookingPage.tsx           -- Thêm skeleton + disable khi submit
│   ├── MyAppointmentsPage.tsx    -- Thêm empty state + toast
│   └── AccountPage.tsx          -- Responsive check
├── main.tsx                      -- Wrap app với ErrorBoundary + Toaster
└── index.css                     -- Media queries cho mobile (nếu cần)
```

## Nhiệm vụ chi tiết

### 1. Error Boundary — `ErrorBoundary.tsx`

```tsx
class ErrorBoundary extends React.Component<
  { children: ReactNode; fallback?: ReactNode },
  { hasError: boolean; error?: Error }
> {
  state = { hasError: false };

  static getDerivedStateFromError(error: Error) {
    return { hasError: true, error };
  }

  componentDidCatch(error: Error, info: React.ErrorInfo) {
    console.error('ErrorBoundary caught:', error, info);
    // Optional: report to Sentry / logging
  }

  render() {
    if (this.state.hasError) {
      return this.props.fallback ?? (
        <div className="flex flex-col items-center justify-center min-h-[400px] gap-4">
          <h2 className="text-xl font-semibold text-destructive">Đã xảy ra lỗi</h2>
          <p className="text-slate-500">Vui lòng tải lại trang hoặc thử lại sau.</p>
          <Button onClick={() => this.setState({ hasError: false })}>Thử lại</Button>
        </div>
      );
    }
    return this.props.children;
  }
}
```

- Wrap `<App />` trong `main.tsx` với `<ErrorBoundary>`
- Wrap từng route page với `<ErrorBoundary>` riêng (fallback nhẹ hơn)

### 2. Toast Notifications (thay `window.alert`)

Kiểm tra `package.json` xem đã có `sonner` chưa:
- Nếu chưa: `npm install sonner`
- Trong `main.tsx`: thêm `<Toaster />` từ `sonner`

**Thay thế trong `MyAppointmentsPage.tsx`:**
```tsx
// Trước:
alert("Hủy lịch thất bại");
// Sau:
import { toast } from 'sonner';
toast.error("Hủy lịch thất bại. Vui lòng thử lại.");
```

**Thay thế trong `StepConfirm.tsx`:**
- Error 409: `toast.error("Slot đã được đặt, chuyển về chọn giờ khác...")`
- Success: `toast.success("Đặt lịch thành công!")`

### 3. Skeleton Loading States

**`PageSkeleton.tsx`** — layout skeleton dùng chung:
```tsx
export const CardSkeleton = () => (
  <div className="rounded-lg border p-4 space-y-3 animate-pulse">
    <div className="h-5 bg-slate-200 rounded w-3/4" />
    <div className="h-4 bg-slate-200 rounded w-1/2" />
    <div className="h-4 bg-slate-200 rounded w-2/3" />
  </div>
);

export const GridSkeleton = ({ count = 3 }) => (
  <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
    {Array.from({ length: count }).map((_, i) => <CardSkeleton key={i} />)}
  </div>
);
```

**Áp dụng vào các trang:**

`LandingPage.tsx`:
```tsx
// Khi isLoading: hiển thị GridSkeleton thay vì "Đang tải..."
{isLoading ? <GridSkeleton count={3} /> : <div className="grid ...">{services.map(...)}</div>}
```

`BookingPage.tsx`:
- Trong `StepService`: skeleton khi fetch services
- Trong `StepDoctor`: skeleton khi fetch doctors theo service
- Trong `StepSlot`: skeleton khi fetch slots

`MyAppointmentsPage.tsx`:
- Skeleton khi fetch appointments

### 4. Empty State Component — `EmptyState.tsx`

```tsx
interface EmptyStateProps {
  icon?: ReactNode;
  title: string;
  description?: string;
  action?: ReactNode;
}

export const EmptyState = ({ icon, title, description, action }: EmptyStateProps) => (
  <div className="flex flex-col items-center justify-center py-16 gap-4 text-center">
    {icon && <div className="text-slate-300 text-5xl">{icon}</div>}
    <h3 className="text-lg font-semibold text-slate-700">{title}</h3>
    {description && <p className="text-slate-500 max-w-sm">{description}</p>}
    {action}
  </div>
);
```

**Dùng trong `MyAppointmentsPage.tsx`:**
```tsx
{appointments.length === 0 && (
  <EmptyState
    title="Bạn chưa có lịch hẹn nào"
    description="Đặt lịch khám ngay để được hỗ trợ sức khỏe tốt nhất."
    action={<Link to="/book"><Button>Đặt lịch ngay</Button></Link>}
  />
)}
```

### 5. Disable Button khi Submit

`StepConfirm.tsx`:
```tsx
<Button 
  onClick={handleConfirm} 
  disabled={isBooking}
  className="min-w-[140px]"
>
  {isBooking ? (
    <span className="flex items-center gap-2">
      <Loader2 className="w-4 h-4 animate-spin" /> Đang xử lý...
    </span>
  ) : "Xác nhận đặt lịch"}
</Button>
```

### 6. SEO Meta Tags

Cài `react-helmet-async` nếu chưa có:
```bash
npm install react-helmet-async
```

Wrap `<App />` với `<HelmetProvider>`.

**`LandingPage.tsx`:**
```tsx
import { Helmet } from 'react-helmet-async';

<Helmet>
  <title>HIS International Clinic — Đặt lịch khám trực tuyến</title>
  <meta name="description" content="Đặt lịch khám bệnh dễ dàng, nhanh chóng tại HIS International Clinic. Đội ngũ bác sĩ chuyên khoa giàu kinh nghiệm." />
  <meta property="og:title" content="HIS International Clinic" />
  <meta property="og:description" content="Đặt lịch khám ngay hôm nay." />
</Helmet>
```

**`BookingPage.tsx`:**
```tsx
<Helmet>
  <title>Đặt lịch khám — HIS International Clinic</title>
  <meta name="description" content="Chọn dịch vụ, bác sĩ và khung giờ phù hợp để đặt lịch khám." />
</Helmet>
```

### 7. Responsive Mobile Check

Test trên viewport 360px (Android S10), 390px (iPhone 14):
- LandingPage: grid 1 cột trên mobile
- BookingPage: form field full-width
- MyAppointmentsPage: card stack dọc
- Fix overflow-x nếu có
- Button size đủ lớn (min 44px height) cho touch

**CSS fixes nếu cần (trong `index.css`):**
```css
@media (max-width: 640px) {
  .booking-stepper { font-size: 11px; gap: 4px; }
}
```

## Kiểm tra hoàn thành
- [ ] `npm run build` thành công
- [ ] Không còn `window.alert` hay `window.confirm` trong codebase web
- [ ] Skeleton hiển thị khi load LandingPage, BookingPage, MyAppointmentsPage
- [ ] Error boundary: simulate lỗi API → UI hiển thị thông báo đẹp, không white screen
- [ ] My Appointments: empty state với nút "Đặt lịch ngay" hoạt động
- [ ] Submit booking: nút disabled + spinner khi đang call API
- [ ] SEO: inspect HTML có đúng `<title>` và `<meta name="description">`
- [ ] Mobile 360px: không bị overflow ngang, button đủ lớn để tap
- [ ] Lighthouse SEO score ≥ 90
