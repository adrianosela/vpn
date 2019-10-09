package main

/**
 * This implementation of Elliptic Curve Diffie-Hellman key exchange
 * is based on Curve25519, a state-of-the-art Diffie-Hellman function
 * suitable for a wide variety of applications.
 *
 * source: https://cr.yp.to/ecdh.html
 *
 * This code has been adapted from https://github.com/aead/ecdh, which
 * is not being used here due to its use of panic()
 */

import (
	"crypto"
	"crypto/rand"
	"errors"
	"golang.org/x/crypto/curve25519"
	"io"
)

const keySize = 32

// GenerateKey generates a new key pair for a key exchange
func GenerateKey() (crypto.PrivateKey, crypto.PublicKey, error) {
	// we start by generating 32 secret random bytes
	// from a cryptographically safe source
	priv := [keySize]byte{}
	_, err := io.ReadFull(rand.Reader, priv[:])
	if err != nil {
		return nil, nil, err
	}
	// as per curve25519
	priv[0] &= 248
	priv[31] &= 127
	priv[31] |= 64
	// we then generate the corresponding 32-byte curve25519 public key
	pub := [keySize]byte{}
	curve25519.ScalarBaseMult(&pub, &priv)
	return priv, pub, nil
}

// ComputeSharedSecret computes the secret shared through the Diffie-Hellman
func ComputeSharedSecret(priv crypto.PrivateKey, peerPub crypto.PublicKey) ([]byte, error) {
	priv32 := [keySize]byte{}
	pub32 := [keySize]byte{}
	// convert keys to bytes (to do math on them)
	if privBytes, ok := priv.([]byte); ok {
		copy(priv32[:], privBytes)
	} else {
		return nil, errors.New("bad private key type")
	}
	if pubBytes, ok := peerPub.([]byte); ok {
		copy(pub32[:], pubBytes)
	} else {
		return nil, errors.New("bad public key type")
	}
	// compute shared secret
	shared := [keySize]byte{}
	curve25519.ScalarMult(&shared, &priv32, &pub32)
	return shared[:], nil
}
