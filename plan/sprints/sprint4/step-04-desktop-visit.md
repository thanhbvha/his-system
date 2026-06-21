# Sprint 4 — Step 4: Desktop — Doctor Worklist + Visit Screen + Vitals (Frontend)

## Mục tiêu
Xây dựng màn hình Worklist của Bác sĩ (danh sách bệnh nhân đang chờ khám) và Visit Screen để nhập Vitals, xem lịch sử bệnh nhân, tạo chỉ định xét nghiệm.

## Prerequisite
- Step 2 hoàn thành: Visit API, Vitals API, Orders API sẵn sàng.
- Step 3 hoàn thành: `useQueueWS` hook sẵn sàng (Worklist cần subscribe WS updates).
- Sprint 3 ✅: `apiClient.ts`, `patientStore` (getPatientDetail), i18n chuẩn.

## Files cần tạo / cập nhật

```
desktop/frontend/src/
├── store/
│   └── visitStore.ts            -- Zustand store cho visit data
├── components/
│   ├── visit/
│   │   ├── DoctorWorklist.tsx   -- Danh sách bệnh nhân chờ khám
│   │   ├── VisitScreen.tsx      -- Màn hình khám chính (tabs)
│   │   ├── VitalsForm.tsx       -- Form nhập vitals
│   │   ├── VitalsHistory.tsx    -- Lịch sử vitals
│   │   ├── OrdersPanel.tsx      -- Panel chỉ định xét nghiệm
│   │   └── PatientHistoryTab.tsx -- Lịch sử khám từ Sprint 3
│   └── icd10/
│       └── ICD10Search.tsx      -- Combobox tìm kiếm ICD-10
├── pages/
│   └── VisitPage.tsx            -- Page wrapper Visit Screen
└── i18n/
    ├── vi.json                  -- Thêm keys visit.*
    └── en.json                  -- Thêm keys visit.*
```

## Nhiệm vụ chi tiết

### 1. Zustand Store — `visitStore.ts`

```typescript
interface VisitVital {
  id: string;
  bp_systolic?: number;
  bp_diastolic?: number;
  heart_rate?: number;
  temperature?: number;
  spo2?: number;
  weight_kg?: number;
  height_cm?: number;
  recorded_at: string;
}

interface VisitOrder {
  id: string;
  order_type: 'LAB' | 'RADIOLOGY' | 'PROCEDURE';
  details: string;
  status: string;
}

interface Visit {
  id: string;
  patient: { id: string; full_name: string; dob: string; gender: string };
  doctor: { id: string; full_name: string };
  status: string;
  chief_complaint?: string;
  started_at?: string;
  vitals: VisitVital[];
  orders: VisitOrder[];
}

interface VisitState {
  worklist: Visit[];
  selectedVisit: Visit | null;
  isLoading: boolean;

  fetchWorklist: (doctorId?: string, date?: string) => Promise<void>;
  fetchVisitDetail: (visitId: string) => Promise<void>;
  createVisit: (payload: CreateVisitPayload) => Promise<Visit>;
  recordVitals: (visitId: string, vitals: Partial<VisitVital>) => Promise<void>;
  createOrder: (visitId: string, order: { order_type: string; details: string }) => Promise<void>;
  closeVisit: (visitId: string) => Promise<void>;
}
```

### 2. Doctor Worklist — `DoctorWorklist.tsx`

- Layout: Table/List với các cột:
  - Số TT (queue_number từ QueueEntry)
  - Tên bệnh nhân (full_name)
  - Giờ check-in (created_at của queue entry)
  - Lý do khám (chief_complaint nếu đã có)
  - Trạng thái (badge màu)
- Subscribe WS event `queue.updated` → auto refresh worklist (không cần F5)
- Click vào row → navigate tới `VisitPage` với `visit_id`
- Nút "Tạo Visit" (nếu queue entry chưa có visit): gọi `createVisit()`

### 3. Visit Screen — `VisitScreen.tsx`

Header cố định:
```
[ Tên bệnh nhân ] | [ Tuổi - Giới tính ] | Status badge | [ Kết thúc khám ]
```

3 Tabs:

**Tab 1: Vitals** — `VitalsForm.tsx`
- Form nhập: Huyết áp (tâm thu / tâm trương), Mạch, Nhiệt độ, SpO2, Cân nặng, Chiều cao
- Highlight bất thường (màu đỏ/cam):
  ```
  BP > 140/90 hoặc < 90/60 → cảnh báo
  HR > 100 hoặc < 60 → cảnh báo
  Temp > 37.5°C → cảnh báo
  SpO2 < 95% → cảnh báo
  ```
- Submit → `visitStore.recordVitals(visitId, vitals)`
- `VitalsHistory.tsx`: danh sách lần đo trước, hiển thị dạng timeline

**Tab 2: Chỉ định & Chẩn đoán** — `OrdersPanel.tsx`
- `ICD10Search` combobox: debounce 300ms → `GET /icd10/search?q=`
- Nút "Thêm chỉ định": chọn loại (LAB/RADIOLOGY/PROCEDURE), nhập mô tả
- Danh sách chỉ định đã tạo với trạng thái
- Placeholder note: "Chức năng kê đơn thuốc sẽ có ở Sprint 5"

**Tab 3: Lịch sử khám** — `PatientHistoryTab.tsx`
- Tái sử dụng component hiển thị lịch sử từ `patientStore.getPatientDetail()`
- Hiển thị các lần khám trước: ngày, bác sĩ, chẩn đoán

