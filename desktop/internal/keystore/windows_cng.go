//go:build windows

package keystore

import (
	"crypto/sha256"
	"encoding/pem"
	"fmt"
	"syscall"
	"unsafe"
)

// Windows CNG bindings
var (
	ncrypt = syscall.NewLazyDLL("ncrypt.dll")

	procNCryptOpenStorageProvider = ncrypt.NewProc("NCryptOpenStorageProvider")
	procNCryptCreatePersistedKey  = ncrypt.NewProc("NCryptCreatePersistedKey")
	procNCryptFinalizeKey         = ncrypt.NewProc("NCryptFinalizeKey")
	procNCryptSignHash            = ncrypt.NewProc("NCryptSignHash")
	procNCryptOpenKey             = ncrypt.NewProc("NCryptOpenKey")
	procNCryptExportKey           = ncrypt.NewProc("NCryptExportKey")
	procNCryptFreeObject          = ncrypt.NewProc("NCryptFreeObject")
)

const (
	MS_PLATFORM_CRYPTO_PROVIDER = "Microsoft Platform Crypto Provider" // TPM provider
	NCRYPT_ECDSA_P256_ALGORITHM = "ECDSA_P256"
	NCRYPT_MACHINE_KEY_FLAG     = 0x00000020
	NCRYPT_SILENT_FLAG          = 0x00000040
	BCRYPT_ECCPUBLIC_BLOB       = "ECCPUBLICBLOB"
	NCRYPT_PAD_PKCS1_FLAG       = 0x00000002
)

type WindowsCNGKeyStore struct {
	keyName string
}

func NewWindowsCNGKeyStore() *WindowsCNGKeyStore {
	return &WindowsCNGKeyStore{
		keyName: "HIS_Hardware_Key",
	}
}

func (s *WindowsCNGKeyStore) GetOrCreate() (*KeyPair, error) {
	// 1. Open Provider
	var provHandle uintptr
	provName, _ := syscall.UTF16PtrFromString(MS_PLATFORM_CRYPTO_PROVIDER)
	r1, _, err := procNCryptOpenStorageProvider.Call(uintptr(unsafe.Pointer(&provHandle)), uintptr(unsafe.Pointer(provName)), 0)
	if r1 != 0 {
		return nil, fmt.Errorf("NCryptOpenStorageProvider failed: %x, %v", r1, err)
	}
	defer procNCryptFreeObject.Call(provHandle)

	// 2. Try to open existing key
	var keyHandle uintptr
	kName, _ := syscall.UTF16PtrFromString(s.keyName)
	r1, _, _ = procNCryptOpenKey.Call(provHandle, uintptr(unsafe.Pointer(&keyHandle)), uintptr(unsafe.Pointer(kName)), 0, 0)
	if r1 != 0 {
		// Key not found, create new
		alg, _ := syscall.UTF16PtrFromString(NCRYPT_ECDSA_P256_ALGORITHM)
		r1, _, err = procNCryptCreatePersistedKey.Call(provHandle, uintptr(unsafe.Pointer(&keyHandle)), uintptr(unsafe.Pointer(alg)), uintptr(unsafe.Pointer(kName)), 0, 0)
		if r1 != 0 {
			return nil, fmt.Errorf("NCryptCreatePersistedKey failed: %x, %v", r1, err)
		}

		// Finalize key to generate it
		r1, _, err = procNCryptFinalizeKey.Call(keyHandle, 0)
		if r1 != 0 {
			procNCryptFreeObject.Call(keyHandle)
			return nil, fmt.Errorf("NCryptFinalizeKey failed: %x", r1)
		}
	}
	defer procNCryptFreeObject.Call(keyHandle)

	// 3. Export Public Key
	blobType, _ := syscall.UTF16PtrFromString(BCRYPT_ECCPUBLIC_BLOB)
	var bytesCopied uint32
	r1, _, err = procNCryptExportKey.Call(keyHandle, 0, uintptr(unsafe.Pointer(blobType)), 0, 0, 0, uintptr(unsafe.Pointer(&bytesCopied)), 0)
	if r1 != 0 || bytesCopied == 0 {
		return nil, fmt.Errorf("NCryptExportKey size failed: %x", r1)
	}

	pubBlob := make([]byte, bytesCopied)
	r1, _, err = procNCryptExportKey.Call(keyHandle, 0, uintptr(unsafe.Pointer(blobType)), 0, uintptr(unsafe.Pointer(&pubBlob[0])), uintptr(bytesCopied), uintptr(unsafe.Pointer(&bytesCopied)), 0)
	if r1 != 0 {
		return nil, fmt.Errorf("NCryptExportKey failed: %x", r1)
	}

	// 4. Convert BCRYPT_ECCPUBLIC_BLOB to X.509 SubjectPublicKeyInfo (PEM)
	// BCRYPT_ECCPUBLIC_BLOB structure:
	// Magic (4 bytes) "ECS1"
	// cbKey (4 bytes) length of key (32)
	// X (32 bytes)
	// Y (32 bytes)
	if len(pubBlob) < 8 {
		return nil, fmt.Errorf("invalid public blob size")
	}
	cbKey := *(*uint32)(unsafe.Pointer(&pubBlob[4]))
	if len(pubBlob) != int(8+2*cbKey) {
		return nil, fmt.Errorf("invalid public blob size match")
	}

	x := pubBlob[8 : 8+cbKey]
	y := pubBlob[8+cbKey : 8+2*cbKey]

	// Manual ASN.1 encoding for ECDSA P-256 SubjectPublicKeyInfo
	// This is a fixed prefix for secp256r1
	prefix := []byte{0x30, 0x59, 0x30, 0x13, 0x06, 0x07, 0x2a, 0x86, 0x48, 0xce, 0x3d, 0x02, 0x01, 0x06, 0x08, 0x2a, 0x86, 0x48, 0xce, 0x3d, 0x03, 0x01, 0x07, 0x03, 0x42, 0x00, 0x04}
	der := append(prefix, x...)
	der = append(der, y...)

	pubPem := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: der})

	return &KeyPair{
		PublicKeyPEM: string(pubPem),
	}, nil
}

