package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"sync"
)

var (
	peers      = make(map[string]string) // Map Address -> Username
	peersMutex sync.RWMutex
)

func main() {
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Println("Error starting signaling server:", err)
		return
	}
	defer listener.Close()
	fmt.Println("Signaling Server running on port 8080...")

	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		go handleSignaling(conn)
	}
}

func handleSignaling(conn net.Conn) {
	defer conn.Close()
	scanner := bufio.NewScanner(conn)

	if !scanner.Scan() {
		return
	}
	msg := scanner.Text()
	parts := strings.Split(msg, "|")
	
	if len(parts) != 3 || parts[0] != "REGISTER" {
		return
	}

	username := parts[1]
	p2pPort := parts[2]
	
	publicIP := strings.Split(conn.RemoteAddr().String(), ":")[0]
	publicAddr := fmt.Sprintf("%s:%s", publicIP, p2pPort)

	// Send list of active peers to the new user
	peersMutex.RLock()
	for existingAddr, existingUser := range peers {
		conn.Write([]byte(fmt.Sprintf("PEER|%s|%s\n", existingUser, existingAddr)))
	}
	peersMutex.RUnlock()

	// Register the new user
	peersMutex.Lock()
	peers[publicAddr] = username
	peersMutex.Unlock()

	fmt.Printf("[Registry] Registered %s at %s\n", username, publicAddr)
}