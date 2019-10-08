package main

import (
	"crypto"
	"crypto/rand"
	"github.com/aead/ecdh"
)

var (
	kx = ecdh.X25519()
)

func generateECDHPair() (crypto.PrivateKey, crypto.PublicKey, error) {
	return kx.GenerateKey(rand.Reader)
}

func generateSharedSecret(priv crypto.PrivateKey, peersPub crypto.PublicKey) []byte {
	return kx.ComputeSecret(priv, peersPub)
}
