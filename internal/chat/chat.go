// internal/chat/chat.go
package chat

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

const ChatPort = ":9997"

// StartChatServer listens for incoming chat requests
func StartChatServer() {
	listener, err := net.Listen("tcp", ChatPort)
	if err != nil {
		fmt.Println("Failed to start chat server:", err)
		return
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		fmt.Printf("\n[Chat session started with %s]\n", conn.RemoteAddr().String())
		handleChatSession(conn)
	}
}

// StartChatClient initiates a chat with a peer
func StartChatClient(targetIP string) {
	conn, err := net.Dial("tcp", targetIP+ChatPort)
	if err != nil {
		fmt.Println("Failed to connect to peer for chat:", err)
		return
	}
	fmt.Printf("Connected to %s. Type your messages:\n", targetIP)
	handleChatSession(conn)
}

func handleChatSession(conn net.Conn) {
	defer conn.Close()

	// Goroutine to read incoming messages
	go func() {
		scanner := bufio.NewScanner(conn)
		for scanner.Scan() {
			fmt.Printf("\rPeer: %s\nYou: ", scanner.Text())
		}
		fmt.Println("\n[Peer disconnected]")
		os.Exit(0)
	}()

	// Main loop to send outgoing messages
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("You: ")
		if !scanner.Scan() {
			break
		}
		msg := scanner.Text()
		conn.Write([]byte(msg + "\n"))
	}
}