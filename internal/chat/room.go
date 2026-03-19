package chat

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"sync"

	"p2p-share/internal/crypto"
)

type RoomPeer struct {
	Username string
	Conn     net.Conn
	AESKey   []byte
}

var (
	activePeers = make(map[string]*RoomPeer)
	peerMutex   sync.RWMutex
	myUsername  string
)

func InitRoom(username string) int {
	myUsername = username

	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		fmt.Println("\n[System Error] Could not bind to any TCP port.")
		return 0
	}

	port := listener.Addr().(*net.TCPAddr).Port
	// fmt.Printf("Joined Common Room as '%s'. Type to broadcast...\n> ", myUsername)

	go func() {
		defer listener.Close()
		for {
			conn, err := listener.Accept()
			if err != nil {
				continue
			}

			aesKey, peerName, err := performHandshake(conn, myUsername)
			if err != nil {
				conn.Close()
				continue
			}

			remoteAddr := conn.RemoteAddr().String()
			addPeerToRoom(remoteAddr, peerName, conn, aesKey)
		}
	}()

	return port
}

func HandleNewDiscovery(ip, peerPort, peerUsername string) {
	peerMutex.Lock()
	defer peerMutex.Unlock()

	peerAddr := ip + ":" + peerPort
	if peerUsername == myUsername {
		return
	}
	if _, exists := activePeers[peerAddr]; exists {
		return
	}

	go connectToRoomPeer(peerAddr)
}

func connectToRoomPeer(peerAddr string) {
	conn, err := net.Dial("tcp", peerAddr)
	if err != nil {
		return
	}

	aesKey, actualPeerName, err := performHandshake(conn, myUsername)
	if err != nil {
		conn.Close()
		return
	}

	addPeerToRoom(peerAddr, actualPeerName, conn, aesKey)
}

func addPeerToRoom(peerAddr, username string, conn net.Conn, aesKey []byte) {
	peerMutex.Lock()
	activePeers[peerAddr] = &RoomPeer{Username: username, Conn: conn, AESKey: aesKey}
	peerMutex.Unlock()

	AddSystemMessage(username + " joined the room.")
	refreshUI_Roster()
	go receiveFromPeer(peerAddr)
}

func receiveFromPeer(peerAddr string) {
	peerMutex.RLock()
	peer := activePeers[peerAddr]
	peerMutex.RUnlock()

	scanner := bufio.NewScanner(peer.Conn)
	for scanner.Scan() {
		encryptedMsg := scanner.Text()
		decryptedMsg, err := crypto.Decrypt(encryptedMsg, peer.AESKey)

		if err == nil {
			if decryptedMsg == "__TYPING__" {
				SetTyping(peer.Username)
			} else {
				ClearTyping(peer.Username)

				AddRemoteMessage(peer.Username, decryptedMsg)
			}
		}
	}

	peerMutex.Lock()
	delete(activePeers, peerAddr)
	peerMutex.Unlock()

	AddSystemMessage(peer.Username + " left the room.")
	refreshUI_Roster() // NEW: Update the side panel
}

func BroadcastToRoom(message string) {
	peerMutex.RLock()
	defer peerMutex.RUnlock()

	for _, peer := range activePeers {
		encryptedMsg, err := crypto.Encrypt(message, peer.AESKey)
		if err == nil {
			peer.Conn.Write([]byte(encryptedMsg + "\n"))
		}
	}
	// fmt.Print("> ")
}

func performHandshake(conn net.Conn, myUsername string) ([]byte, string, error) {
	binary.Write(conn, binary.LittleEndian, int64(len(myUsername)))
	conn.Write([]byte(myUsername))

	privKey, pubKeyBytes, _ := crypto.GenerateKeyPair()
	binary.Write(conn, binary.LittleEndian, int64(len(pubKeyBytes)))
	conn.Write(pubKeyBytes)

	var peerNameLen int64
	binary.Read(conn, binary.LittleEndian, &peerNameLen)
	peerNameBytes := make([]byte, peerNameLen)
	io.ReadFull(conn, peerNameBytes)
	peerName := string(peerNameBytes)

	var peerKeyLen int64
	binary.Read(conn, binary.LittleEndian, &peerKeyLen)
	peerPubKeyBytes := make([]byte, peerKeyLen)
	io.ReadFull(conn, peerPubKeyBytes)

	aesKey, err := crypto.ComputeSharedSecret(privKey, peerPubKeyBytes)
	return aesKey, peerName, err
}

func refreshUI_Roster() {
	peerMutex.RLock()
	var users []string
	users = append(users, myUsername+" (You)")
	for _, p := range activePeers {
		users = append(users, p.Username)
	}
	peerMutex.RUnlock()
	UpdateRoster(users)
}
