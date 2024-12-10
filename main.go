package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/sirupsen/logrus"

)

var (
	mode     = flag.String("mode", "", "Operation mode: server or client (required)")
	wsAddr   = flag.String("ws-addr", ":8080", "WebSocket address to listen on (server) or connect to (client)")
	tcpAddr  = flag.String("tcp-addr", ":9000", "TCP address to proxy")
	certFile = flag.String("cert", "server.crt", "SSL certificate file (server mode)")
	keyFile  = flag.String("key", "server.key", "SSL key file (server mode)")
	proxy    = flag.String("proxy", "", "HTTP proxy address (client mode)")
	logLevel = flag.String("log-level", "info", "Set log level (debug, info, warn, error)")
	help     = flag.Bool("help", false, "Display usage information")
)

var log *logrus.Logger

func usage() {
	fmt.Fprintf(os.Stderr, "TCP over WebSocket Proxy\n\n")
	fmt.Fprintf(os.Stderr, "Usage:\n")
	fmt.Fprintf(os.Stderr, "  Server mode: %s -mode server [options]\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "  Client mode: %s -mode client [options]\n\n", os.Args[0])

	fmt.Fprintf(os.Stderr, "Options:\n")
	flag.PrintDefaults()

	fmt.Fprintf(os.Stderr, "\nExamples:\n")
	fmt.Fprintf(os.Stderr, "  Start server:\n")
	fmt.Fprintf(os.Stderr, "    %s -mode server -ws-addr :8080 -tcp-addr :9000 -cert server.crt -key server.key\n\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "  Start client:\n")
	fmt.Fprintf(os.Stderr, "    %s -mode client -ws-addr example.com:8080 -tcp-addr :9000\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "    %s -mode client -ws-addr example.com:8080 -tcp-addr :9000 -proxy http://proxy:8080\n\n", os.Args[0])

	fmt.Fprintf(os.Stderr, "Description:\n")
	fmt.Fprintf(os.Stderr, "  This program creates a TCP-over-WebSocket proxy with both server and client components.\n")
	fmt.Fprintf(os.Stderr, "  The server accepts WebSocket connections and proxies them to a TCP address.\n")
	fmt.Fprintf(os.Stderr, "  The client accepts TCP connections and proxies them over WebSocket to the server.\n")
}

func main() {
	flag.Usage = usage
	flag.Parse()

	// Show help if -help flag is provided or if mode is empty
	if *help || *mode == "" {
		flag.Usage()
		// Exit with error status if mode is missing (but not if help was explicitly requested)
		if *mode == "" && !*help {
			fmt.Fprintf(os.Stderr, "\nError: -mode parameter is required (must be 'server' or 'client')\n")
			os.Exit(1)
		}
		os.Exit(0)
	}

// Initialize error variable for later use
var err error

// Create a new instance of logrus logger
log = logrus.New()

// Configure the log output formatter
log.SetFormatter(&logrus.TextFormatter{
    // Set timestamp format to include milliseconds
    TimestampFormat: "2006-01-02 15:04:05.000",
    // Force colored output even when not running in a terminal
    ForceColors: true,
    // Show full timestamp in each log entry
    FullTimestamp: true,
})

// Set initial log level to Debug (will be overridden by user setting)
log.SetLevel(logrus.DebugLevel)

// Log the user-specified log level
log.Infof("Log Level: %s\n", *logLevel)

// Parse the user-provided log level string into a logrus Level type
logrusLevel, err := logrus.ParseLevel(*logLevel)
if err != nil {
    // If parsing fails, log the error and default to "info" level
    log.Fatalf("Failed to parse log level \"%s\"\n%v\nDefaulting to info", *logLevel, err)
    logrusLevel = logrus.InfoLevel
}

// Set the logger to use the parsed (or defaulted) log level
log.SetLevel(logrusLevel)

	switch *mode {
	case "server":
		server := &Server{
			wsAddr:   *wsAddr,
			tcpAddr:  *tcpAddr,
			certFile: *certFile,
			keyFile:  *keyFile,
		}
		log.Fatal(server.Run())
	case "client":
		client := &Client{
			wsAddr:  *wsAddr,
			tcpAddr: *tcpAddr,
			proxy:   *proxy,
		}
		log.Fatal(client.Run())
	default:
		log.Fatalf("Invalid mode: %s", *mode)
	}
}
