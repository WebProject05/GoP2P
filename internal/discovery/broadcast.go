package discovery

import (
	"fmt"
	"net"
	"time"
)

const (
	DiscoveryPort = ":9999"
	BroadCastMsg  = "P2P_HELLO"
)

func ScanNetwork() {
	// Converts a string like ":9999" into a usable UDP address (IP + port) for networking
	addr, err := net.ResolveUDPAddr("udp", DiscoveryPort)
	if err != nil {
		fmt.Println("Error resolving UDP address:", err)
		return
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		fmt.Println("Error listening for discovery:", err)
		return
	}

	defer conn.Close()

	fmt.Println("Scanning the network......")
	peers := make(map[string]string)
	buffer := make([]byte, 1024)

	for {
		// Sets a 5 second timeout for read operations on the connection (prevents blocking forever)
		conn.SetReadDeadline(time.Now().Add(5 * time.Second))
		n, remoteAddr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			// If the Timeout has reached then stop scanning for peers...
			break
		}

		msg := string(buffer[:n])
		ip := remoteAddr.IP.String()

		if msg == BroadCastMsg && peers[ip] == "" {
			peers[ip] = "Active"
			fmt.Printf("Found Device: %s \n", ip)
		}
	}

	if len(peers) == 0 {
		fmt.Println("No devices found...")
	}

}


// The below function will provide the network it's avaliability
func StartBroadcaster() {
	conn, err := net.Dial("udp", "255.255.255.255"+DiscoveryPort)
	if err != nil {
		fmt.Println("BroadCast failed:", err)
		return
	}

	defer conn.Close()

	for {
		conn.Write([]byte(BroadCastMsg))
		time.Sleep(3 * time.Second)
	}
}