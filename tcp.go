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

// func (v *VPN) writer(conn net.Conn) {
// 	for {
// 		msg := "I'm Alice"
// 		err := writeToConn(conn, msg, v.masterSecret)
// 		if err != nil {
// 			log.Printf("[vpn] could not send message to client: %s", err)
// 			return
// 		}
// 		log.Printf("[vpn] sent message: %s", msg)
// 		time.Sleep(time.Second * 1)
// 	}
// }
//
// func (v *VPN) reader(conn net.Conn) {
// 	for {
// 		msg, err := readFromConn(conn, v.masterSecret)
// 		if err != nil {
// 			if err == io.EOF {
// 				log.Printf("[vpn] connection finished (%s) - dropping client", err)
// 			} else {
// 				log.Printf("[vpn] could not read from conn: %s", err)
// 			}
// 			return
// 		}
// 		log.Printf("[vpn] received message: %s", msg)
// 	}
// }
