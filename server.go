package main

import (
	"net"
	"net/http"

	"github.com/gorilla/websocket"
)

type Server struct {
	wsAddr   string
	tcpAddr  string
	certFile string
	keyFile  string
}

func (s *Server) Run() error {
	http.HandleFunc("/proxy", s.handleWebSocket)
	log.Printf("Starting WebSocket server on %s", s.wsAddr)
	return http.ListenAndServeTLS(s.wsAddr, s.certFile, s.keyFile, nil)
}

func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}

	wsConn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}
	defer wsConn.Close()

	tcpConn, err := net.Dial("tcp", s.tcpAddr)
	if err != nil {
		log.Printf("TCP connection failed: %v", err)
		return
	}
	defer tcpConn.Close()

	errChan := make(chan error, 2)

	// TCP -> WebSocket
	go func() {
		for {
			buf := make([]byte, 32*1024)
			n, err := tcpConn.Read(buf)
			if err != nil {
				errChan <- err
				return
			}
			if err := wsConn.WriteMessage(websocket.BinaryMessage, buf[:n]); err != nil {
				errChan <- err
				return
			}
		}
	}()

	// WebSocket -> TCP
	go func() {
		for {
			_, message, err := wsConn.ReadMessage()
			if err != nil {
				errChan <- err
				return
			}
			_, err = tcpConn.Write(message)
			if err != nil {
				errChan <- err
				return
			}
		}
	}()

	// Wait for error
	<-errChan
}
