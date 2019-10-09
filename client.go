package main

// func (c *Client) start() error {
// 	// establish tcp conn
// 	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", c.vpnHost, c.vpnPort), time.Second*10)
// 	if err != nil {
// 		log.Fatalf("could not establish tcp connection to vpn server: %s", err)
// 	}
//
// 	// schedule conn close and add to client
// 	defer conn.Close()
// 	c.conn = conn
// 	log.Printf("[client] established tcp connection with %s:%d", c.vpnHost, c.vpnPort)
//
// 	tcpRxChan := make(chan []byte)
// 	tcpTxChan := make(chan []byte)
//
// 	// this thread reads messages from the the TCP connection
// 	// onto the TCP receive channel and writes messages from the
// 	// TCP transmission channel to the TCP connection
// 	// It also writes to the TCP connection from the TCP
// 	go tcpConnHandler(conn, tcpRxChan, tcpTxChan)
//
// 	// this thread reads messages from the TCP transmission
// 	// channel, then b64 decodes and decrypts it, and finally
// 	// forwards the plaintext to the websocket transmission channel
// 	// (to then be displayed in the UI by the UI thread)
// 	go decodeAndDecrypt(tcpRxChan, uiconf.wsTxChan)
//
// 	// this thread reads messages from the websocket receive
// 	// channel, then encrypts and b64 encodes it, and finally
// 	// forwards the b64-encoded-ciphertext to the TCP
// 	// transmission channel
// 	go encryptAndEncode(uiconf.wsRxChan, tcpTxChan)
//
// 	// catch shutdown
// 	signalCatch := make(chan os.Signal, 1)
// 	signal.Notify(signalCatch, syscall.SIGINT, syscall.SIGKILL, syscall.SIGTERM)
// 	for {
// 		<-signalCatch
// 		// log.Printf("[client] shutdown signal received, terminating")
// 		return nil
// 	}
// }
