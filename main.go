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
		if len(os.Args) < 4 {
			fmt.Println("Usage: p2p room <username> <signaling_ip>")
			return
		}
		username := os.Args[2]
		signalingIP := os.Args[3]

		tcpPort := chat.InitRoom(username)
		if tcpPort == 0 {
			return
		}

		go discovery.ConnectToSignaling(signalingIP, username, tcpPort, chat.HandleNewDiscovery)

		// Start the UI and pass the broadcast function as a callback
		chat.StartUI(username,
			func(msg string) {
				// On Enter Press: Send actual message
				chat.BroadcastToRoom(msg)
			},
			func() {
				// On Keystroke: Send hidden typing indicator
				chat.BroadcastToRoom("__TYPING__")
			},
		)

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
