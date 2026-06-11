# Sprint 6 — Pharmacy & Inventory (Tuần 11–12)

> **Mục tiêu:** Kê đơn thuốc, kiểm tra tương tác, quy trình xuất thuốc, quản lý tồn kho.
> **Prerequisite:** Sprint 5 hoàn thành. Visit orders hoạt động.
> **Kết thúc sprint:** Doctor kê đơn trong Visit screen; Pharmacist duyệt và xuất thuốc; Stock tự giảm.

---

## BACKEND

### Module `internal/pharmacy`

**Domain layer:**
- [ ] Entity `DrugCatalog`: id, name_generic, name_brand, atc_code, form (tablet/syrup...), unit, strength, is_controlled, is_active
- [ ] Entity `DrugInteraction`: drug_a_id, drug_b_id, severity (MINOR/MODERATE/MAJOR/CONTRAINDICATED), description
- [ ] Entity `DrugContraindication`: drug_id, condition (e.g. "pregnancy", "renal failure"), severity
- [ ] Entity `Prescription`: id, visit_id, patient_id, doctor_id, status, prescribed_at
- [ ] Status: `PENDING` → `PHARMACIST_VERIFIED` → `DISPENSED` | `REJECTED`
- [ ] Entity `PrescriptionItem`: prescription_id, drug_id, quantity, dosage, frequency, duration_days, note
- [ ] Entity `DispensingRecord`: prescription_id, pharmacist_id, dispensed_at
- [ ] Entity `DispensingItem`: record_id, drug_id, lot_id, quantity_dispensed
- [ ] Repository interfaces: `PrescriptionRepository`, `DrugRepository`

**Application layer:**
- [ ] Command: `CreatePrescriptionCommand` — tạo đơn thuốc từ visit
- [ ] Command: `VerifyPrescriptionCommand` — pharmacist review
- [ ] Command: `DispensePrescriptionCommand` — xuất thuốc → atomic stock deduction
- [ ] Command: `RejectPrescriptionCommand`
- [ ] Query: `SearchDrugs`, `GetDrugInteractions`, `GetPrescriptionQueue`

> ⚠️ **NOTE:** `DispensePrescriptionCommand` phải dùng **PostgreSQL transaction**:
> 1. Lock stock lot rows (`SELECT FOR UPDATE`)
> 2. Check đủ số lượng
> 3. Deduct stock (INSERT stock_transaction type=OUT)
> 4. Update prescription status = DISPENSED
> Nếu bất kỳ bước nào fail → rollback toàn bộ.

### Drug Interaction Check Logic
- [ ] `GetDrugInteractions(drug_ids []string)` — tìm tất cả cặp tương tác trong `drug_interactions`
- [ ] Group theo severity: CONTRAINDICATED > MAJOR > MODERATE > MINOR
- [ ] Response: `{ interactions: [{ drug_a, drug_b, severity, description }] }`

### APIs — Pharmacy

```
GET  /api/v1/drugs/search               [DOCTOR, PHARMACIST]
  query: q (name, ATC code)
  response: list drugs (id, name, form, unit, strength)

GET  /api/v1/drugs/:id                  [DOCTOR, PHARMACIST]
GET  /api/v1/drugs/interactions         [DOCTOR, PHARMACIST]
  query: drug_ids=id1,id2,id3 (comma-separated)
  response: { interactions: [...] }

POST /api/v1/pharmacy/prescriptions     [DOCTOR]
  body: { visit_id, items: [{ drug_id, quantity, dosage, frequency, duration_days }] }

GET  /api/v1/pharmacy/prescriptions     [PHARMACIST]
  query: status, date
  response: prescription queue

GET  /api/v1/pharmacy/prescriptions/:id [PHARMACIST, DOCTOR]

PUT  /api/v1/pharmacy/prescriptions/:id/verify   [PHARMACIST]
  body: { notes? }

PUT  /api/v1/pharmacy/prescriptions/:id/dispense [PHARMACIST]
  body: { items: [{ drug_id, lot_id, quantity_dispensed }] }
  → Atomic transaction: check stock → deduct → update status

PUT  /api/v1/pharmacy/prescriptions/:id/reject   [PHARMACIST]
  body: { reason }
```

> ⚠️ **NOTE:** Nếu stock không đủ khi dispense → trả `422 Unprocessable Entity`
> với message "Không đủ tồn kho: {drug_name} (cần {required}, còn {available})".
> Frontend phải hiển thị lỗi này rõ ràng.

---

### Module `internal/inventory`

**Domain layer:**
- [ ] Entity `Warehouse`: id, name, location, type (PHARMACY/GENERAL)
- [ ] Entity `InventoryItem`: drug_id, warehouse_id, min_stock_quantity
- [ ] Entity `StockLot`: id, drug_id, warehouse_id, quantity, lot_number, expiry_date, unit_price
- [ ] Entity `StockTransaction`: lot_id, type (IN/OUT/ADJUSTMENT), quantity, reference_type, reference_id, created_by, created_at
- [ ] Repository interface: `InventoryRepository`

**Application layer:**
- [ ] Query: `GetStockByDrug`, `GetLowStockItems`, `GetStockTransactions`
- [ ] Command: `AdjustStockCommand` — manual adjustment
- [ ] Low stock check after dispense: nếu current_stock < min_stock → publish event

### APIs — Inventory

