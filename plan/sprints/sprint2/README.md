# Sprint 2 — README

> **Sprint:** 2 — Identity & Auth (Tuần 3–4)
> **Mục tiêu:** Hoàn thiện hệ thống xác thực: JWT AES-GCM, RBAC, TOTP MFA cho Desktop; OTP SĐT cho Web.
> **Tài liệu tổng quan:** Xem `plan/sprints/sprint-02-identity-auth.md`
> **Nền tảng:** Xem `plan/report_sprint1.md`

---

## Thứ tự thực hiện

```
Step 1 → Step 2 → Step 3 → Step 4 → Step 5 → Step 6 → Step 7
```

> ⚠️ **Bắt buộc:** Step 1 → Step 2 và Step 3 phải hoàn thành trước Step 4, 5, 6, 7.
> Step 2 và Step 3 có thể làm **song song** vì chúng độc lập nhau.
> Step 5 phụ thuộc Step 2. Step 6 phụ thuộc Step 3.

```
Step 1 (Domain + JWT)
  ├── Step 2 (Desktop Auth API)  →  Step 5 (Desktop Frontend)  →  Step 7 (Admin UI)
  └── Step 3 (Web Auth API)     →  Step 6 (Web Frontend)
         ↓
      Step 4 (RBAC + Middleware + Admin APIs)  [cần Step 1 + Step 2]
```

---

## Tóm tắt các Step

| Step | File | Tầng | Mô tả |
|------|------|------|-------|
| **Step 1** | `step1-identity-domain-jwt.md` | Backend | Identity Domain (Entity, Repository interface, Value Object) + `pkg/auth/jwt.go` với AES-GCM encrypted payload |
| **Step 2** | `step2-desktop-auth-api.md` | Backend | Challenge-response login flow, TOTP MFA, refresh token rotation — API cho Desktop |
| **Step 3** | `step3-web-auth-api.md` | Backend | OTP flow bệnh nhân (Zalo ZNS + SMS fallback), register, HttpOnly cookie — API cho Web |
| **Step 4** | `step4-rbac-middleware-admin.md` | Backend | JWT Auth middleware, RBAC middleware, Request Signature middleware, Admin CRUD APIs |
| **Step 5** | `step5-desktop-frontend-auth.md` | Desktop | TPM/Keychain hardware key, signature interceptor, Login/MFA/MFA Setup screens |
| **Step 6** | `step6-web-frontend-auth.md` | Web | Cookie refresh interceptor, OTPInput component, LoginPage OTP flow, RegisterPage |
| **Step 7** | `step7-admin-ui-desktop.md` | Desktop | Admin UI: User list/create/deactivate, Role-Permission matrix, Department management |

---

## Điểm kết nối quan trọng giữa các tầng

| Vấn đề | Backend (Source) | Desktop (Consumer) | Web (Consumer) |
|--------|-----------------|-------------------|----------------|
| JWT format | AES-GCM encrypted payload + `cnf.jkt` claim | Không decode, chỉ forward trong `Authorization` header | Không decode, chỉ forward |
| Hardware Binding | Verify `cnf.jkt` = SHA256(PublicKeyPEM) trong Signature Middleware | TPM sinh key pair, ký mọi request | Không áp dụng |
| Refresh Token | Desktop: opaque random string trong Redis TTL 7d | Body: `{ refresh_token, signature, public_key_pem }` | HttpOnly Cookie, TTL 7d |
| CORS | `AllowCredentials: true` cho Web origins | — | `withCredentials: true` trong apiClient |
| OTP Delivery | Queue worker: Zalo ZNS → SMS fallback | — | UI hiển thị countdown + Gửi lại |
