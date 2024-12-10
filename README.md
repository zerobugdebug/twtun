# TCP over WebSocket Proxy (twtun)

A secure and flexible proxy that tunnels TCP connections through WebSocket, supporting both client and server components. Ideal for scenarios where direct TCP connectivity is restricted but WebSocket connections are allowed.

## Features

- Bidirectional TCP-to-WebSocket proxy
- TLS encryption for secure WebSocket connections
- HTTP proxy support for client connections
- Configurable logging levels
- Support for both long and short command-line arguments
- Efficient binary message handling

## Installation

```bash
go get github.com/zerobugdebug/twtun
```

## Usage

The application can run in two modes: server and client.

### Server Mode

The server accepts WebSocket connections and proxies them to a specified TCP address.

```bash
# Using long form arguments
twtun --mode server --ws-addr :8080 --tcp-addr :9000 --cert server.crt --key server.key

# Using short form arguments
twtun -m server -w :8080 -t :9000 -c server.crt -k server.key
```

### Client Mode

The client accepts TCP connections and proxies them over WebSocket to the server.

```bash
# Basic client setup
twtun --mode client --ws-addr example.com:8080 --tcp-addr :9000

# Client with HTTP proxy
twtun -m client -w example.com:8080 -t :9000 -p http://proxy:8080
```

## Command Line Options

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--mode` | `-m` | Operation mode: server or client (required) | |
| `--ws-addr` | `-w` | WebSocket address to listen on (server) or connect to (client) | `:8080` |
| `--tcp-addr` | `-t` | TCP address to proxy | `:9000` |
| `--cert` | `-c` | SSL certificate file (server mode) | `server.crt` |
| `--key` | `-k` | SSL key file (server mode) | `server.key` |
| `--proxy` | `-p` | HTTP proxy address (client mode) | |
| `--log-level` | `-l` | Set log level (debug, info, warn, error) | `info` |
| `--help` | `-h` | Display usage information | |

## Common Use Cases

### Remote Service Access

Access internal services through firewalls that allow WebSocket connections:

```bash
# Server (internal network)
twtun -m server -w :8080 -t localhost:5432 -c cert.pem -k key.pem

# Client (external network)
twtun -m client -w internal-server:8080 -t localhost:5432
```

### Database Connection Through Proxy

Connect to a database server through a corporate proxy:

```bash
# Server (database network)
twtun -m server -w :8443 -t db-server:3306 -c cert.pem -k key.pem

# Client (local machine)
twtun -m client -w db-gateway:8443 -t localhost:3306 -p http://corporate-proxy:8080
```

## Security Considerations

1. Always use TLS certificates in production environments
2. The server's `InsecureSkipVerify` option in client mode should be disabled in production
3. Implement appropriate access controls and firewall rules
4. Regularly update TLS certificates and security credentials

## Building from Source

```bash
git clone https://github.com/zerobugdebug/twtun
cd twtun
go build
```

## Requirements

- Go 1.16 or higher
- gorilla/websocket package
- sirupsen/logrus package

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details.
