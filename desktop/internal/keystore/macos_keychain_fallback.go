//go:build darwin && !cgo

package keystore

import "fmt"

type MacOSFallbackKeyStore struct {
	fallback *SoftwareKeyStore
}

func NewMacOSFallbackKeyStore() *MacOSFallbackKeyStore {
	return &MacOSFallbackKeyStore{
		fallback: NewSoftwareKeyStore(),
	}
}

func (s *MacOSFallbackKeyStore) GetOrCreate() (*KeyPair, error) {
	fmt.Println("[WARN] macOS Secure Enclave is disabled because CGO is not available. Falling back to software keystore.")
	return s.fallback.GetOrCreate()
}

func (s *MacOSFallbackKeyStore) Sign(data []byte) ([]byte, error) {
	return s.fallback.Sign(data)
}

func New() KeyStore {
	return NewMacOSFallbackKeyStore()
}
