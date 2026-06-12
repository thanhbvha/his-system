# Sprint 5 — EMR & LIS (Tuần 9–10)

> **Mục tiêu:** Bệnh án điện tử (EMR/SOAP với versioning), hoàn chỉnh luồng xét nghiệm LIS, Web xem kết quả.
> **Prerequisite:** Sprint 4 hoàn thành. Visit APIs, VisitOrder creation hoạt động.
> **Kết thúc sprint:** Doctor viết SOAP + chẩn đoán; Lab Tech nhận mẫu → nhập kết quả → verify; Web bệnh nhân xem kết quả.

---

## BACKEND

### Module `internal/emr` (MongoDB)

**Domain layer:**
- [ ] Document `MedicalRecord`:
  ```js
  {
    _id, patient_id, visit_id, version: int,
    soap: {
      subjective: string,   // Lý do khám, triệu chứng (bệnh nhân mô tả)
      objective: string,    // Khám thực thể, vitals (bác sĩ ghi nhận)
      assessment: string,   // Nhận định lâm sàng
      plan: string          // Kế hoạch điều trị
    },
    diagnoses: [{ icd10_code, description, type: primary|secondary }],
    orders: [{ order_type, ref_id, created_at }],
    attachments: [{ file_name, minio_key, uploaded_at }],
    created_by, updated_by, created_at,
    history: [{ version, soap, diagnoses, changed_by, changed_at }]
  }
  ```
- [ ] **Immutable versioning**: mỗi lần update → tạo version mới, append history, KHÔNG xóa version cũ
- [ ] Repository interface: `EMRRepository`

**Application layer:**
- [ ] Command: `CreateOrUpdateSOAPCommand` — upsert với version increment
- [ ] Command: `FinalizeDiagnosisCommand` — lock diagnoses sau khi khám xong
- [ ] Command: `UploadAttachmentCommand` — upload MinIO → lưu ref trong MongoDB
- [ ] Query: `GetMedicalRecord(patient_id, visit_id)`, `GetPatientEMRHistory(patient_id)`
- [ ] Query: `GetRecordVersion(patient_id, visit_id, version)`

> ⚠️ **NOTE:** Versioning logic phải test kỹ:
> - Không được overwrite version cũ
> - `history` array phải append-only
> - Nếu 2 request đồng thời update → dùng MongoDB optimistic lock (version check)

### APIs — EMR

```
GET  /api/v1/emr/:patient_id             [DOCTOR, NURSE]
  query: visit_id (optional, lấy record của visit cụ thể)
  response: latest version of medical record

POST /api/v1/emr                         [DOCTOR]
  body: { patient_id, visit_id, soap: { subjective, objective, assessment, plan } }
  → Upsert với version++ → append history entry
  response: { version, updated_at }

GET  /api/v1/emr/:patient_id/history     [DOCTOR]
  response: list versions (summary: version, changed_by, changed_at)

GET  /api/v1/emr/:patient_id/history/:version  [DOCTOR]
  response: full record tại version đó

POST /api/v1/emr/:patient_id/diagnoses   [DOCTOR]
  body: { diagnoses: [{ icd10_code, type }] }
  → Append diagnoses vào current version

POST /api/v1/emr/attachments             [DOCTOR, NURSE]
  body: multipart/form-data (file)
  → Upload MinIO → lưu ref trong EMR document
  response: { file_name, url }

GET  /api/v1/patients/me/lab-results     [PATIENT - Web]
  response: list lab orders có status=VERIFIED của patient hiện tại
```

---

### Module `internal/lis` (LIS Full Flow)

**Domain layer:**
- [ ] Entity `LabOrder`: id, visit_id, patient_id, doctor_id, items[], status, ordered_at
- [ ] Entity `LabOrderItem`: order_id, test_id (ref lab_test_catalog), status
- [ ] Entity `LabSample`: order_id, barcode, collected_at, collected_by, rejection_reason?
- [ ] Status machine: `ORDERED` → `SAMPLE_RECEIVED` → `PROCESSING` → `RESULTED` → `VERIFIED`
- [ ] Entity `LabResult`: order_id, items[], verified_by, verified_at
- [ ] Entity `LabResultItem`: result_id, test_id, value, unit, reference_range, flag (H/L/N)
- [ ] MongoDB document `LabResultDocument`: order_id, results_detail[], raw_data
- [ ] Repository interfaces: `LabOrderRepository`, `LabResultRepository`

