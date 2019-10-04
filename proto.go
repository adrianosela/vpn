package main

import (
	"bufio"
	"fmt"
	"net"
	"net/textproto"
)

const mockPassphrase = "<< secret passphrase >>"

func writeToConn(conn net.Conn, msg, secret string) error {
	// encrypt message
	ciphertext, err := aesEncrypt([]byte(msg), secret)
	if err != nil {
		return fmt.Errorf("could not encrypt message: %s", err)
	}
	// send message
	if _, err = conn.Write(append(ciphertext, []byte("\n")...)); err != nil {
		return fmt.Errorf("could not write to tcp conn: %s", err)
	}
	return nil
}

func readFromConn(conn net.Conn, secret string) ([]byte, error) {
	// receive data ended with \n or \r\n
	reader := bufio.NewReader(conn)
	tp := textproto.NewReader(reader)
	line, err := tp.ReadLine()
	if err != nil {
		return nil, err
	}
	// decrypt ciphertext
	plaintext, err := aesDecrypt([]byte(line), secret)
	if err != nil {
		return nil, fmt.Errorf("could not decrypt message: %s", err)
	}
	return plaintext, nil
}
