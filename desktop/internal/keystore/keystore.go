package keystore

type KeyPair struct {
	PublicKeyPEM string
}

type KeyStore interface {
	GetOrCreate() (*KeyPair, error)
	Sign(data []byte) ([]byte, error)
}
