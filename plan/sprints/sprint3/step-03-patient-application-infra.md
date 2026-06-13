# Sprint 3 — Step 3: Patient Application Layer + Infrastructure (Backend)

## Mục tiêu
Implement Use Cases (Commands/Queries) và Repository Postgres cho Patient. Sau step này, toàn bộ business logic Patient hoạt động được và có thể test qua unit test.

## Cây thư mục cần tạo
```
internal/patient/
├── application/
│   ├── command/
│   │   ├── create_patient.go
│   │   ├── update_patient.go
│   │   └── update_insurance.go
│   └── query/
│       ├── search_patients.go
│       ├── get_patient_by_id.go
│       └── get_patient_history.go
└── infrastructure/
    └── patient_repository_pg.go
```

---

## Nhiệm vụ chi tiết

### 1. `CreatePatientCommand`
**Input:** FullName, DOB, Gender, Phone (plaintext), CCCD (optional), Email (optional), Address (optional)

**Logic:**
1. Validate: Phone 10 số, CCCD 12 số (nếu có).
2. Dùng `domain.NewPhoneNumber(phone, cipher)` → sinh `PhoneEncrypted`, `PhoneHMAC`.
3. Check unique: `patientRepo.GetByPhoneHMAC(phoneHMAC)` → nếu đã có → return error `PHONE_ALREADY_EXISTS`.
4. Tạo `domain.Patient`, gọi `patientRepo.Create`.
5. Publish Redis Stream event `HIS.PATIENT.PatientRegistered`.

### 2. `UpdatePatientCommand`
**Logic:**
1. Lấy patient hiện tại.
2. Nếu Phone thay đổi → validate lại + re-encrypt + cập nhật HMAC.
3. Gọi `patientRepo.Update`.
4. Publish `HIS.PATIENT.PatientUpdated`.

### 3. `UpdateInsuranceCommand`
- Upsert `PatientInsurance` (dùng `ON CONFLICT` ở DB level).

### 4. `SearchPatients` Query
```go
type SearchPatientsQuery struct {
    Phone  string  // sẽ HMAC internally nếu có
    CCCD   string  // sẽ HMAC internally nếu có
    Name   string  // full-text search nếu không có phone/cccd
    Page   int
    Limit  int
}
```
**Priority:** Phone > CCCD > Name (full-text)

**Response DTO (`PatientListItem`):** Luôn trả `phone_masked` — KHÔNG BAO GIỜ trả plaintext trong list.
```go
type PatientListItem struct {
    ID           string `json:"id"`
    FullName     string `json:"full_name"`
    DOB          string `json:"dob"`
    Gender       string `json:"gender"`
    PhoneMasked  string `json:"phone_masked"` // "09x***xxx"
    PatientCode  string `json:"patient_code"` // prefix ID
}
```

### 5. `GetPatientByID` Query
- **Nếu caller là Staff** (RECEPTIONIST/DOCTOR/NURSE): Decrypt PII và trả đầy đủ.
- **Nếu caller là PATIENT**: Trả masked PII (chỉ trả record của chính họ).

### 6. `PatientRepositoryPG` (`infrastructure/patient_repository_pg.go`)
- Inject `*pgxpool.Pool` và `*crypto.FieldCipher`.
- **Tự động encrypt** PII trước khi ghi DB.
- **Tự động decrypt** PII sau khi đọc DB.
- Implement tất cả methods của `domain.PatientRepository`.

**Search by name (SQL snippet):**
```sql
SELECT ... FROM patients
WHERE full_name_search @@ plainto_tsquery('simple', $1)
ORDER BY ts_rank(full_name_search, plainto_tsquery('simple', $1)) DESC
LIMIT $2 OFFSET $3
```

---

## Masked Phone Logic
```go
func MaskPhone(plain string) string {
    // "0912345678" → "091***678"
    if len(plain) < 7 { return "***" }
    return plain[:3] + "***" + plain[len(plain)-3:]
}
```
> ⚠️ Hàm mask chỉ chạy SAU khi decrypt. Không bao giờ mask từ ciphertext.
