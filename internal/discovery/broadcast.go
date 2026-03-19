package discovery

import (
	"bufio"
	"fmt"
	"net"
	"strings"
)

// ConnectToSignaling reaches out to the matchmaker, registers itself, and learns about others
func ConnectToSignaling(serverIP, username string, myP2PPort int, onPeerDiscovered func(ip, port, username string)) {
	// Connect to the central signaling server
	conn, err := net.Dial("tcp", serverIP+":8080")
	if err != nil {
		fmt.Println("\n[System Error] Could not reach Signaling Server at", serverIP)
		return
	}
	defer conn.Close()

	// 1. Register ourselves with our specific P2P TCP port
	registration := fmt.Sprintf("REGISTER|%s|%d\n", username, myP2PPort)
	conn.Write([]byte(registration))

	// 2. Wait for the server to send us the list of active peers
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		msg := scanner.Text()
		parts := strings.Split(msg, "|")
		
		if len(parts) == 3 && parts[0] == "PEER" {
			peerUsername := parts[1]
			peerAddr := parts[2]
			
			addrParts := strings.Split(peerAddr, ":")
			if len(addrParts) == 2 {
				// Pass the discovered peer back to the Room Manager to establish the E2E tunnel
				onPeerDiscovered(addrParts[0], addrParts[1], peerUsername)
			}
		}
	}
}