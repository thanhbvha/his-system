//go:build linux

package keystore

import (
	"fmt"
)

type LinuxTPMKeyStore struct {
	fallback *SoftwareKeyStore
}

func NewLinuxTPMKeyStore() *LinuxTPMKeyStore {
	return &LinuxTPMKeyStore{
		fallback: NewSoftwareKeyStore(),
	}
}

func (s *LinuxTPMKeyStore) GetOrCreate() (*KeyPair, error) {
	// TODO: Tương tác với phần cứng TPM 2.0 trên Linux (ví dụ thông qua /dev/tpmrm0)
	// Sử dụng package "github.com/google/go-tpm/tpm2"
	//
	// Hiện tại cho MVP trên Ubuntu, chúng ta sẽ tạm fallback về file-based (giống như chưa cắm chip TPM).
	fmt.Println("[WARN] Linux TPM 2.0 is not fully implemented yet. Falling back to software keystore.")
	return s.fallback.GetOrCreate()
}

func (s *LinuxTPMKeyStore) Sign(data []byte) ([]byte, error) {
	// TODO: Gửi lệnh TPM2_Sign xuống phần cứng
	return s.fallback.Sign(data)
}

func New() KeyStore {
	return NewLinuxTPMKeyStore()
}
