// internal/discovery/broadcast.go
package discovery

import (
	"fmt"
	"net"
	"strings"
	"time"
)

const DiscoveryPort = ":9999"

// StartBroadcaster now includes the username in the payload
func StartBroadcaster(username string) {
	conn, err := net.Dial("udp", "255.255.255.255"+DiscoveryPort)
	if err != nil {
		fmt.Println("Broadcast setup failed:", err)
		return
	}
	defer conn.Close()

	msg := fmt.Sprintf("P2P_HELLO|%s", username)
	for {
		conn.Write([]byte(msg))
		time.Sleep(3 * time.Second)
	}
}

// ScanNetwork now takes a callback to trigger TCP connections when a peer is found
func ScanNetwork(onPeerDiscovered func(ip string, username string)) {
	addr, _ := net.ResolveUDPAddr("udp", DiscoveryPort)
	conn, _ := net.ListenUDP("udp", addr)
	defer conn.Close()

	buffer := make([]byte, 1024)
	for {
		n, remoteAddr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			continue
		}

		payload := string(buffer[:n])
		parts := strings.Split(payload, "|")
		
		if len(parts) == 2 && parts[0] == "P2P_HELLO" {
			ip := remoteAddr.IP.String()
			peerUsername := parts[1]
			// Pass the discovered peer to the room manager
			onPeerDiscovered(ip, peerUsername)
		}
	}
}