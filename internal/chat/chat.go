// internal/chat/chat.go
package chat

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"io"
	"encoding/binary"
	"p2p-share/internal/crypto"
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

// internal/chat/chat.go
// (Add "encoding/binary" and "p2p-share/internal/crypto" to your imports)

func handleChatSession(conn net.Conn) {
	defer conn.Close()

	fmt.Println("[Negotiating secure connection...]")

	// --- 1. KEY EXCHANGE HANDSHAKE ---
	
	// Generate our local ECDH keys
	privKey, pubKeyBytes, err := crypto.GenerateKeyPair()
	if err != nil {
		fmt.Println("Error generating keys:", err)
		return
	}

	// Send our public key length, then the key itself
	binary.Write(conn, binary.LittleEndian, int64(len(pubKeyBytes)))
	conn.Write(pubKeyBytes)

	// Receive the peer's public key length, then the key itself
	var peerKeyLen int64
	binary.Read(conn, binary.LittleEndian, &peerKeyLen)
	
	peerPubKeyBytes := make([]byte, peerKeyLen)
	io.ReadFull(conn, peerPubKeyBytes)

	// Compute the shared AES key
	aesKey, err := crypto.ComputeSharedSecret(privKey, peerPubKeyBytes)
	if err != nil {
		fmt.Println("Secure connection failed:", err)
		return
	}

	fmt.Println("[Secure connection established! Messages are End-to-End Encrypted]")

	// --- 2. SECURE CHAT LOOP ---

	// Goroutine to read and DECRYPT incoming messages
	go func() {
		scanner := bufio.NewScanner(conn)
		for scanner.Scan() {
			encryptedMsg := scanner.Text()
			
			// Decrypt the message
			decryptedMsg, err := crypto.Decrypt(encryptedMsg, aesKey)
			if err != nil {
				fmt.Println("\n[Failed to decrypt incoming message]")
				continue
			}
			
			fmt.Printf("\rPeer: %s\nYou: ", decryptedMsg)
		}
		fmt.Println("\n[Peer disconnected]")
		os.Exit(0)
	}()

	// Main loop to ENCRYPT and send outgoing messages
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("You: ")
		if !scanner.Scan() {
			break
		}
		msg := scanner.Text()
		
		// Encrypt before sending over the wire
		encryptedMsg, err := crypto.Encrypt(msg, aesKey)
		if err != nil {
			fmt.Println("Encryption error:", err)
			continue
		}
		
		conn.Write([]byte(encryptedMsg + "\n"))
	}
}