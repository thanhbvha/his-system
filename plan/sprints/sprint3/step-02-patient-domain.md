# Sprint 3 — Step 2: Patient Domain Layer (Backend)

## Mục tiêu
Implement module `internal/patient` — Domain layer: Entity, Value Objects, Repository interface. Đây là nền tảng để tất cả các lớp trên (Application, Infrastructure, API) có thể build lên.

## Cây thư mục cần tạo
```
internal/patient/
├── domain/
│   ├── patient.go           -- Entity Patient, PatientInsurance, PatientContact
│   ├── value_objects.go     -- PhoneNumber, CCCD, BHYTNumber
│   └── repository.go        -- Interface PatientRepository
```

---

## Nhiệm vụ chi tiết

### 1. Entity `Patient` (`domain/patient.go`)
```go
type Patient struct {
    ID                      uuid.UUID
    FullName                string
    DOB                     time.Time
    Gender                  string      // "MALE" | "FEMALE" | "OTHER"
    BloodType               string
    IsActive                bool

    // PII — chứa dạng mã hóa, không bao giờ plaintext
    PhoneEncrypted          string
    PhoneHMAC               string
    CCCDEncrypted           string
    CCCDHMAC                string
    EmailEncrypted          string
    EmailHMAC               string
    AddressDetailEncrypted  string

    AvatarURL               string
    CreatedAt               time.Time
    UpdatedAt               time.Time
}
```

### 2. Value Objects (`domain/value_objects.go`)
```go
// PhoneNumber — validate 10 chữ số VN, tự sinh Encrypted + HMAC
type PhoneNumber struct {
    Encrypted string
    HMAC      string
}
func NewPhoneNumber(plaintext string, cipher *crypto.FieldCipher) (PhoneNumber, error)

// CCCD — validate 12 chữ số
type CCCD struct {
    Encrypted string
    HMAC      string
}
func NewCCCD(plaintext string, cipher *crypto.FieldCipher) (CCCD, error)

// BHYTNumber — validate format 15 ký tự
type BHYTNumber struct {
    Encrypted string
    HMAC      string
}
func NewBHYTNumber(plaintext string, cipher *crypto.FieldCipher) (BHYTNumber, error)
```

**Validate rules:**
- `PhoneNumber`: regex `^(0[3|5|7|8|9])[0-9]{8}$`
- `CCCD`: regex `^[0-9]{12}$`
- `BHYTNumber`: regex `^[A-Z]{2}[0-9]{13}$` (tham khảo chuẩn BYT)

### 3. Entity `PatientInsurance` & `PatientContact`
```go
type PatientInsurance struct {
    ID                   uuid.UUID
    PatientID            uuid.UUID
    BHYTNumberEncrypted  string
    BHYTNumberHMAC       string
    ValidFrom            time.Time
    ValidTo              time.Time
    CoverageLevel        string
    IssuingProvince      string
}

type PatientContact struct {
    ID              uuid.UUID
    PatientID       uuid.UUID
    Name            string
    Relationship    string
    PhoneEncrypted  string
    PhoneHMAC       string
    IsPrimary       bool
}
```

### 4. Repository Interface (`domain/repository.go`)
```go
type PatientRepository interface {
    Create(ctx context.Context, patient *Patient) error
    GetByID(ctx context.Context, id uuid.UUID) (*Patient, error)
    GetByPhoneHMAC(ctx context.Context, phoneHMAC string) (*Patient, error)
    GetByCCCDHMAC(ctx context.Context, cccdHMAC string) (*Patient, error)
    SearchByName(ctx context.Context, query string, page, limit int) ([]*Patient, int64, error)
    Update(ctx context.Context, patient *Patient) error
    List(ctx context.Context, page, limit int) ([]*Patient, int64, error)

    // Insurance
    UpsertInsurance(ctx context.Context, ins *PatientInsurance) error
    GetInsurance(ctx context.Context, patientID uuid.UUID) (*PatientInsurance, error)
}
```

---

## Quy tắc quan trọng
> ❌ Domain layer **không import** bất kỳ thư viện infrastructure nào (pgx, redis...).
> ✅ Domain chỉ phụ thuộc vào `pkg/crypto` để mã hóa trong Value Objects.
> ✅ Mọi struct trong Domain đều dùng kiểu Go thuần — không có struct tag JSON/DB.
