package main

import (
	"bufio"
	"io"
	"log"
	"net"
)

func tcpConnHandler(conn net.Conn, tcpRxChan, tcpTxChan chan []byte) {
	go tcpReader(conn, tcpRxChan)
	go tcpWriter(conn, tcpTxChan)
}

// the tcpReader reads from the tcp connection
// onto the tcp receive channel
func tcpReader(conn net.Conn, tcpRxChan chan []byte) {
	bufReader := bufio.NewReader(conn)

	for {
		// read until newline
		line, err := bufReader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				return
			}
			log.Printf("could not read conn to newline: %s", err)
			return
		}
		// write message to tcp receive chan
		tcpRxChan <- line
	}
}

// the tcpWriter reads from the tcp transmit channel
// and writes onto the tp connection
func tcpWriter(conn net.Conn, tcpTxChan chan []byte) {
	for {
		select {
		case data := <-tcpTxChan:
			if _, err := conn.Write(data); err != nil {
				log.Printf("[proto] could not write to tcp conn: %s", err)
				return
			}
		}
	}
}
