package main

import (
	"crypto/tls"
	"net"
	"net/http"
	"net/url"

	"github.com/gorilla/websocket"

)

type Client struct {
	wsAddr  string
	tcpAddr string
	proxy   string
}

func (c *Client) Run() error {
	listener, err := net.Listen("tcp", c.tcpAddr)
	if err != nil {
		return err
	}
	log.Infof("TCP listener started on %s", c.tcpAddr)
	log.Debugf("Client configuration: WebSocket address=%s, Proxy=%v", c.wsAddr, c.proxy != "")

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Errorf("Failed to accept connection: %v", err)
			continue
		}
		log.Debugf("New TCP connection accepted from %s", conn.RemoteAddr())
		go c.handleConnection(conn)
	}
}

func (c *Client) handleConnection(conn net.Conn) {
	defer conn.Close()
	log.Debugf("Handling new connection from %s", conn.RemoteAddr())

	dialer := websocket.Dialer{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, // Note: In production, properly verify certificates
		},
	}

	if c.proxy != "" {
		proxyURL, err := url.Parse(c.proxy)
		if err != nil {
			log.Errorf("Invalid proxy URL: %v", err)
			return
		}
		dialer.Proxy = http.ProxyURL(proxyURL)
		log.Debugf("Using proxy: %s", c.proxy)
	}

	wsURL := "wss://" + c.wsAddr + "/proxy"
	log.Debugf("Attempting WebSocket connection to %s", wsURL)
	wsConn, _, err := dialer.Dial(wsURL, nil)
	if err != nil {
		log.Errorf("WebSocket connection failed: %v", err)
		return
	}
	defer wsConn.Close()
	log.Infof("WebSocket connection established with %s", c.wsAddr)

	errChan := make(chan error, 2)

	// TCP -> WebSocket
	go func() {
		log.Debug("Starting TCP -> WebSocket forwarding")
		for {
			buf := make([]byte, 32*1024)
			n, err := conn.Read(buf)
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

			_, err = conn.Write(message)
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
