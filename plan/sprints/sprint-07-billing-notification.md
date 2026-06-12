# Sprint 7 — Billing & Notification (Tuần 13–14)

> **Mục tiêu:** Hóa đơn tự động sau khám, thanh toán, PDF, SMS/Email notification.
> **Prerequisite:** Sprint 6 hoàn thành. `VisitClosed` event publish đúng.
> **Kết thúc sprint:** Hóa đơn tự tạo sau khám; Receptionist thu tiền + in biên lai; Web bệnh nhân xem hóa đơn; SMS/Email hoạt động.

---

## BACKEND

### Module `internal/billing`

**Domain layer:**
- [ ] Entity `PriceCatalog`: service_id, price, effective_from, effective_to, version
- [ ] Entity `PriceListItem`: catalog_id, item_type (SERVICE/DRUG/LAB/RADIOLOGY), item_id, price
- [ ] Entity `Invoice`: id, visit_id, patient_id, items[], subtotal, discount, total, status, issued_at
- [ ] Status: `DRAFT` → `ISSUED` → `PAID` | `PARTIALLY_PAID` | `CANCELLED`
- [ ] Entity `InvoiceItem`: invoice_id, item_type, description, quantity, unit_price, amount
- [ ] Entity `Payment`: invoice_id, amount, method (CASH/TRANSFER/CARD), reference_no?, paid_at, received_by
- [ ] Entity `Discount`: code, type (PERCENT/FIXED), value, condition, valid_until
- [ ] Repository interfaces: `InvoiceRepository`, `PaymentRepository`

**Application layer:**
- [ ] Command: `CreateInvoiceCommand` — triggered bởi `VisitClosed` event
  - Aggregate items: visit fees + lab orders + drugs từ prescription
  - Apply price catalog → tính subtotal, tax, total
  - **Idempotency**: check `visit_id` đã có invoice chưa trước khi tạo
- [ ] Command: `RecordPaymentCommand` — thu tiền
- [ ] Command: `ApplyDiscountCommand`
- [ ] Command: `CancelInvoiceCommand`
- [ ] Query: `GetInvoiceByVisit`, `GetInvoicesByPatient`, `GetPaymentHistory`

> ⚠️ **NOTE:** `CreateInvoiceCommand` phải **idempotent** — nếu `VisitClosed` event bị gửi 2 lần
> (Redis Stream at-least-once delivery), invoice chỉ được tạo 1 lần.
> Dùng `INSERT ... ON CONFLICT (visit_id) DO NOTHING`.

### Invoice PDF Generation
- [ ] Cài `chromedp` hoặc `go-wkhtmltopdf`
- [ ] HTML template invoice (CSS inline)
- [ ] `GET /billing/invoices/:id/pdf` → render HTML → convert PDF → stream response
- [ ] Content-Disposition: attachment hoặc inline (query param `?download=true`)

> ⚠️ **NOTE:** PDF generation nặng CPU. Implement với goroutine + timeout 30s.
> Cache PDF trong MinIO sau khi generate lần đầu (key: `invoices/{id}.pdf`).

### APIs — Billing

```
GET  /api/v1/billing/invoices            [RECEPTIONIST, ACCOUNTANT, ADMIN]
  query: patient_id, visit_id, status, from_date, to_date

GET  /api/v1/billing/invoices/:id        [RECEPTIONIST, ACCOUNTANT, PATIENT (của mình)]
  response: full invoice với items, payments

GET  /api/v1/billing/invoices/:id/pdf    [RECEPTIONIST, PATIENT]
  response: PDF stream
  query: download=true → Content-Disposition: attachment

POST /api/v1/billing/invoices/:id/pay    [RECEPTIONIST]
  body: { amount, method, reference_no? }
  → Tạo Payment record → cập nhật status invoice

GET  /api/v1/billing/payments            [ACCOUNTANT, ADMIN]
  query: from_date, to_date, method

GET  /api/v1/patients/me/invoices        [PATIENT - Web]
  response: list invoices của bệnh nhân hiện tại
```

> ⚠️ **NOTE:** `GET /billing/invoices/:id` cho PATIENT phải verify `invoice.patient_id = JWT.patient_id`.
> Không cho patient xem hóa đơn của người khác.

---

### Module `internal/notification` (Redis Stream Worker)

**Notification Worker (`cmd/worker/main.go`):**
- [ ] Subscribe các subjects sau:
  - `HIS.APPOINTMENT.AppointmentScheduled` → SMS xác nhận lịch hẹn
  - `HIS.APPOINTMENT.AppointmentCancelled` → SMS thông báo hủy
  - `HIS.APPOINTMENT.AppointmentConfirmed` → SMS xác nhận của phòng khám
  - `HIS.LIS.LabResultVerified` → SMS "Kết quả xét nghiệm đã có"
  - `HIS.BILLING.InvoiceCreated` → Email hóa đơn (optional)
  - `HIS.NOTIFICATION.NotificationRequested` → Generic notification

