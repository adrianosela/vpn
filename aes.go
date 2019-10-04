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
func aesKeyFromPassphrase(passphrase string) []byte {
	hasher := md5.New()
	hasher.Write([]byte(passphrase))
	return []byte(hex.EncodeToString(hasher.Sum(nil)))
}

func aesEncrypt(data []byte, passphrase string) ([]byte, error) {
	block, err := aes.NewCipher(aesKeyFromPassphrase(passphrase))
	if err != nil {
		return nil, fmt.Errorf("could not create AES block cipher: %s", err)
	}
	// NewGCM returns the given 128-bit, block cipher wrapped
	// in Galois Counter Mode with the standard nonce length.
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("could not wrap block cipher in GCM: %s", err)
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("could not generate random nonce: %s", err)
	}
	ciphertext := gcm.Seal(nonce, nonce, data, nil)
	return ciphertext, nil
}

func aesDecrypt(data []byte, passphrase string) []byte {
	key := aesKeyFromPassphrase(passphrase)
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err.Error())
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		panic(err.Error())
	}
	nonceSize := gcm.NonceSize()
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		panic(err.Error())
	}
	return plaintext
}
