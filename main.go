// main.go
package main

import (
	"fmt"
	"os"

	"p2p-share/internal/chat"
	"p2p-share/internal/discovery"
	"p2p-share/internal/transfer"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		return
	}

	command := os.Args[1]

	switch command {
	case "scan":
		// Updated to use the new callback-based scanner
		fmt.Println("Scanning network... (Press Ctrl+C to stop)")
		discovery.ScanNetwork(func(ip, username string) {
			fmt.Printf("Found device: %s (User: %s)\n", ip, username)
		})

	case "listen":
		// Original 1-on-1 listener
		if len(os.Args) < 3 {
			fmt.Println("Usage: p2p listen <username>")
			return
		}
		username := os.Args[2]
		fmt.Printf("Starting P2P node as '%s'...\n", username)
		
		go discovery.StartBroadcaster(username)
		go chat.StartChatServer()  // Listens on 9997 for private chats
		transfer.StartFileServer() // Listens on 9998 for files (blocks main thread)

	case "send":
		// Original file transfer
		if len(os.Args) < 4 {
			fmt.Println("Usage: p2p send <IP> <file>")
			return
		}
		transfer.SendFile(os.Args[2], os.Args[3])

	case "chat":
		// Original 1-on-1 private chat
		if len(os.Args) < 3 {
			fmt.Println("Usage: p2p chat <IP>")
			return
		}
		chat.StartChatClient(os.Args[2])

	case "room":
		// The new decentralized LAN common room
		if len(os.Args) < 3 {
			fmt.Println("Usage: p2p room <username>")
			return
		}
		username := os.Args[2]
		
		go discovery.StartBroadcaster(username)
		go discovery.ScanNetwork(chat.HandleNewDiscovery)
		chat.StartRoom(username) // Listens on 9996 for mesh network

	default:
		printUsage()
	}
}

func printUsage() {
	fmt.Println("P2P File Sharing & Chat Tool")
	fmt.Println("Commands:")
	fmt.Println("  listen <user>      - Listen for private 1-on-1 chats and files")
	fmt.Println("  scan               - Scan local network for active peers")
	fmt.Println("  send <IP> <file>   - Send a file to a specific peer")
	fmt.Println("  chat <IP>          - Start a private, encrypted 1-on-1 chat")
	fmt.Println("  room <user>        - Join the encrypted LAN common room")
}