# Sprint 3 — Step 4: Patient HTTP API (Backend)

## Mục tiêu
Expose toàn bộ Patient Use Cases qua HTTP. Tất cả routes được bảo vệ bởi `JWTAuth` + `RequireRole` middleware từ Sprint 2.

## File cần tạo
```
internal/api/patient/
└── patient_handler.go
```

Đăng ký routes trong `cmd/api/main.go`.

---

## Routes & Middleware

```
[JWTAuth]
├── GET    /api/v1/patients              → ListPatients (search)
│          [RequireRole: RECEPTIONIST, DOCTOR, NURSE, LAB_TECH, PHARMACIST]
├── POST   /api/v1/patients              → CreatePatient
│          [RequireRole: RECEPTIONIST, ADMIN]
│          [RequestSignature — Desktop only]
├── GET    /api/v1/patients/:id          → GetPatientByID (full PII — staff)
│          [RequireRole: RECEPTIONIST, DOCTOR, NURSE]
├── PUT    /api/v1/patients/:id          → UpdatePatient
│          [RequireRole: RECEPTIONIST, ADMIN]
│          [RequestSignature — Desktop only]
│
├── GET    /api/v1/patients/me           → GetMyProfile (masked PII — patient)
│          [RequireRole: PATIENT]
├── PUT    /api/v1/patients/me           → UpdateMyProfile
│          [RequireRole: PATIENT]
│
├── GET    /api/v1/patients/:id/insurance
│          [RequireRole: RECEPTIONIST, DOCTOR, NURSE]
├── GET    /api/v1/patients/me/insurance
│          [RequireRole: PATIENT]
└── PUT    /api/v1/patients/me/insurance
           [RequireRole: PATIENT]
```

---

## Request/Response DTOs

### `POST /api/v1/patients`
```json
// Request
{
    "full_name": "Nguyễn Văn A",
    "dob": "1990-05-15",
    "gender": "MALE",
    "phone": "0912345678",
    "cccd": "012345678901",
    "email": "a@example.com",
    "address": "123 Đường ABC, Q1, HCM"
}

// Response 201
{
    "success": true,
    "data": {
        "id": "uuid",
        "full_name": "Nguyễn Văn A",
        "phone_masked": "091***678",
        "patient_code": "BN-xxxxxxxx"
    }
}
```

### `GET /api/v1/patients?q=Nguyen&page=1&limit=20`
```json
{
    "success": true,
    "data": {
        "items": [
            {
                "id": "uuid",
                "full_name": "Nguyễn Văn A",
                "dob": "1990-05-15",
                "gender": "MALE",
                "phone_masked": "091***678"
            }
        ],
        "total": 1,
        "page": 1,
        "limit": 20
    }
}
```

### `GET /api/v1/patients/:id` (Staff — full PII)
```json
{
    "success": true,
    "data": {
        "id": "uuid",
        "full_name": "Nguyễn Văn A",
        "phone": "0912345678",     // decrypted, chỉ staff thấy
        "cccd": "012345678901",    // decrypted
        "email": "a@example.com",
        "dob": "1990-05-15",
        "blood_type": "A+",
        "insurance": { ... }
    }
}
```

---

## Lưu ý quan trọng
> ✅ Handler `/patients/:id` vs `/patients/me` phải phân biệt rõ qua JWT role.
> ✅ Route `/patients/me` phải được đăng ký TRƯỚC `/patients/:id` để Fiber không nhầm `me` là `:id`.
> ❌ Không bao giờ gọi decrypt trong `/patients` (list) — chỉ mask.
