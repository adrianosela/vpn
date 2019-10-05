package main

import (
	"encoding/json"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second    // time allowed to write a message to the peer
	pongWait       = 60 * time.Second    // time allowed to read the next pong message from the peer
	pingPeriod     = (pongWait * 9) / 10 // send pings to peer with this period. Must be less than pongWait
	maxMessageSize = 512                 // Maximum message size allowed from peer
)

func wsConnHandler(conn *websocket.Conn, wsRxChan, wsTxChan chan []byte) {
	go wsReader(conn, wsRxChan)
	go wsWriter(conn, wsTxChan)
}

type msgJSON struct {
	From string `json:"from,omitempty"` // peer ID of sender
	To   string `json:"to,omitempty"`   // peer ID of receiver
	Data string `json:"data"`           // message body
}

// the wsReader reads from the websocket connection
// onto the websocket receive channel
func wsReader(conn *websocket.Conn, tcpRxChan chan []byte) {
	defer conn.Close()
	conn.SetReadLimit(maxMessageSize)
	conn.SetReadDeadline(time.Now().Add(pongWait))
	conn.SetPongHandler(func(string) error { conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, jsonMsg, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WS connection was closed unexpectedly: %s", err)
			}
			break
		}
		var msg msgJSON
		if err = json.Unmarshal(jsonMsg, &msg); err != nil {
			log.Printf("WS connection was closed unexpectedly: %s", err)
			break
		}
		tcpRxChan <- []byte(msg.Data)
	}
}

// the wsWriter reads from the websocket transmit channel
// and writes onto the websocket connection
func wsWriter(conn *websocket.Conn, wsTxChan chan []byte) {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		conn.Close()
	}()
	for {
		select {
		case message, ok := <-wsTxChan:
			conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			w, err := conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)
			// Add queued chat messages to the current websocket message.
			n := len(wsTxChan)
			for i := 0; i < n; i++ {
				w.Write([]byte("\n"))
				w.Write(<-wsTxChan)
			}
			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
