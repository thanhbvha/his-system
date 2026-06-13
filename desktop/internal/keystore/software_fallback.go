package keystore

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
)

type SoftwareKeyStore struct {
	keyPath string
}

func NewSoftwareKeyStore() *SoftwareKeyStore {
	home, _ := os.UserHomeDir()
	return &SoftwareKeyStore{
		keyPath: filepath.Join(home, ".his", "device_key.pem"),
	}
}

func (s *SoftwareKeyStore) GetOrCreate() (*KeyPair, error) {
	if _, err := os.Stat(s.keyPath); os.IsNotExist(err) {
		// Generate new key
		privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		if err != nil {
			return nil, err
		}

		b, err := x509.MarshalECPrivateKey(privateKey)
		if err != nil {
			return nil, err
		}

		os.MkdirAll(filepath.Dir(s.keyPath), 0700)
		pemBlock := &pem.Block{Type: "EC PRIVATE KEY", Bytes: b}
		if err := os.WriteFile(s.keyPath, pem.EncodeToMemory(pemBlock), 0600); err != nil {
			return nil, err
		}
	}

	// Read key
	pemBytes, err := os.ReadFile(s.keyPath)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return nil, os.ErrInvalid
	}

	priv, err := x509.ParseECPrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	pubBytes, err := x509.MarshalPKIXPublicKey(&priv.PublicKey)
	if err != nil {
		return nil, err
	}

	pubPem := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubBytes})

	return &KeyPair{
		PublicKeyPEM: string(pubPem),
	}, nil
}

func (s *SoftwareKeyStore) Sign(data []byte) ([]byte, error) {
	pemBytes, err := os.ReadFile(s.keyPath)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(pemBytes)
	priv, err := x509.ParseECPrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	hash := sha256.Sum256(data)
	return ecdsa.SignASN1(rand.Reader, priv, hash[:])
}