**Application layer:**
- [ ] Command: `ReceiveSampleCommand`
- [ ] Command: `RecordResultsCommand`
- [ ] Command: `VerifyResultsCommand` → publish `LabResultVerified`
- [ ] Query: `GetLabWorklist`, `GetOrderDetail`, `GetResultDetail`

### APIs — LIS

```
GET  /api/v1/lab/orders                  [LAB_TECH]
  query: status, date
  response: worklist (orders cần xử lý)

GET  /api/v1/lab/orders/:id              [LAB_TECH, DOCTOR]

PUT  /api/v1/lab/orders/:id/sample       [LAB_TECH]
  body: { barcode, notes? }
  → Update status SAMPLE_RECEIVED

POST /api/v1/lab/orders/:id/results      [LAB_TECH]
  body: { items: [{ test_id, value, unit }] }
  → So sánh với reference_range → tính flag H/L/N
  → Lưu PG (summary) + MongoDB (chi tiết)
  → Update status RESULTED

PUT  /api/v1/lab/orders/:id/verify       [LAB_TECH]
  → Update status VERIFIED
  → Publish HIS.LIS.LabResultVerified

GET  /api/v1/lab/results/:id             [LAB_TECH, DOCTOR, PATIENT (nếu verified)]
  response: { items: [{ name, value, unit, reference_range, flag }] }
```

> ⚠️ **CRITICAL — Cross-client dependency:**
> Sau khi Lab Tech gọi `PUT /lab/orders/:id/verify`:
> - Backend publish `HIS.LIS.LabResultVerified`
> - Notification worker gửi SMS cho bệnh nhân
> - Web `/results` mới hiển thị được kết quả
>
> **Trước verify:** `GET /patients/me/lab-results` KHÔNG trả về order này.
> Backend phải filter `status = VERIFIED` nghiêm ngặt.

### Redis Stream Events
- [ ] `HIS.LIS.SampleCollected` — audit log
- [ ] `HIS.LIS.LabResultVerified` → notification worker gửi SMS "Kết quả xét nghiệm đã có"

---

## DESKTOP

### Prerequisite
- EMR APIs (SOAP, diagnoses, attachments) phải ready
- LIS APIs (worklist, sample, results, verify) phải ready

### Visit Screen — Hoàn chỉnh (Doctor)

**Tab SOAP Editor:**
- [ ] 4 section: Subjective / Objective / Assessment / Plan
- [ ] Textarea hoặc rich text (quill editor nhẹ)
- [ ] **Auto-save mỗi 30 giây** → `POST /emr` silent (không hiện spinner)
- [ ] Indicator "Đã lưu lúc HH:mm"
- [ ] Version history button → mở panel bên phải xem lịch sử

> ⚠️ **NOTE:** Auto-save phải dùng debounce 30s từ lần gõ cuối cùng, không phải interval cứng.
> Dùng `useRef` để lưu timeout ID, clear khi unmount component.

**Tab Chẩn đoán:**
- [ ] ICD-10 search (autocomplete): `GET /icd10/search?q=...` debounce 300ms
- [ ] Thêm nhiều chẩn đoán (chính + phụ)
- [ ] Tag list hiển thị các chẩn đoán đã chọn
- [ ] Submit → `POST /emr/:patient_id/diagnoses`

**Tab Chỉ định:**
- [ ] Form tạo lab order: chọn test catalog → `POST /visits/:id/orders`
- [ ] Form kê đơn thuốc (placeholder → Sprint 6)
- [ ] Kết quả XN inline viewer: `GET /lab/results/:id` (nếu đã có)

**Nút "Kết thúc khám":**
- [ ] Confirm dialog: "Xác nhận kết thúc? Hóa đơn sẽ được tạo tự động."
- [ ] Submit → `POST /visits/:id/close`

**EMR History Panel:**
- [ ] Timeline versions: v1, v2, v3...
- [ ] Click version → xem diff hoặc full content tại thời điểm đó
- [ ] `GET /emr/:patient_id/history` + `GET /emr/:patient_id/history/:version`

