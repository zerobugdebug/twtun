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
	log.Printf("Listening for TCP connections on %s", c.tcpAddr)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Failed to accept connection: %v", err)
			continue
		}
		go c.handleConnection(conn)
	}
}

func (c *Client) handleConnection(conn net.Conn) {
	defer conn.Close()

	dialer := websocket.Dialer{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, // Note: In production, properly verify certificates
		},
	}

	if c.proxy != "" {
		proxyURL, err := url.Parse(c.proxy)
		if err != nil {
			log.Printf("Invalid proxy URL: %v", err)
			return
		}
		dialer.Proxy = http.ProxyURL(proxyURL)
	}

	wsURL := "wss://" + c.wsAddr + "/proxy"
	wsConn, _, err := dialer.Dial(wsURL, nil)
	if err != nil {
		log.Printf("WebSocket connection failed: %v", err)
		return
	}
	defer wsConn.Close()

	errChan := make(chan error, 2)

	// TCP -> WebSocket
	go func() {
		for {
			buf := make([]byte, 32*1024)
			n, err := conn.Read(buf)
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
			_, err = conn.Write(message)
			if err != nil {
				errChan <- err
				return
			}
		}
	}()

	// Wait for error
	<-errChan
}