### 4. ICD-10 Search Component — `ICD10Search.tsx`

```typescript
// Ant Design AutoComplete hoặc Select với search
export const ICD10Search = ({ onSelect }) => {
  const [options, setOptions] = useState([]);
  
  const handleSearch = useMemo(
    () => debounce(async (query: string) => {
      const res = await apiClient.get(`/icd10/search?q=${query}`);
      setOptions(res.data.data.map(item => ({
        label: `${item.code} — ${item.description_vi}`,
        value: item.code,
      })));
    }, 300),
    []
  );
  
  return <AutoComplete options={options} onSearch={handleSearch} onSelect={onSelect} />;
};
```

### 5. Navigation
- Route `/visits/:id` → `<VisitPage>` (protected, DOCTOR/NURSE)
- Thêm menu item "Worklist" trong Sidebar (role DOCTOR, NURSE)

### 6. i18n Keys mới

**`vi.json`** — thêm namespace `visit`:
```json
{
  "visit": {
    "worklist": "Danh sách bệnh nhân",
    "startVisit": "Bắt đầu khám",
    "closeVisit": "Kết thúc khám",
    "vitals": "Sinh hiệu",
    "addVitals": "Ghi nhận sinh hiệu",
    "orders": "Chỉ định",
    "addOrder": "Thêm chỉ định",
    "history": "Lịch sử khám",
    "bp": "Huyết áp",
    "heartRate": "Nhịp tim",
    "temp": "Nhiệt độ",
    "spo2": "SpO2",
    "weight": "Cân nặng",
    "height": "Chiều cao",
    "abnormal": "Bất thường",
    "icd10Search": "Tìm mã ICD-10",
    "confirmClose": "Xác nhận kết thúc buổi khám?",
    "noOrders": "Chưa có chỉ định nào"
  }
}
```

### 2. Cập nhật Kiến trúc WebSocket: Phân Shard theo Phòng Khám (Clinic Room-based Sharding)

**Mục tiêu:**
Thay vì phân Shard bằng cách băm (hash) ngẫu nhiên UserID hoặc gộp toàn bộ vào một `queue_shard` duy nhất, chúng ta sẽ chuyển sang kiến trúc **Phân mảnh theo Phòng khám (Room-based Sharding)** để phù hợp với Domain-Driven Design (DDD).

**Lý do:**
1. **Phù hợp với nghiệp vụ thực tế:** Bác sĩ, Y tá, và Bệnh nhân quan tâm đến dữ liệu Hàng Đợi (Queue) theo từng phòng khám riêng biệt (Ví dụ: Phòng Khám Nội 1, Phòng Chụp X-Quang 2).
2. **Tối ưu Hóa Hiệu Năng (Scalability):** Hệ thống không cần Broadcast sự kiện gọi số tới toàn bộ mạng WebSocket. Chỉ những Desktop/Màn hình TV đang trực tại phòng khám đó mới nhận được bản tin qua `shard:room-1`.
3. **Thống kê Real-time chính xác:** Số lượng bệnh nhân đang chờ, đang khám của từng phòng được khoanh vùng cục bộ và đẩy ra chính xác màn hình ngoài cửa phòng khám.

**Kế hoạch triển khai:**
1. **Frontend (Desktop App):**
   - Lấy thông tin **Room ID** mà Bác sĩ/Y tá đang trực (từ `authStore` hoặc `userContext` sau khi chọn phòng khám lúc bắt đầu ca làm việc).
   - Truyền `room_id` lên Backend khi mở kết nối WebSocket (vd: `ws://.../ws/queue?token=...&room_id=R-101`).

2. **Backend - Tầng Adapter (`pkg/ws/adapter.go`):**
   - Đọc tham số `room_id` từ query string (hoặc từ thông tin ca trực của JWT payload).
   - Thiết lập `shardID := "room_" + roomID` thay vì chuỗi cứng `"queue_shard"`.
   - Nếu Client là Tivi hiển thị của khu vực (chưa login cụ thể), TV đó cũng sẽ connect với `room_id` tương ứng của khu vực.

3. **Backend - Tầng API / Command (`commands/call_queue.go`, `event.go`):**
   - Các API thay đổi trạng thái hàng đợi (Gọi số, Bỏ qua, Hoàn tất) cần query/trích xuất thông tin `room_id` đang xử lý từ Context/Entity.
   - Sửa lại hàm `Broadcast` để bắn đích danh Pub/Sub vào Shard của phòng đó:
     ```go
     pubsubManager.BroadcastMessage("room_" + cmd.RoomID, data)
     ```

## Kiểm tra hoàn thành
- [x] `npm run build` thành công
- [x] Worklist hiển thị đúng danh sách bệnh nhân của bác sĩ hôm nay
- [x] Worklist tự cập nhật khi có check-in mới (WS `queue.updated`)
- [x] Kiến trúc Room-based Sharding hoạt động ổn định trên môi trường cluster
- [x] Nhập vitals thành công, giá trị bất thường highlight màu đỏ
- [x] ICD-10 search: gõ "tim" → trả kết quả, debounce 300ms
- [x] Tạo chỉ định xét nghiệm thành công
- [x] Nút "Kết thúc khám" → visit status = COMPLETED
- [x] i18n hoạt động đúng khi đổi ngôn ngữ