### Lab Tech Screens

**Worklist:**
- [ ] `GET /lab/orders?status=ORDERED,SAMPLE_RECEIVED,RESULTED&date=today`
- [ ] Kanban hoặc Table theo status
- [ ] Màu badge theo status (Ant Design Tag màu)

**Nhận mẫu:**
- [ ] Input barcode (text input, tương thích barcode scanner USB HID)
- [ ] Submit → `PUT /lab/orders/:id/sample`
- [ ] Beep/feedback khi nhận thành công

> ⚠️ **NOTE:** Barcode scanner USB HID hoạt động như keyboard input thông thường —
> không cần Wails binding đặc biệt, chỉ cần focus vào input.

**Nhập kết quả:**
- [ ] Table: danh sách test items trong order
- [ ] Input value cho từng item
- [ ] Hiển thị reference range bên cạnh (lấy từ `lab_test_catalog`)
- [ ] Highlight H/L preview (client-side tính dựa trên reference range)
- [ ] Submit → `POST /lab/orders/:id/results`

**Verify & Approve:**
- [ ] Review lại kết quả đã nhập
- [ ] Xác nhận → `PUT /lab/orders/:id/verify`
- [ ] Sau verify: hiển thị thông báo "SMS đã gửi cho bệnh nhân"

**In phiếu kết quả:**
- [ ] React-to-print: render template phiếu kết quả XN
- [ ] `GET /lab/results/:id` → render → print

---

## WEB

### Prerequisite
- `GET /patients/me/lab-results` phải ready (chỉ trả VERIFIED)
- `GET /lab/results/:id` phải accessible cho PATIENT role

### Lab Results Page (`/results`)
- [ ] `GET /patients/me/lab-results` → danh sách lần xét nghiệm
- [ ] Sort by date descending
- [ ] Card mỗi lần XN: ngày, tên panel, bác sĩ chỉ định, số lượng items bất thường
- [ ] Click → chi tiết: `GET /lab/results/:id`

**Chi tiết kết quả:**
- [ ] Table: Tên xét nghiệm | Kết quả | Đơn vị | Bình thường | Đánh giá
- [ ] Row highlight đỏ nếu flag = `H` (cao), cam nếu `L` (thấp)
- [ ] Badge `H` / `L` / `N` trong cột Đánh giá
- [ ] Header: ngày lấy mẫu, ngày có kết quả, tên Lab Tech verify

**Polling:**
- [ ] TanStack Query `refetchInterval: 5 * 60 * 1000` (5 phút) để tự động cập nhật
- [ ] Nút "Làm mới" manual

**Empty state:**
- [ ] "Chưa có kết quả xét nghiệm" + icon

---

## ĐIỂM KẾT NỐI Sprint 5

| Vấn đề | Backend | Desktop (Lab Tech) | Web (Patient) |
|--------|---------|-------------------|---------------|
| Lab result gating | Filter `status=VERIFIED` trong patient endpoint | Lab Tech click "Verify" | Kết quả xuất hiện sau verify |
| SMS notification | Redis Stream → notification worker | Hiện thông báo "SMS đã gửi" | Bệnh nhân nhận SMS |
| EMR versioning | Append-only history | Auto-save SOAP | — |
| Attachment | MinIO upload | Form upload file | — |

## DEFINITION OF DONE

- [ ] SOAP editor auto-save hoạt động (test: ngắt mạng giữa chừng, không mất data)
- [ ] EMR versioning: mỗi update tạo version mới, history đầy đủ
- [ ] ICD-10 search autocomplete hoạt động
- [ ] LIS full flow: ORDERED → SAMPLE_RECEIVED → RESULTED → VERIFIED
- [ ] Sau verify: Web patient thấy kết quả, trước verify không thấy
- [ ] Flag H/L tính đúng theo reference range
- [ ] SMS notification gửi sau verify (hoặc log mock)
- [ ] Web: lab results page hiển thị đúng, highlight bất thường
- [ ] In phiếu kết quả từ Desktop thành công
- [ ] MongoDB EMR versioning test: concurrent update không overwrite
