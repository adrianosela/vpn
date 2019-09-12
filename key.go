package main

import (
	"crypto/rand"
	"crypto/rsa"
)

// generateRSAKey generates an RSA key pair (priv contains pub)
func generateRSAKey(bits int) (*rsa.PrivateKey, error) {
	priv, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return nil, err
	}
	return priv, nil
}
