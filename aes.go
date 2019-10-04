package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
)

// AES passphrases must be 32 bytes in length.
// we ensure that length by taking an md5 hash.
func keyFromPassphrase(pass string) []byte {
	hasher := md5.New()
	hasher.Write([]byte(pass))
	return []byte(hex.EncodeToString(hasher.Sum(nil)))
}

// builds an AES block cipher from a given passphrase
func aesBlockFromPassphrase(pass string) (cipher.AEAD, error) {
	// get a 32 byte key from the passphrase
	key := keyFromPassphrase(pass)
	// create new block cipher from key
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("could not create AES block cipher: %s", err)
	}
	// wrap the 32 byte block cipher in Galois Counter Mode
	// https://en.wikipedia.org/wiki/Galois/Counter_Mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("could not wrap block cipher in GCM: %s", err)
	}
	return gcm, nil
}

func aesEncrypt(data []byte, passphrase string) ([]byte, error) {
	// get block cipher
	gcm, err := aesBlockFromPassphrase(passphrase)
	if err != nil {
		return nil, fmt.Errorf("could not build AES block from passphrase: %s", err)
	}
	// fill up nonce bytes randomly
	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("could not generate random nonce: %s", err)
	}
	// encrypt and return
	ciphertext := gcm.Seal(nonce, nonce, data, nil)
	return ciphertext, nil
}

func aesDecrypt(data []byte, passphrase string) ([]byte, error) {
	// get block cipher
	gcm, err := aesBlockFromPassphrase(passphrase)
	if err != nil {
		return nil, fmt.Errorf("could not build AES block from passphrase: %s", err)
	}
	// split up nonce from actual ciphertext
	nonceSize := gcm.NonceSize()
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	// decrypt and return
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("could not decrypt ciphertext: %s", err)
	}
	return plaintext, nil
}
