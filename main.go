// main.go
package main

import (
	"bufio"
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
	case "send":
		// Direct file transfer (requires known IP)
		if len(os.Args) < 4 {
			fmt.Println("Usage: p2p send <IP> <file>")
			return
		}
		transfer.SendFile(os.Args[2], os.Args[3])

	case "chat":
		// Direct 1-on-1 private chat (requires known IP)
		if len(os.Args) < 3 {
			fmt.Println("Usage: p2p chat <IP>")
			return
		}
		chat.StartChatClient(os.Args[2])

	case "room":
		// The new Internet-ready Decentralized Mesh Room
		if len(os.Args) < 4 {
			fmt.Println("Usage: p2p room <username> <signaling_ip>")
			fmt.Println("Example: p2p room Alice 127.0.0.1")
			return
		}
		username := os.Args[2]
		signalingIP := os.Args[3]
		
		// 1. Start the local P2P listener on a dynamic port
		tcpPort := chat.InitRoom(username)
		if tcpPort == 0 {
			return
		}
		
		// 2. Reach out to the Matchmaker to find peers and punch holes
		go discovery.ConnectToSignaling(signalingIP, username, tcpPort, chat.HandleNewDiscovery)

		// 3. Handle typing in the main thread (Blocks)
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			msg := scanner.Text()
			if msg != "" {
				chat.BroadcastToRoom(msg)
				// Local echo: moves cursor up, clears line, and prints formatted message
				fmt.Printf("\033[1A\033[K%s: %s\n> ", username, msg)
			}
		}

	default:
		printUsage()
	}
}

func printUsage() {
	fmt.Println("P2P File Sharing & Chat Tool (WAN Edition)")
	fmt.Println("Commands:")
	fmt.Println("  room <user> <signaling_ip> - Join the decentralized chat mesh via Matchmaker")
	fmt.Println("  send <IP> <file>           - Send a file directly to a known peer IP")
	fmt.Println("  chat <IP>                  - Start a private 1-on-1 chat with a known IP")
}