package main

import (
	b64 "encoding/base64"
	"log"
)

// encrypts and base64 encodes data from the input channel
// and then writes it to the output channel
func encryptAndEncode(in, out chan []byte, passphrase string) {
	for {
		select {
		case msg := <-in:
			// encrypt message
			ciphertext, err := aesEncrypt(msg, passphrase)
			if err != nil {
				log.Printf("[proto] could not encrypt message: %s", err)
				return
			}
			// ascii armour encrypted message (and add newline for easy msg cutting)
			b64ciphertext := append([]byte(b64.StdEncoding.EncodeToString(ciphertext)), '\n')
			// write b64 ciphertext to out chan
			out <- b64ciphertext
		}
	}
}

// base64 decodes and then decrypts data from the input channel
// then writes it to the output channel
func decodeAndDecrypt(in, out chan []byte, passphrase string) {
	for {
		select {
		case data := <-in:
			// base64 decode the ciphertext (chop off newline)
			decodedCiphertext, err := b64.StdEncoding.DecodeString(string(data[:len(data)-1]))
			if err != nil {
				log.Printf("[proto] could not b64 decode message: %s", err)
				return
			}
			// decrypt ciphertext
			plaintext, err := aesDecrypt(decodedCiphertext, passphrase)
			if err != nil {
				log.Printf("[proto] could not decrypt ciphertext: %s", err)
				return
			}
			// write plaintext to out chan
			out <- plaintext
		}
	}
}
