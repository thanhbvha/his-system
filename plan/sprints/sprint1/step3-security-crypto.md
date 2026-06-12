# Sprint 1 — Step 3: Security Package (Crypto)

> **Mục tiêu:** Xây dựng `pkg/crypto` — layer mã hoá AES-GCM và HMAC dùng cho toàn hệ thống.
> **Phụ thuộc:** Step 1 (Go project init xong).
> **Output:** `pkg/crypto` pass 100% unit test. Đây là **điều kiện bắt buộc** trước khi viết bất kỳ module nghiệp vụ nào.

---

> ⚠️ **CRITICAL:** `pkg/crypto` phải xong và pass test 100% trước khi viết bất kỳ module nào khác. Dữ liệu PII (SĐT, CCCD, Email) phải được mã hoá ngay từ đầu.

---

## 1. `pkg/crypto/aes_gcm.go`

```go
// EncryptAESGCM mã hoá plaintext bằng AES-256-GCM.
// Output format: nonce(12B) | ciphertext | tag(16B) → base64 URL-safe
func EncryptAESGCM(plaintext, key []byte) ([]byte, error)

// DecryptAESGCM giải mã ciphertext đã được mã hoá bởi EncryptAESGCM.
func DecryptAESGCM(ciphertext, key []byte) ([]byte, error)
```

**Yêu cầu kỹ thuật:**

- [ ] Nonce: **96-bit random** mỗi lần encrypt (dùng `crypto/rand`)
- [ ] Output format: `nonce(12B) || ciphertext || tag(16B)` → encode base64 URL-safe (no padding)
- [ ] Key: 256-bit (32 bytes), load từ env `FIELD_ENCRYPTION_KEY` (base64 encoded)
- [ ] Trả về lỗi rõ ràng nếu key không đúng kích thước

---

## 2. `pkg/crypto/field_cipher.go`

```go
// FieldCipher là wrapper tiện lợi cho việc mã hoá field-level trong database.
type FieldCipher struct {
    key    []byte // AES-256 key
    hmacKey []byte // HMAC-SHA256 key (có thể dùng chung hoặc key riêng)
}

func NewFieldCipher(encKey, hmacKey []byte) (*FieldCipher, error)

// Encrypt mã hoá một string field, trả về chuỗi base64.
func (f *FieldCipher) Encrypt(plaintext string) (string, error)

// Decrypt giải mã một string field đã mã hoá.
func (f *FieldCipher) Decrypt(ciphertext string) (string, error)

// HMAC tạo HMAC-SHA256 deterministic của value, dùng để tìm kiếm (lookup).
// Trả về hex string.
func (f *FieldCipher) HMAC(value string) string
```

**Use case pattern trong model:**

```go
// Khi lưu patient
patient.PhoneEncrypted, _ = cipher.Encrypt(phone)
patient.PhoneHMAC        = cipher.HMAC(phone)  // dùng để WHERE phone_hmac = ?

// Khi tìm kiếm theo phone
hmac := cipher.HMAC(inputPhone)
db.Where("phone_hmac = ?", hmac).First(&patient)

// Khi hiển thị
phone, _ := cipher.Decrypt(patient.PhoneEncrypted)
```

---

## 3. Unit Tests

File: `pkg/crypto/crypto_test.go`

- [ ] **AES-GCM tests:**
  ```go
  // Round-trip: EncryptAESGCM → DecryptAESGCM trả về plaintext gốc
  TestEncryptDecryptRoundTrip(t *testing.T)
  
  // Mỗi lần encrypt cùng plaintext phải tạo ra ciphertext khác nhau (nonce random)
  TestEncryptProducesUniqueCiphertexts(t *testing.T)
  
  // Key sai kích thước phải trả về error
  TestInvalidKeySize(t *testing.T)
  
  // Ciphertext bị tamper phải trả về error khi decrypt
  TestTamperedCiphertextDetected(t *testing.T)
  ```

- [ ] **FieldCipher tests:**
  ```go
  // HMAC của cùng input luôn trả về cùng output (deterministic)
  TestHMACIsDeterministic(t *testing.T)
  
  // Round-trip Encrypt/Decrypt
  TestFieldCipherRoundTrip(t *testing.T)
  
  // Empty string handling
  TestEncryptEmptyString(t *testing.T)
  ```

- [ ] Chạy test: `go test ./pkg/crypto/... -v -cover` → **coverage ≥ 95%**

---

## 4. Config & Key Management

- [ ] Load key từ environment trong `pkg/crypto/config.go`:
  ```go
  func LoadFieldCipherFromEnv() (*FieldCipher, error) {
      encKey  := os.Getenv("FIELD_ENCRYPTION_KEY")  // base64
      hmacKey := os.Getenv("FIELD_HMAC_KEY")        // base64
      // decode và validate length
  }
  ```
- [ ] Tạo helper script để generate key: `scripts/gen-crypto-keys.sh`
  ```bash
  #!/bin/bash
  echo "FIELD_ENCRYPTION_KEY=$(openssl rand -base64 32)"
  echo "FIELD_HMAC_KEY=$(openssl rand -base64 32)"
  ```

---

## Definition of Done (Step 3)

- [ ] `go test ./pkg/crypto/... -v` → **tất cả test PASS**
- [ ] Coverage ≥ 95%
- [ ] Encrypt/decrypt round-trip đúng
- [ ] HMAC deterministic (cùng input → cùng output)
- [ ] Ciphertext bị tamper → decrypt trả về error (không panic)
- [ ] Key sai size → error rõ ràng
