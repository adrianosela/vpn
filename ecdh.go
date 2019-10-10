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
	"crypto/rand"
	"io"

	"golang.org/x/crypto/curve25519"
)

// DH contains fields relevant to a
// diffie hellman key exchange
type DH struct {
	priv         [32]byte
	pub          [32]byte
	peerPub      []byte
	sharedSecret [32]byte
}

const keySize = 32

// GenerateKey generates a new key pair for a key exchange
func (x *DH) GenerateKey() error {
	// we start by generating 32 secret random bytes
	// from a cryptographically safe source
	priv := [keySize]byte{}
	_, err := io.ReadFull(rand.Reader, priv[:])
	if err != nil {
		return err
	}
	// as per curve25519
	priv[0] &= 248
	priv[31] &= 127
	priv[31] |= 64
	// we then generate the corresponding 32-byte curve25519 public key
	pub := [keySize]byte{}
	curve25519.ScalarBaseMult(&pub, &priv)
	// set on DH
	x.priv = priv
	x.pub = pub
	return nil
}

// ComputeSharedSecret computes the secret shared through the Diffie-Hellman
func (x *DH) ComputeSharedSecret() error {
	// compute shared secret
	shared := [keySize]byte{}

	peerPub := [keySize]byte{}
	copy(peerPub[:], x.peerPub)
	curve25519.ScalarMult(&shared, &x.priv, &peerPub)

	// set on DH
	x.sharedSecret = shared
	return nil
}
