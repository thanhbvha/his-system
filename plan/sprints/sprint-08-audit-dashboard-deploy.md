# Sprint 8 — Audit, Dashboard, Polish & Deploy (Tuần 15–16)

> **Mục tiêu:** Audit trail, Admin dashboard báo cáo, hoàn thiện UX, production deploy.
> **Prerequisite:** Sprint 7 hoàn thành. Toàn bộ business flow từ Sprint 1–7 hoạt động.
> **Kết thúc sprint:** Hệ thống production-ready, deploy VPS, monitoring hoạt động.

---

## BACKEND

### Module `internal/audit` (NATS Worker)

**Audit Worker:**
- [ ] Subscribe **TẤT CẢ** NATS subjects: `HIS.>`
- [ ] Mỗi event → ghi vào MongoDB `audit_logs`:
  ```js
  {
    _id,
    user_id, user_role,
    action,             // "PatientCreated", "LabResultVerified", ...
    entity,             // "Patient", "LabOrder", ...
    entity_id,
    diff: {},           // before/after nếu là update event
    ip_address,
    device_fingerprint,
    timestamp
  }
  ```
- [ ] PII masking: không ghi SĐT/CCCD plaintext vào audit log

> ⚠️ **NOTE:** Audit log là immutable — không có API DELETE/UPDATE audit records.
> Collection MongoDB phải có TTL index (lưu 5 năm theo quy định y tế).

### Reporting APIs

- [ ] `GET /api/v1/reports/revenue`
  - query: `from_date`, `to_date`, `group_by` (day/week/month)
  - response: `[{ period, total_revenue, total_invoices, total_paid }]`
  - Source: PostgreSQL `payments` table

- [ ] `GET /api/v1/reports/patients`
  - query: `from_date`, `to_date`
  - response: `{ total_new, total_returning, by_gender, by_age_group }`

- [ ] `GET /api/v1/reports/services`
  - response: top 10 dịch vụ theo lượt khám + doanh thu

- [ ] `GET /api/v1/reports/visits`
  - query: `from_date`, `to_date`, `doctor_id?`
  - response: `{ total_visits, avg_per_day, by_status }`

**Report Snapshot Worker:**
- [ ] Subscribe `HIS.BILLING.PaymentReceived` + `HIS.VISIT.VisitClosed`
- [ ] Cập nhật daily snapshot trong MongoDB `report_snapshots` (upsert by date)
- [ ] Dashboard query từ snapshots thay vì real-time aggregate (performance)

### APIs — Audit

```
GET /api/v1/audit/logs              [ADMIN]
  query: user_id, entity, action, from_date, to_date, page, limit
  → Query MongoDB audit_logs với filter

GET /api/v1/audit/logs/:id          [ADMIN]
  → Chi tiết 1 audit entry (có diff)
```

### Integration & Load Testing
- [ ] Integration test suite (`testify` + `httptest`):
  - Auth flow end-to-end
  - Patient create → appointment → checkin → visit → EMR → billing
  - Concurrent booking (test double-booking prevention)
  - Concurrent stock deduction (test race condition)
- [ ] Load test với `k6`:
  - Target: 100 concurrent users
  - Test: GET /patients/search, POST /appointments, GET /queue/ws

### API Documentation
- [ ] Swagger hoàn chỉnh tất cả endpoints (Sprint 1–8)
- [ ] Mỗi endpoint có: description, request body schema, response schema, error codes
- [ ] Swagger UI accessible tại `/swagger/`

### Security Hardening
- [ ] Security headers middleware: `X-Content-Type-Options`, `X-Frame-Options`, `Strict-Transport-Security`
- [ ] Request body size limit (10MB default, 50MB cho file upload)
- [ ] SQL injection prevention: review tất cả raw queries (dùng parameterized)
- [ ] Log rotation config

---

## DESKTOP

### Prerequisite
- Reporting APIs và Audit log API phải ready

### Admin Dashboard

**Revenue Chart (Recharts):**
- [ ] Line chart: doanh thu 30 ngày gần nhất
- [ ] `GET /reports/revenue?from_date=...&to_date=...&group_by=day`
- [ ] Bar chart: top 10 dịch vụ doanh thu
- [ ] Date range picker (Ant Design RangePicker)

**Patient Stats:**
- [ ] Card KPIs: tổng BN mới tháng này, tổng lượt khám, trung bình/ngày
- [ ] Pie chart: theo giới tính, theo nhóm tuổi
- [ ] `GET /reports/patients?from_date=...&to_date=...`

**Visit Summary:**
- [ ] `GET /reports/visits?from_date=...&to_date=...`
- [ ] By status breakdown: completed vs cancelled

### Audit Log Viewer (Admin)

- [ ] Table với filter: User, Action type, Entity, Date range
- [ ] `GET /audit/logs` với pagination
- [ ] Click row → expand: xem full diff (before/after)
- [ ] Highlight sensitive actions: login_failed, patient_accessed, drug_dispensed
- [ ] Export CSV (client-side, từ data đã load)

### System Config UI (Admin)

- [ ] Clinic info: tên, địa chỉ, SĐT, logo (upload MinIO)
- [ ] Working hours: giờ mở cửa, ngày nghỉ
- [ ] Service management: thêm/sửa/xóa dịch vụ + giá
- [ ] Department management

### Polish toàn bộ Desktop

