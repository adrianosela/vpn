package main

import (
	"bufio"
	b64 "encoding/base64"
	"fmt"
	"io"
	"net"
)

func writeToConn(conn net.Conn, msg, secret string) error {
	// encrypt message
	ciphertext, err := aesEncrypt([]byte(msg), secret)
	if err != nil {
		return fmt.Errorf("could not encrypt message: %s", err)
	}
	// ascii armour encrypted message
	b64ciphertext := []byte(b64.StdEncoding.EncodeToString(ciphertext))
	// send message
	if _, err = conn.Write(append(b64ciphertext, '\n')); err != nil {
		return fmt.Errorf("could not write to tcp conn: %s", err)
	}
	return nil
}

func readFromConn(conn net.Conn, secret string) ([]byte, error) {
	// receive data ended with \n or \r\n
	bufReader := bufio.NewReader(conn)
	// Read tokens delimited by newline
	line, err := bufReader.ReadBytes('\n')
	if err != nil {
		if err == io.EOF {
			return nil, err
		}
		return nil, fmt.Errorf("could not read conn to newline: %s", err)
	}
	n := len(line)
	// base64 decode the ciphertext
	decodedCiphertext, err := b64.StdEncoding.DecodeString(string(line[:n-1]))
	if err != nil {
		return nil, fmt.Errorf("could not base64 decode message: %s", err)
	}
	// decrypt ciphertext
	plaintext, err := aesDecrypt(decodedCiphertext, secret)
	if err != nil {
		return nil, fmt.Errorf("could not decrypt message: %s", err)
	}
	return plaintext, nil
}
