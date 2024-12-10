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
	log.Infof("WebSocket server starting on %s", s.wsAddr)
	log.Debugf("Server configuration: TCP address=%s, Cert=%s, Key=%s",
		s.tcpAddr, s.certFile, s.keyFile)
	return http.ListenAndServeTLS(s.wsAddr, s.certFile, s.keyFile, nil)
}

func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	log.Debugf("New WebSocket connection request from %s", r.RemoteAddr)

	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}

	wsConn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Errorf("WebSocket upgrade failed: %v", err)
		return
	}
	defer wsConn.Close()
	log.Infof("WebSocket connection established with %s", r.RemoteAddr)

	log.Debugf("Attempting TCP connection to %s", s.tcpAddr)
	tcpConn, err := net.Dial("tcp", s.tcpAddr)
	if err != nil {
		log.Errorf("TCP connection failed: %v", err)
		return
	}
	defer tcpConn.Close()
	log.Infof("TCP connection established with %s", s.tcpAddr)

	errChan := make(chan error, 2)

	// TCP -> WebSocket
	go func() {
		log.Debug("Starting TCP -> WebSocket forwarding")
		for {
			buf := make([]byte, 32*1024)
			n, err := tcpConn.Read(buf)
			if err != nil {
				log.Debugf("TCP read error: %v", err)
				errChan <- err
				return
			}
			log.Tracef("Read %d bytes from TCP connection", n)

			if err := wsConn.WriteMessage(websocket.BinaryMessage, buf[:n]); err != nil {
				log.Debugf("WebSocket write error: %v", err)
				errChan <- err
				return
			}
			log.Tracef("Forwarded %d bytes to WebSocket", n)
		}
	}()

	// WebSocket -> TCP
	go func() {
		log.Debug("Starting WebSocket -> TCP forwarding")
		for {
			_, message, err := wsConn.ReadMessage()
			if err != nil {
				log.Debugf("WebSocket read error: %v", err)
				errChan <- err
				return
			}
			log.Tracef("Read %d bytes from WebSocket", len(message))

			_, err = tcpConn.Write(message)
			if err != nil {
				log.Debugf("TCP write error: %v", err)
				errChan <- err
				return
			}
			log.Tracef("Forwarded %d bytes to TCP", len(message))
		}
	}()

	// Wait for error
	err = <-errChan
	log.Infof("Connection closed: %v", err)
}
