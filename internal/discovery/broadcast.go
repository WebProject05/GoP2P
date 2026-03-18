package discovery

import (
	"fmt"
	"net"
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

	conn, err := net.ListenUDP("udp", addr);
	if err != nil {
		fmt.Println("Error listening for discovery:", err);
		return
	}

	defer conn.Close()

	fmt.Println("Scanning the network......")
	peers := make(map[string]string)
	buffer := make([]byte, 1024)

	
	
}