- [ ] Loading states: Skeleton cho tất cả data-fetching
- [ ] Error boundary: fallback UI khi component crash
- [ ] Empty states: tất cả list/table
- [ ] Confirm dialogs: tất cả destructive actions (delete, deactivate, reject)
- [ ] Keyboard shortcuts: F5 refresh worklist, Enter submit form
- [ ] Print preview trước khi in (React-to-print preview)

### Wails Production Build
- [ ] `wails build -platform windows/amd64`
- [ ] Code signing (Windows) nếu có certificate
- [ ] Auto-updater config (Wails built-in)
- [ ] Installer (NSIS hoặc WiX)

---

## WEB

### Polish toàn bộ Web

**UX Improvements:**
- [ ] Loading skeleton cho tất cả pages
- [ ] Error boundary + generic error page
- [ ] 404 page (với nút về trang chủ)
- [ ] 403 page (unauthorized)
- [ ] Toast notifications: thành công, lỗi, warning

**Responsive:**
- [ ] Test trên mobile 375px, tablet 768px, desktop 1280px
- [ ] Booking flow hoạt động tốt trên mobile
- [ ] My appointments card responsive
- [ ] Lab results table scroll horizontal trên mobile

**SEO:**
- [ ] `<title>` unique cho mỗi page
- [ ] `<meta name="description">` cho landing, booking pages
- [ ] Open Graph tags (og:title, og:description, og:image)
- [ ] Canonical URL
- [ ] robots.txt + sitemap.xml

**Performance:**
- [ ] Route-based code splitting (`React.lazy + Suspense`)
- [ ] Image optimization: WebP, lazy loading
- [ ] Bundle size target: < 500KB initial JS
- [ ] Lighthouse score > 80 (Performance, Accessibility)

**Accessibility:**
- [ ] Tất cả interactive elements có aria-label
- [ ] Form validation errors announce cho screen readers
- [ ] Color contrast đạt WCAG AA

### E2E Testing (Playwright hoặc Cypress)
- [ ] Test: Register → Login → Book appointment → View appointment
- [ ] Test: Login → View lab results (mock verified result)
- [ ] Test: Login → View invoice

---

## DEVOPS — Production Deploy

### Docker Compose Production
- [ ] `docker-compose.prod.yml`:
  - Không có hot-reload (air)
  - Go binary build (multi-stage Dockerfile)
  - Resource limits cho từng service
  - Restart policy: `unless-stopped`
- [ ] `.env.production` template (với placeholder, không commit secrets)

### Nginx Config
- [ ] Reverse proxy: `/api/` → Go Fiber, `/` → React web (static files)
- [ ] SSL termination (Let's Encrypt via Certbot)
- [ ] HTTP → HTTPS redirect
- [ ] Gzip compression
- [ ] Security headers: HSTS, X-Frame-Options, CSP

### Backup Script
- [ ] `scripts/backup.sh`:
  - `pg_dump` → compress → upload S3/Backblaze
  - `mongodump` → compress → upload
  - MinIO bucket sync → remote
  - Retention: 30 ngày
- [ ] Cron job: `0 2 * * *` (2h sáng hàng ngày)

### Monitoring Setup
- [ ] Prometheus scrape config cho Go metrics
- [ ] Grafana dashboards:
  - API: request rate, latency p50/p95/p99, error rate
  - Infrastructure: CPU, RAM, disk
  - Business: visits/day, invoices/day, queue length
- [ ] Grafana alerts → Telegram bot:
  - API error rate > 5%
  - Response time p95 > 2s
  - Disk usage > 80%
  - Service down (health check fail)

### CI/CD — GitHub Actions hoàn chỉnh
- [ ] `ci.yml`: lint → unit test → integration test → build check (PR)
- [ ] `deploy.yml`: build image → push GHCR → SSH to VPS → `docker compose pull && up -d` (tag push)
- [ ] Secrets: VPS_SSH_KEY, GHCR_TOKEN, .env values

---

## CHECKLIST CUỐI SPRINT 8 (Production Readiness)

### Functional
- [ ] Toàn bộ flow: Đặt lịch → Check-in → Khám → XN → Kê đơn → Xuất thuốc → Thanh toán
- [ ] Tất cả role login và dùng được chức năng đúng quyền
- [ ] Notification SMS gửi đúng triggers
- [ ] Audit log ghi đầy đủ mọi action

### Security
- [ ] JWT AES-GCM hoạt động đúng
- [ ] PII encrypt/decrypt đúng, không lộ plaintext trong log
- [ ] RBAC block đúng unauthorized access
- [ ] HTTPS trên production
- [ ] Rate limiting auth endpoints

### Performance
- [ ] API response time p95 < 500ms (trừ PDF generate)
- [ ] WebSocket queue update < 100ms
- [ ] Web Lighthouse score > 80
- [ ] k6 load test: 100 users, không lỗi

### Operations
- [ ] `docker-compose up` production OK
- [ ] Backup script chạy và upload thành công
- [ ] Grafana dashboards hiển thị metrics
- [ ] Alert Telegram khi service down

## DEFINITION OF DONE

- [ ] Tất cả integration tests pass
- [ ] k6 load test pass (100 concurrent users, error rate < 1%)
- [ ] Swagger documentation hoàn chỉnh
- [ ] Production deploy thành công trên VPS
- [ ] Backup script chạy được
- [ ] Monitoring dashboards active
- [ ] Alert notifications hoạt động
- [ ] Admin dashboard hiển thị đúng số liệu
- [ ] Audit log ghi đầy đủ
- [ ] Web Lighthouse > 80
- [ ] Wails desktop build installer thành công
