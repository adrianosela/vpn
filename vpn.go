package main

// // start runs the vpn TCP service
// func (v *VPN) start() error {
// 	// establish tcp listener for the vpn service
// 	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", v.listenTCPPort))
// 	if err != nil {
// 		return fmt.Errorf("could not establish tcp listener: %s", err)
// 	}
// 	defer ln.Close()
// 	log.Printf("[vpn] started tcp listener on :%d", v.listenTCPPort)
//
// 	// dispatch UI thread, wait a sec, open browser
// 	uiconf := &uiConfig{
// 		wsRxChan: make(chan []byte),
// 		wsTxChan: make(chan []byte),
// 		uiPort:   v.uiPort,
// 	}
// 	go ui(uiconf)
//
// 	// accept and handle client
// 	for {
// 		conn, err := ln.Accept()
// 		if err != nil {
// 			log.Printf("[vpn] failed to accept tcp connection: %s", err)
// 			continue
// 		}
// 		log.Println("[vpn] accepted new tcp connection")
//
// 		tcpRxChan := make(chan []byte)
// 		tcpTxChan := make(chan []byte)
//
// 		// this thread reads messages from the the TCP connection
// 		// onto the TCP receive channel and writes messages from the
// 		// TCP transmission channel to the TCP connection
// 		// It also writes to the TCP connection from the TCP
// 		go tcpConnHandler(conn, tcpRxChan, tcpTxChan)
//
// 		// this thread reads messages from the TCP transmission
// 		// channel, then b64 decodes and decrypts it, and finally
// 		// forwards the plaintext to the websocket transmission channel
// 		// (to then be displayed in the UI by the UI thread)
// 		go decodeAndDecrypt(tcpRxChan, uiconf.wsTxChan)
//
// 		// this thread reads messages from the websocket receive
// 		// channel, then encrypts and b64 encodes it, and finally
// 		// forwards the b64-encoded-ciphertext to the TCP
// 		// transmission channel
// 		go encryptAndEncode(uiconf.wsRxChan, tcpTxChan)
// 	}
// }
