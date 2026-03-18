package chat

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"sync"

	"p2p-share/internal/crypto"
)

const RoomPort = ":9996"

type RoomPeer struct {
	Username string
	Conn     net.Conn
	AESKey   []byte
}

var (
	activePeers = make(map[string]*RoomPeer) // Map IP to RoomPeer
	peerMutex   sync.RWMutex
	myUsername  string
)

// StartRoom Node initialization for the common room
func StartRoom(username string) {
	myUsername = username

	// 1. Start listening for incoming mesh connections
	go listenForRoomPeers()

	// 2. Start terminal input loop for broadcasting messages
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Printf("Joined Common Room as '%s'. Type to broadcast...\n", myUsername)
	
	for {
		if !scanner.Scan() {
			break
		}
		msg := scanner.Text()
		broadcastToRoom(msg)
	}
}

// HandleNewDiscovery is called by the UDP scanner when a peer announces itself
func HandleNewDiscovery(ip, peerUsername string) {
	peerMutex.Lock()
	defer peerMutex.Unlock()

	// Check for username collisions
	for existingIP, peer := range activePeers {
		if peer.Username == peerUsername && existingIP != ip {
			fmt.Printf("\n[System] Rejected connection from %s: Username '%s' already in use.\n", ip, peerUsername)
			return
		}
	}

	// If we aren't already connected to this IP, establish the secure mesh link
	if _, exists := activePeers[ip]; !exists {
		go connectToRoomPeer(ip, peerUsername)
	}
}

func connectToRoomPeer(ip, peerUsername string) {
	conn, err := net.Dial("tcp", ip+RoomPort)
	if err != nil {
		return
	}

	aesKey, err := performECDHHandshake(conn)
	if err != nil {
		conn.Close()
		return
	}

	addPeerToRoom(ip, peerUsername, conn, aesKey)
}

func listenForRoomPeers() {
	listener, _ := net.Listen("tcp", RoomPort)
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}

		// When receiving a connection, we must do the handshake, but we 
		// get the IP. We will learn their username from their first encrypted message 
		// or rely on the UDP broadcast state. For simplicity, we trust the connection.
		aesKey, err := performECDHHandshake(conn)
		if err != nil {
			conn.Close()
			continue
		}

		ip := strings.Split(conn.RemoteAddr().String(), ":")[0]
		addPeerToRoom(ip, "Unknown (Syncing...)", conn, aesKey)
	}
}

func addPeerToRoom(ip, username string, conn net.Conn, aesKey []byte) {
	peerMutex.Lock()
	activePeers[ip] = &RoomPeer{
		Username: username,
		Conn:     conn,
		AESKey:   aesKey,
	}
	peerMutex.Unlock()

	fmt.Printf("\n[System] %s joined the room.\n", username)
	go receiveFromPeer(ip)
}

func receiveFromPeer(ip string) {
	peerMutex.RLock()
	peer := activePeers[ip]
	peerMutex.RUnlock()

	scanner := bufio.NewScanner(peer.Conn)
	for scanner.Scan() {
		encryptedMsg := scanner.Text()
		decryptedMsg, err := crypto.Decrypt(encryptedMsg, peer.AESKey)
		
		if err == nil {
			// Clear current line, print message, and restore prompt
			fmt.Printf("\r[%s]: %s\n> ", peer.Username, decryptedMsg)
		}
	}

	// If the loop breaks, the peer disconnected
	peerMutex.Lock()
	delete(activePeers, ip)
	peerMutex.Unlock()
	fmt.Printf("\n[System] %s left the room.\n> ", peer.Username)
}

func broadcastToRoom(message string) {
	peerMutex.RLock()
	defer peerMutex.RUnlock()

	// Iterate through every active peer and encrypt the message specifically for them
	for _, peer := range activePeers {
		encryptedMsg, err := crypto.Encrypt(message, peer.AESKey)
		if err == nil {
			peer.Conn.Write([]byte(encryptedMsg + "\n"))
		}
	}
	fmt.Print("> ")
}

// Helper function to handle the ECDH exchange (reused from your previous chat setup)
func performECDHHandshake(conn net.Conn) ([]byte, error) {
	privKey, pubKeyBytes, _ := crypto.GenerateKeyPair()
	
	binary.Write(conn, binary.LittleEndian, int64(len(pubKeyBytes)))
	conn.Write(pubKeyBytes)

	var peerKeyLen int64
	binary.Read(conn, binary.LittleEndian, &peerKeyLen)
	peerPubKeyBytes := make([]byte, peerKeyLen)
	io.ReadFull(conn, peerPubKeyBytes)

	return crypto.ComputeSharedSecret(privKey, peerPubKeyBytes)
}