func (s *WindowsCNGKeyStore) Sign(data []byte) ([]byte, error) {
	var provHandle uintptr
	provName, _ := syscall.UTF16PtrFromString(MS_PLATFORM_CRYPTO_PROVIDER)
	r1, _, _ := procNCryptOpenStorageProvider.Call(uintptr(unsafe.Pointer(&provHandle)), uintptr(unsafe.Pointer(provName)), 0)
	if r1 != 0 {
		return nil, fmt.Errorf("NCryptOpenStorageProvider failed: %x", r1)
	}
	defer procNCryptFreeObject.Call(provHandle)

	var keyHandle uintptr
	kName, _ := syscall.UTF16PtrFromString(s.keyName)
	r1, _, _ = procNCryptOpenKey.Call(provHandle, uintptr(unsafe.Pointer(&keyHandle)), uintptr(unsafe.Pointer(kName)), 0, 0)
	if r1 != 0 {
		return nil, fmt.Errorf("Key not found")
	}
	defer procNCryptFreeObject.Call(keyHandle)

	hash := sha256.Sum256(data)
	var bytesCopied uint32

	r1, _, _ = procNCryptSignHash.Call(keyHandle, 0, uintptr(unsafe.Pointer(&hash[0])), uintptr(len(hash)), 0, 0, uintptr(unsafe.Pointer(&bytesCopied)), 0)
	if r1 != 0 || bytesCopied == 0 {
		return nil, fmt.Errorf("NCryptSignHash size failed: %x", r1)
	}

	sigBlob := make([]byte, bytesCopied)
	r1, _, _ = procNCryptSignHash.Call(keyHandle, 0, uintptr(unsafe.Pointer(&hash[0])), uintptr(len(hash)), uintptr(unsafe.Pointer(&sigBlob[0])), uintptr(bytesCopied), uintptr(unsafe.Pointer(&bytesCopied)), 0)
	if r1 != 0 {
		return nil, fmt.Errorf("NCryptSignHash failed: %x", r1)
	}

	// CNG returns raw r || s. We need to convert it to ASN.1
	if len(sigBlob) != 64 {
		return nil, fmt.Errorf("invalid signature size")
	}
	r := sigBlob[:32]
	s_sig := sigBlob[32:]

	// Simple ASN.1 encoding (might need padding for MSB=1)
	asn1Sig := encodeASN1(r, s_sig)
	return asn1Sig, nil
}

func encodeASN1(r, s []byte) []byte {
	// Strip leading zeros
	for len(r) > 1 && r[0] == 0 {
		r = r[1:]
	}
	// Add leading zero if MSB is 1
	if r[0]&0x80 != 0 {
		r = append([]byte{0}, r...)
	}

	for len(s) > 1 && s[0] == 0 {
		s = s[1:]
	}
	if s[0]&0x80 != 0 {
		s = append([]byte{0}, s...)
	}

	res := make([]byte, 0, 6+len(r)+len(s))
	res = append(res, 0x30, byte(4+len(r)+len(s)))
	res = append(res, 0x02, byte(len(r)))
	res = append(res, r...)
	res = append(res, 0x02, byte(len(s)))
	res = append(res, s...)
	return res
}

func New() KeyStore {
	return NewWindowsCNGKeyStore()
}