```
GET  /api/v1/inventory/items            [PHARMACIST, ADMIN]
  query: warehouse_id, low_stock (bool)
  response: list items với current_stock (sum của lots)

GET  /api/v1/inventory/items/:drug_id/lots [PHARMACIST]
  response: list lots còn hàng (quantity > 0, chưa hết hạn)
  → Dùng để chọn lot khi dispense

GET  /api/v1/inventory/transactions     [ADMIN, PHARMACIST]
  query: drug_id, from, to
```

### NATS Events
- [ ] Sau dispense: check stock → nếu < min → publish `HIS.INVENTORY.LowStockAlert`
- [ ] audit worker log prescription events

---

## DESKTOP

### Prerequisite
- Drug search, Interaction check, Prescription CRUD APIs phải ready
- Inventory lots API phải ready

### Visit Screen — Kê Đơn Thuốc (Doctor)

**Tab Kê đơn (bổ sung vào Visit screen Sprint 5):**
- [ ] Drug search autocomplete: `GET /drugs/search?q=...` debounce 300ms
  - Hiển thị: tên generic, dạng bào chế, hàm lượng, đơn vị
- [ ] Thêm vào đơn: form nhỏ { liều dùng, tần suất, số ngày }
- [ ] Tag list các thuốc đã thêm
- [ ] **Drug interaction check realtime**: mỗi khi thêm thuốc mới → `GET /drugs/interactions?drug_ids=...`
  - MINOR/MODERATE: hiện warning badge, vẫn cho kê
  - MAJOR: hiện warning modal nổi bật (màu cam), hỏi confirm
  - CONTRAINDICATED: hiện error modal (màu đỏ), **block submit**
- [ ] Nút "Gửi đơn thuốc" → `POST /pharmacy/prescriptions`

> ⚠️ **NOTE:** CONTRAINDICATED phải **block hoàn toàn** — không cho submit dù user muốn.
> MAJOR chỉ cần warning và confirm, không block.

### Pharmacist Screens

**Prescription Queue:**
- [ ] Tab: Chờ duyệt | Đã duyệt | Đã xuất | Từ chối
- [ ] `GET /pharmacy/prescriptions?status=PENDING&date=today`
- [ ] Card mỗi đơn: tên bệnh nhân, bác sĩ, số lượng thuốc, thời gian kê
- [ ] Polling mỗi 15s để cập nhật hàng đợi mới

**Chi tiết đơn thuốc + Verify:**
- [ ] Danh sách items: tên thuốc, liều, tần suất, số ngày, tổng số lượng
- [ ] Hiển thị drug interactions nếu có (từ backend check)
- [ ] Nút "Duyệt" → `PUT /pharmacy/prescriptions/:id/verify`
- [ ] Nút "Từ chối" → modal nhập lý do → `PUT /pharmacy/prescriptions/:id/reject`

**Dispensing (Xuất thuốc):**
- [ ] Sau khi verify → nút "Xuất thuốc"
- [ ] Mỗi item: chọn lot từ dropdown (`GET /inventory/items/:drug_id/lots`)
  - Dropdown hiển thị: lot number, hạn dùng, số lượng còn
  - Mặc định chọn lot FEFO (First Expired First Out — hết hạn trước xuất trước)
- [ ] Confirm số lượng xuất (auto-fill = số lượng kê đơn)
- [ ] Submit → `PUT /pharmacy/prescriptions/:id/dispense`
- [ ] Xử lý lỗi stock không đủ: hiển thị message cụ thể từ backend
- [ ] In nhãn thuốc sau khi xuất thành công

**In nhãn thuốc (React-to-print):**
- [ ] Template nhãn: tên thuốc, liều dùng, tần suất, tên bệnh nhân, ngày kê
- [ ] In nhiều nhãn nếu nhiều loại thuốc

**Inventory View:**
- [ ] Table tồn kho: tên thuốc, số lot, tổng số lượng, đơn vị, cảnh báo hết hàng
- [ ] Filter: kho, low_stock only
- [ ] Badge đỏ nếu stock < min_stock

---

## WEB

> Sprint 6 không có feature mới cho Web.
> Web team tiếp tục:

- [ ] Hoàn thiện Account page (`/account`): form cập nhật thông tin, BHYT
- [ ] Lịch sử khám (`/account` tab "Lịch sử"): `GET /patients/me/visits`
- [ ] Xem hóa đơn (placeholder — `GET /billing/invoices/:id` sẽ ready Sprint 7)
- [ ] Performance audit: bundle size, lazy loading

---

## ĐIỂM KẾT NỐI Sprint 6

| Vấn đề | Backend | Desktop (Doctor) | Desktop (Pharmacist) |
|--------|---------|-----------------|---------------------|
| Drug interaction | Query tất cả cặp từ DB | Realtime check khi thêm thuốc | Hiển thị khi review đơn |
| Stock deduction | Atomic PG transaction | — | Xử lý lỗi stock thiếu |
| FEFO lot selection | Trả lots sorted by expiry_date ASC | — | Auto-select lot đầu tiên |
| Low stock alert | NATS event sau dispense | — | Badge cảnh báo trong Inventory |

## DEFINITION OF DONE

- [ ] Drug search autocomplete < 200ms response
- [ ] Drug interaction check đúng với dữ liệu seed (ít nhất 10 cặp test)
- [ ] CONTRAINDICATED block submit hoàn toàn
- [ ] Prescription full flow: Doctor kê → Pharmacist verify → Pharmacist dispense
- [ ] Stock deduction atomic: test concurrent dispense không race condition
- [ ] Lỗi stock không đủ trả đúng message
- [ ] FEFO: lot gần hết hạn được chọn trước
- [ ] In nhãn thuốc từ Desktop thành công
- [ ] Low stock alert publish khi stock < min