**SMS Gateway:**
- [ ] Interface `SMSProvider` với method `Send(phone, message string) error`
- [ ] Implementation: VNPT SMS hoặc mock (log to stdout nếu provider chưa config)
- [ ] Template message VI: `"Lịch hẹn của bạn với BS {name} lúc {time} đã được xác nhận. PK Xxx"`

**Email:**
- [ ] Interface `EmailProvider` với method `Send(to, subject, body string) error`
- [ ] Implementation: SMTP (config từ env) hoặc mock
- [ ] HTML email template cho hóa đơn

> ⚠️ **NOTE:** Provider không config → fallback sang mock (log). KHÔNG panic.
> Implement retry: nếu gửi fail → retry 3 lần với delay 5s, 10s, 30s.

### Redis Stream Events
- [ ] `HIS.BILLING.InvoiceCreated` — publish sau khi tạo invoice thành công
- [ ] `HIS.BILLING.PaymentReceived` — publish sau khi thu tiền

---

## DESKTOP

### Prerequisite
- Billing APIs phải ready
- PDF generation hoạt động

### Billing Screen (Receptionist)

**Tìm hóa đơn:**
- [ ] Search theo patient name/mã BN hoặc visit_id
- [ ] `GET /billing/invoices?patient_id=...&status=ISSUED`
- [ ] Hiển thị: tên bệnh nhân, tổng tiền, trạng thái, ngày khám

**Chi tiết hóa đơn:**
- [ ] Danh sách items: dịch vụ, xét nghiệm, thuốc (tên, số lượng, đơn giá, thành tiền)
- [ ] Subtotal, discount (nếu có), total
- [ ] Lịch sử thanh toán (nếu có partial payment)

**Thu tiền:**
- [ ] Form: Số tiền thu (auto-fill = total), Phương thức (Tiền mặt / Chuyển khoản / Thẻ)
- [ ] Nếu chuyển khoản: input số tham chiếu
- [ ] Nút "Xác nhận thu tiền" → `POST /billing/invoices/:id/pay`
- [ ] Hiển thị tiền thừa (nếu thu > total)

**In biên lai:**
- [ ] Sau khi thanh toán thành công → nút "In biên lai"
- [ ] React-to-print: template biên lai (tên PK, ngày giờ, tên BN, items, tổng, đã thanh toán)

**In hóa đơn PDF:**
- [ ] Nút "Xuất PDF" → `GET /billing/invoices/:id/pdf?download=true`
- [ ] Mở dialog save file hoặc tự download

**Payment History:**
- [ ] Tab riêng: danh sách thanh toán trong ngày/tuần
- [ ] `GET /billing/payments?from_date=...&to_date=...`
- [ ] Tổng kết theo phương thức thanh toán

---

## WEB

### Prerequisite
- `GET /billing/invoices/:id` accessible cho PATIENT role

### Invoice View (bổ sung vào Account page, tab "Hóa đơn")
- [ ] `GET /patients/me/invoices` → danh sách hóa đơn
- [ ] Card hóa đơn: ngày khám, bác sĩ, tổng tiền, trạng thái badge
- [ ] Click → chi tiết hóa đơn
  - Danh sách items
  - Tổng tiền + trạng thái
  - Lịch sử thanh toán
- [ ] Nút "Tải PDF" → `GET /billing/invoices/:id/pdf?download=true`

### SMS Notification UX
- [ ] Sau khi đặt lịch: hiển thị `"SMS xác nhận đã gửi đến {phone_masked}"`
- [ ] Toast notification khi nhận kết quả XN (polling hoặc user refresh)

---

## ĐIỂM KẾT NỐI Sprint 7

| Event | Trigger | Backend Action | Client Effect |
|-------|---------|----------------|---------------|
| `VisitClosed` | Doctor kết thúc khám | Billing worker tạo invoice | Receptionist thấy invoice mới |
| `InvoiceCreated` | Billing worker | Publish Redis Stream event | — |
| `PaymentReceived` | Receptionist thu tiền | Redis Stream event | — |
| `AppointmentScheduled` | Bệnh nhân đặt lịch (Web) | SMS xác nhận | Bệnh nhân nhận SMS |
| `LabResultVerified` | Lab Tech verify | SMS thông báo | Bệnh nhân nhận SMS + xem trên Web |

## DEFINITION OF DONE

- [ ] Invoice tự tạo sau `VisitClosed` event (idempotent — test gửi event 2 lần)
- [ ] Invoice items đúng: phí dịch vụ + XN + thuốc
- [ ] Thanh toán: cập nhật status invoice đúng (PAID / PARTIALLY_PAID)
- [ ] PDF invoice generate thành công (< 5s)
- [ ] PDF cache trong MinIO sau lần đầu
- [ ] Desktop: thu tiền flow hoàn chỉnh, in biên lai OK
- [ ] Web: bệnh nhân xem hóa đơn của mình, tải PDF OK
- [ ] SMS gửi thành công (hoặc log mock) cho: đặt lịch, hủy lịch, kết quả XN
- [ ] Notification retry hoạt động khi gửi fail
