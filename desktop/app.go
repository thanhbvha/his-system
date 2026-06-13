package main

import (
	"context"
	"desktop/internal/keystore"
	"encoding/base64"
)

// App struct
type App struct {
	ctx      context.Context
	keystore keystore.KeyStore
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{
		keystore: keystore.New(),
	}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// GetPublicKey reads or generates the key pair from hardware keystore
func (a *App) GetPublicKey() (string, error) {
	kp, err := a.keystore.GetOrCreate()
	if err != nil {
		return "", err
	}
	return kp.PublicKeyPEM, nil
}

// SignData signs the provided data string using the hardware private key
func (a *App) SignData(data string) (string, error) {
	sig, err := a.keystore.Sign([]byte(data))
	if err != nil {
		return "", err
	}
	// We need base64 encoding for signature transport
	encoded := base64.StdEncoding.EncodeToString(sig)
	return encoded, nil
}
