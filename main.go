package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
)

// Define a custom flag type to handle both short and long forms
type Flag struct {
	Name     string
	Short    string
	Usage    string
	Default  string
	Variable *string
}

// Define all flags with their short and long forms
var flags = []Flag{
	{Name: "mode", Short: "m", Usage: "Operation mode: server or client (required)", Default: "", Variable: new(string)},
	{Name: "ws-addr", Short: "w", Usage: "WebSocket address to listen on (server) or connect to (client)", Default: ":8080", Variable: new(string)},
	{Name: "tcp-addr", Short: "t", Usage: "TCP address to proxy", Default: ":9000", Variable: new(string)},
	{Name: "cert", Short: "c", Usage: "SSL certificate file (server mode)", Default: "server.crt", Variable: new(string)},
	{Name: "key", Short: "k", Usage: "SSL key file (server mode)", Default: "server.key", Variable: new(string)},
	{Name: "proxy", Short: "p", Usage: "HTTP proxy address (client mode)", Default: "", Variable: new(string)},
	{Name: "log-level", Short: "l", Usage: "Set log level (debug, info, warn, error)", Default: "info", Variable: new(string)},
}

var (
	help      = flag.Bool("help", false, "Display usage information")
	helpShort = flag.Bool("h", false, "Display usage information")
)

var log *logrus.Logger

func usage() {
	fmt.Fprintf(os.Stderr, "TCP over WebSocket Proxy\n\n")
	fmt.Fprintf(os.Stderr, "Usage:\n")
	fmt.Fprintf(os.Stderr, "  Server mode: %s (-m|--mode) server [options]\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "  Client mode: %s (-m|--mode) client [options]\n\n", os.Args[0])

	fmt.Fprintf(os.Stderr, "Options:\n")
	for _, f := range flags {
		fmt.Fprintf(os.Stderr, "  -%s, --%-15s %s (default: %q)\n",
			f.Short, f.Name, f.Usage, f.Default)
	}
	fmt.Fprintf(os.Stderr, "  -h, --help            Display usage information\n")

	fmt.Fprintf(os.Stderr, "\nExamples:\n")
	fmt.Fprintf(os.Stderr, "  Start server:\n")
	fmt.Fprintf(os.Stderr, "    %s --mode server --ws-addr :8080 --tcp-addr :9000 --cert server.crt --key server.key\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "    %s -m server -w :8080 -t :9000 -c server.crt -k server.key\n\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "  Start client:\n")
	fmt.Fprintf(os.Stderr, "    %s --mode client --ws-addr example.com:8080 --tcp-addr :9000\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "    %s -m client -w example.com:8080 -t :9000 -p http://proxy:8080\n\n", os.Args[0])

	fmt.Fprintf(os.Stderr, "Description:\n")
	fmt.Fprintf(os.Stderr, "  This program creates a TCP-over-WebSocket proxy with both server and client components.\n")
	fmt.Fprintf(os.Stderr, "  The server accepts WebSocket connections and proxies them to a TCP address.\n")
	fmt.Fprintf(os.Stderr, "  The client accepts TCP connections and proxies them over WebSocket to the server.\n")
}

// parseFlags parses both short and long form flags
func parseFlags() error {
	// Check for help first, before any other parsing
	for _, arg := range os.Args[1:] {
		if arg == "-h" || arg == "--help" {
			usage()
			os.Exit(0)
		}
	}

	// Set default values
	for _, f := range flags {
		*f.Variable = f.Default
	}

	args := os.Args[1:]
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if len(arg) == 0 || arg[0] != '-' {
			return fmt.Errorf("unexpected argument: %s", arg)
		}

		isLong := arg[1] == '-'
		name := arg[1:]
		if isLong {
			name = arg[2:]
		}

		// Look for matching flag
		var matchedFlag *Flag
		for i := range flags {
			f := &flags[i]
			if (isLong && f.Name == name) || (!isLong && f.Short == name) {
				matchedFlag = f
				break
			}
		}

		if matchedFlag == nil {
			return fmt.Errorf("unknown flag: %s", arg)
		}

		// Check if there's a value following the flag
		if i+1 >= len(args) {
			return fmt.Errorf("missing value for flag: %s", arg)
		}

		// Set the value
		*matchedFlag.Variable = args[i+1]
		i++ // skip the value in next iteration
	}

	return nil
}

func main() {

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

	if err := parseFlags(); err != nil {
		log.Errorf("Error: %v\n", err)
		usage()
		os.Exit(1)
	}

	// Get mode from our flags slice
	mode := ""
	for _, f := range flags {
		if f.Name == "mode" {
			mode = *f.Variable
			break
		}
	}

	// Show help if -help flag is provided or if mode is empty
	if *help || *helpShort || mode == "" {
		flag.Usage()
		// Exit with error status if mode is missing (but not if help was explicitly requested)
		if mode == "" && !*help && !*helpShort {
			fmt.Fprintf(os.Stderr, "\nError: -m|--mode parameter is required (must be 'server' or 'client')\n")
			os.Exit(1)
		}
		os.Exit(0)
	}

	// Get other flag values
	var wsAddr, tcpAddr, certFile, keyFile, proxy, logLevel string
	for _, f := range flags {
		switch f.Name {
		case "ws-addr":
			wsAddr = *f.Variable
		case "tcp-addr":
			tcpAddr = *f.Variable
		case "cert":
			certFile = *f.Variable
		case "key":
			keyFile = *f.Variable
		case "proxy":
			proxy = *f.Variable
		case "log-level":
			logLevel = *f.Variable
		}
	}

	// Log the user-specified log level
	log.Infof("Log Level: %s\n", logLevel)

	// Parse the user-provided log level string into a logrus Level type
	logrusLevel, err := logrus.ParseLevel(logLevel)
	if err != nil {
		// If parsing fails, log the error and default to "info" level
		log.Infof("Failed to parse log level \"%s\"\n%v\nDefaulting to info", logLevel, err)
		logrusLevel = logrus.InfoLevel
	}

	// Set the logger to use the parsed (or defaulted) log level
	log.SetLevel(logrusLevel)

	switch mode {
	case "server":
		server := &Server{
			wsAddr:   wsAddr,
			tcpAddr:  tcpAddr,
			certFile: certFile,
			keyFile:  keyFile,
		}
		log.Fatal(server.Run())
	case "client":
		client := &Client{
			wsAddr:  wsAddr,
			tcpAddr: tcpAddr,
			proxy:   proxy,
		}
		log.Fatal(client.Run())
	default:
		log.Fatalf("Invalid mode: %s", mode)
	}
}
