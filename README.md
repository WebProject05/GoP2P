# GoP2P

A decentralized peer-to-peer file sharing and chat application built in Go. This tool enables secure, direct communication and file transfers between peers without relying on centralized servers for data transmission.

## Features

- **Decentralized Chat Rooms**: Join mesh networks for group chat with end-to-end encryption
- **Direct File Transfer**: Send files directly to known peer IPs
- **Private 1-on-1 Chat**: Encrypted private messaging between two peers
- **Peer Discovery**: Uses a signaling server for NAT traversal and peer discovery
- **End-to-End Encryption**: AES encryption with ECDH key exchange for secure communication
- **Cross-Platform**: Built with Go, runs on Windows, Linux, and macOS

## Architecture

The application consists of several components:

- **Main Application** (`main.go`): Command-line interface for different modes
- **Chat Module** (`internal/chat/`): Handles chat functionality and room management
- **Crypto Module** (`internal/crypto/`): Provides encryption and key exchange
- **Discovery Module** (`internal/discovery/`): Manages peer discovery via signaling server
- **Transfer Module** (`internal/transfer/`): Handles file transfers
- **Signaling Server** (`cmd/signaler/`): Matchmaker server for peer discovery

## Prerequisites

- Go 1.23.4 or later
- Network connectivity for P2P communication

## Installation

1. Clone the repository:
   ```bash
   git clone <repository-url>
   cd p2p-share
   ```

2. Build the application:
   ```bash
   go build -o p2p .
   ```

3. (Optional) Build the signaling server:
   ```bash
   go build -o signaler ./cmd/signaler
   ```

## Usage

### Starting the Signaling Server

First, start the signaling server on a publicly accessible machine:

```bash
./signaler
```

The server runs on port 8080 by default.

### Joining a Chat Room

Join a decentralized chat mesh:

```bash
./p2p room <username> <signaling_server_ip>
```

Example:
```bash
./p2p room Alice 192.168.1.100
```

This will:
- Connect to the signaling server
- Discover other peers
- Establish encrypted connections
- Allow broadcasting messages to all peers in the room

### Direct File Transfer

Send a file to a known peer IP:

```bash
./p2p send <peer_ip> <file_path>
```

Example:
```bash
./p2p send 192.168.1.101 /path/to/file.txt
```

### Private 1-on-1 Chat

Start a private chat with a known peer:

```bash
./p2p chat <peer_ip>
```

Example:
```bash
./p2p chat 192.168.1.102
```

## Security

All communications are encrypted using:
- AES-256 for message encryption
- ECDH (Elliptic Curve Diffie-Hellman) for key exchange
- Each peer-to-peer connection establishes its own encryption keys

## Development

### Project Structure

```
p2p-share/
├── main.go                 # Main entry point
├── cmd/
│   └── signaler/
│       └── main.go         # Signaling server
├── internal/
│   ├── chat/               # Chat functionality
│   │   ├── chat.go
│   │   └── room.go
│   ├── crypto/             # Encryption utilities
│   │   ├── aes.go
│   │   └── key_exchange.go
│   ├── discovery/          # Peer discovery
│   │   └── broadcast.go
│   └── network/
│       └── transfer/       # File transfer logic
│           └── transfer.go
└── pkg/
    └── utils/              # Utility functions
```

### Building from Source

```bash
# Build main application
go build -o p2p .

# Build signaling server
go build -o signaler ./cmd/signaler

# Run tests
go test ./...
```

## Troubleshooting

### Common Issues

1. **Connection Refused**: Ensure the signaling server is running and accessible
2. **NAT Traversal Issues**: For internet-wide communication, ensure proper port forwarding or use STUN/TURN servers (future enhancement)
3. **Firewall Blocking**: Allow the application through your firewall for P2P connections

### Debug Mode

Add verbose logging by modifying the source code to include debug prints in the relevant modules.

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Disclaimer

This tool is for educational and personal use. Ensure compliance with local laws and regulations regarding peer-to-peer communication and file sharing.