# Sprint 4 — Step 3: Desktop — Queue Dashboard + Check-in Flow (Frontend)

## Mục tiêu
Xây dựng màn hình Queue Dashboard realtime (WebSocket) cho Lễ tân và luồng Check-in bệnh nhân trên Desktop App.

## Prerequisite
- Step 1 hoàn thành: WebSocket endpoint `/api/v1/queue/ws` và Queue APIs sẵn sàng.
- Sprint 3 ✅: `PatientSearchModal`, `publicStore` (services), `appointmentStore`, `apiClient.ts`, i18n chuẩn.

## Files cần tạo / cập nhật

```
desktop/frontend/src/
├── store/
│   └── queueStore.ts           -- Zustand store cho queue state + WS connection
├── components/
│   ├── queue/
│   │   ├── QueueDashboard.tsx  -- Màn hình Queue chính (Lễ tân)
│   │   ├── QueueColumn.tsx     -- Cột queue theo service_type
│   │   └── CheckInModal.tsx    -- Modal check-in bệnh nhân
│   └── ws/
│       └── useQueueWS.ts       -- Custom hook quản lý WebSocket connection
├── pages/
│   └── QueuePage.tsx           -- Page wrapper
└── i18n/
    ├── vi.json                 -- Thêm keys queue.*
    └── en.json                 -- Thêm keys queue.*
```

## Nhiệm vụ chi tiết

### 1. Zustand Store — `queueStore.ts`

```typescript
interface QueueEntry {
  id: string;
  patient: { id: string; full_name: string; patient_code: string };
  service_type: string;
  queue_number: string;
  status: 'WAITING' | 'CALLED' | 'IN_PROGRESS' | 'DONE' | 'SKIPPED';
  created_at: string;
  called_at?: string;
}

interface QueueState {
  entries: QueueEntry[];
  stats: { waiting_count: number; called_count: number; avg_wait_minutes: number } | null;
  isLoading: boolean;
  
  fetchQueue: () => Promise<void>;
  fetchStats: () => Promise<void>;
  checkIn: (payload: CheckInPayload) => Promise<QueueEntry>;
  callNext: (id: string) => Promise<void>;
  skip: (id: string) => Promise<void>;
  complete: (id: string) => Promise<void>;
  
  // WS sync
  applyWSEvent: (event: WSEvent) => void;
}
```

### 2. WebSocket Hook — `useQueueWS.ts`

```typescript
export function useQueueWS(token: string) {
  const { applyWSEvent } = useQueueStore();
  const wsRef = useRef<WebSocket | null>(null);
  const retryRef = useRef(0);
  const DELAYS = [1000, 2000, 4000, 8000, 30000]; // exponential backoff

  const connect = useCallback(() => {
    const ws = new WebSocket(`ws://...api/v1/queue/ws?token=${token}`);
    ws.onmessage = (e) => applyWSEvent(JSON.parse(e.data));
    ws.onclose = () => {
      const delay = DELAYS[Math.min(retryRef.current, DELAYS.length - 1)];
      retryRef.current++;
      setTimeout(() => connect(), delay);
    };
    wsRef.current = ws;
  }, [token]);

  useEffect(() => {
    connect();
    return () => wsRef.current?.close();
  }, [connect]);
}
```

> ⚠️ Khi WS disconnect do token expire (close code 1008):
> - Gọi `authStore.refreshToken()` để lấy token mới.
> - Reconnect với token mới.

### 3. Queue Dashboard — `QueueDashboard.tsx`

- Layout: grid các cột `<QueueColumn>` theo `service_type` (GENERAL, LAB, RADIOLOGY, ...)
- Mỗi cột hiển thị:
  - **Badge to** số đang được gọi (CALLED) — font lớn, màu nổi bật
  - **Danh sách** các số đang WAITING bên dưới
  - **Nút "Gọi số tiếp"** → `callNext(nextWaiting.id)` → hiển thị animation
  - **Nút "Bỏ qua"** → `skip(calledEntry.id)`
- WS event `queue.called` → animation highlight + sound (optional)
- Header: nút "Check-in mới" mở `<CheckInModal>`

### 4. Check-in Modal — `CheckInModal.tsx`

3 bước trong Ant Design `Steps` hoặc tự viết wizard:

**Bước 1 — Tìm bệnh nhân:**
- Tái sử dụng `<PatientSearchModal>` từ Sprint 3.
- Hiển thị thông tin bệnh nhân đã chọn.

**Bước 2 — Chọn dịch vụ:**
- Lấy danh sách từ `publicStore.services` (đã có ✅).
- Select dropdown hiển thị tên dịch vụ.

**Bước 3 — Link lịch hẹn (Optional):**
- Query `appointmentStore` lấy lịch hẹn hôm nay của bệnh nhân này.
- Nếu có: hiển thị danh sách để chọn link. Nếu không: bỏ qua.

**Submit:**
- Gọi `queueStore.checkIn({ patient_id, service_type, appointment_id? })`
- Hiển thị toast với số thứ tự được cấp (vd: `KB001`)

### 5. Navigation & Routing
- Thêm route `/queue` → `<QueuePage>` vào router Desktop App.
- Thêm menu item "Hàng đợi" trong Sidebar (dành cho role RECEPTIONIST, DOCTOR, NURSE).

### 6. i18n Keys mới

**`vi.json`** — thêm namespace `queue`:
```json
{
  "queue": {
    "title": "Hàng đợi hôm nay",
    "checkIn": "Check-in mới",
    "callNext": "Gọi số tiếp",
    "skip": "Bỏ qua",
    "waiting": "Đang chờ",
    "called": "Đang gọi",
    "noQueue": "Không có bệnh nhân đang chờ",
    "queueNumber": "Số thứ tự",
    "checkinSuccess": "Check-in thành công! Số: {{number}}",
    "searchPatient": "Tìm bệnh nhân",
    "selectService": "Chọn dịch vụ",
    "linkAppointment": "Liên kết lịch hẹn (Tùy chọn)"
  }
}
```

## Kiểm tra hoàn thành
- [ ] `npm run build` thành công, không có lỗi TypeScript
- [ ] WS kết nối thành công, nhận trạng thái queue ngay khi load trang
- [ ] Reconnect tự động sau khi mất kết nối (test bằng cách tắt server rồi bật lại)
- [ ] Gọi số → badge cập nhật trên tất cả màn hình Desktop đang mở
- [ ] Check-in flow: tìm bệnh nhân → chọn dịch vụ → submit → hiện số thứ tự
- [ ] i18n hoạt động đúng khi đổi ngôn ngữ
