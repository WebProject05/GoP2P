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
		discovery.ScanNetwork()
	case "listen":
		// Starts background listeners for discovery, chat, and files
		fmt.Println("Starting P2P node...")
		go discovery.StartBroadcaster()
		go chat.StartChatServer()
		transfer.StartFileServer() // Blocks the main thread
	case "send":
		if len(os.Args) < 4 {
			fmt.Println("Usage: p2p send <IP> <file>")
			return
		}
		transfer.SendFile(os.Args[2], os.Args[3])
	case "chat":
		if len(os.Args) < 3 {
			fmt.Println("Usage: p2p chat <IP>")
			return
		}
		chat.StartChatClient(os.Args[2])
	default:
		printUsage()
	}
}

func printUsage() {
	fmt.Println("P2P File Sharing Tool")
	fmt.Println("Commands:")
	fmt.Println("  listen             - Start listening for peers, files, and chats")
	fmt.Println("  scan               - Scan local network for peers")
	fmt.Println("  send <IP> <file>   - Send a file to a peer")
	fmt.Println("  chat <IP>          - Start an ephemeral chat with a peer")
}