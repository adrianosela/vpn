package main

import (
	"crypto/md5"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha512"
	"strings"

	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"fmt"
)

// generateRSAKey generates an RSA key-pair (priv contains pub)
func generateRSAKey(bits int) (*rsa.PrivateKey, error) {
	priv, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return nil, err
	}
	return priv, nil
}

// getKeyID returns the md5 sum of a public key
// this is also known as the keys "fingerprint"
func getKeyID(pub *rsa.PublicKey) string {
	md5sum := md5.Sum(x509.MarshalPKCS1PublicKey(pub))
	hexarray := make([]string, len(md5sum))
	for i, c := range md5sum {
		hexarray[i] = hex.EncodeToString([]byte{c})
	}
	return strings.Join(hexarray, ":")
}

// encodePrivKeyPEM encodes an *rsa.PrivateKey onto a PEM block
func encodePrivKeyPEM(priv *rsa.PrivateKey) []byte {
	return pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(priv),
	})
}

// encodePubKeyPEM encodes an *rsa.PublicKey onto a PEM block
func encodePubKeyPEM(pub *rsa.PublicKey) []byte {
	return pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: x509.MarshalPKCS1PublicKey(pub),
	})
}

// decodePrivKeyPEM decodes a PEM encoded public key to an *rsa.PublicKey
func decodePrivKeyPEM(pk []byte) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(pk)
	if block == nil || block.Type != "RSA PRIVATE KEY" {
		return nil, fmt.Errorf("failed to decode PEM block containing private key")
	}
	return x509.ParsePKCS1PrivateKey(block.Bytes)
}

// decodePubKeyPEM decodes a PEM encoded public key to an *rsa.PublicKey
func decodePubKeyPEM(pk []byte) (*rsa.PublicKey, error) {
	block, _ := pem.Decode(pk)
	if block == nil || block.Type != "RSA PUBLIC KEY" {
		return nil, fmt.Errorf("failed to decode PEM block containing public key")
	}
	return x509.ParsePKCS1PublicKey(block.Bytes)
}

// encryptMessage encrypts a plaintext message with a public key
func encryptMessage(plaintxt []byte, pub *rsa.PublicKey) ([]byte, error) {
	hash := sha512.New()
	cyphertxt, err := rsa.EncryptOAEP(hash, rand.Reader, pub, plaintxt, nil)
	if err != nil {
		return nil, err
	}
	return cyphertxt, nil
}

// decryptMessage decrypts an encrypted message with a private key
func decryptMessage(cyphertxt []byte, priv *rsa.PrivateKey) ([]byte, error) {
	hash := sha512.New()
	plaintxt, err := rsa.DecryptOAEP(hash, rand.Reader, priv, cyphertxt, nil)
	if err != nil {
		return nil, err
	}
	return plaintxt, nil
}

// encryptMessageWithPEMKey encrypts a plaintext message with a PEM encoded public key
func encryptMessageWithPEMKey(plaintxt []byte, pub []byte) ([]byte, error) {
	k, err := decodePubKeyPEM(pub)
	if err != nil {
		return nil, err
	}
	return encryptMessage(plaintxt, k)
}

// decryptMessageWithPEMKey decrypts an encrypted message with a PEM private key
func decryptMessageWithPEMKey(cyphertxt []byte, priv []byte) ([]byte, error) {
	k, err := decodePrivKeyPEM(priv)
	if err != nil {
		return nil, err
	}
	return decryptMessage(cyphertxt, k)
}
