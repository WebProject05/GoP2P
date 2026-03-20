package chat

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"p2p-share/internal/crypto"
	"p2p-share/internal/transfer"
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
			} else if strings.HasPrefix(decryptedMsg, "__FILE_OFFER__|") { // <--- THIS IS THE FIX

				// Parse the incoming hidden file offer
				parts := strings.Split(decryptedMsg, "|")
				if len(parts) == 4 {
					filename := parts[1]
					filesize := parts[2]
					port := parts[3]

					// Extract the sender's base IP from their chat connection
					ip := strings.Split(peer.Conn.RemoteAddr().String(), ":")[0]
					targetAddr := fmt.Sprintf("%s:%s", ip, port)

					AddSystemMessage(fmt.Sprintf("%s is sending '%s'. Downloading...", peer.Username, filename))

					// Start the download in the background
					go func() {
						var size int64
						fmt.Sscanf(filesize, "%d", &size)

						err := transfer.FetchFile(targetAddr, filename, size, func(pct int) {
							UpdateTransferUI(filename, pct, false)
						})

						if err != nil {
							AddSystemMessage("Failed to download " + filename)
						} else {
							AddSystemMessage("Download complete: " + filename)
						}
					}()
				}
			} else if strings.HasPrefix(decryptedMsg, "__PRIVATE__|") {
				//Catching private messages
				ClearTyping(peer.Username)
				actualMsg := strings.TrimPrefix(decryptedMsg, "__PRIVATE__|")
				// Displaying as a highlighted system message so it stands out
				AddSystemMessage(fmt.Sprintf("(Whisper from %s): %s", peer.Username, actualMsg))
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

func SendFileToPeer(targetUser, filePath string) {
	peerMutex.RLock()
	var targetPeer *RoomPeer
	for _, p := range activePeers {
		if p.Username == targetUser {
			targetPeer = p
			break
		}
	}
	peerMutex.RUnlock()

	if targetPeer == nil {
		AddSystemMessage("User '" + targetUser + "' not found.")
		return
	}

	stat, err := os.Stat(filePath)
	if err != nil {
		AddSystemMessage("File not found: " + filePath)
		return
	}

	filename := filepath.Base(filePath)

	// Open the local dynamic port
	port, err := transfer.ServeFile(filePath, func(pct int) {
		UpdateTransferUI(filename, pct, true)
	})

	if err != nil {
		AddSystemMessage("Failed to start transfer server.")
		return
	}

	// Send the hidden control packet to the specific user
	offerMsg := fmt.Sprintf("__FILE_OFFER__|%s|%d|%d", filename, stat.Size(), port)
	encryptedMsg, _ := crypto.Encrypt(offerMsg, targetPeer.AESKey)
	targetPeer.Conn.Write([]byte(encryptedMsg + "\n"))

	AddSystemMessage("Offering file '" + filename + "' to " + targetUser + "...")
}

func SendPrivateMessage(targetUser, message string) {
	peerMutex.RLock()
	var targetPeer *RoomPeer
	for _, p := range activePeers {
		if p.Username == targetUser {
			targetPeer = p
			break
		}
	}
	peerMutex.RUnlock()

	if targetPeer == nil {
		AddSystemMessage("User '" + targetUser + "' not found.")
		return
	}

	// Prefix the message so the receiver knows it's private
	privateMsg := "__PRIVATE__|" + message
	encryptedMsg, _ := crypto.Encrypt(privateMsg, targetPeer.AESKey)

	// Send ONLY to this specific peer
	targetPeer.Conn.Write([]byte(encryptedMsg + "\n"))

	// Print it to our own screen in a special color
	AddSystemMessage(fmt.Sprintf("(Whisper to %s): %s", targetUser, message))
}

// ShowActivePeers formats and displays the IP addresses of everyone in the room
func ShowActivePeers() {
	peerMutex.RLock()
	defer peerMutex.RUnlock()

	if len(activePeers) == 0 {
		AddSystemMessage("No other peers are currently in the room.")
		return
	}

	AddSystemMessage("--- Active Peers List ---")
	for addr, peer := range activePeers {
		// The addr string usually looks like "192.168.1.5:49152"
		// We split it to just grab the IP part
		ip := strings.Split(addr, ":")[0]
		AddSystemMessage(fmt.Sprintf("%s: %s", peer.Username, ip))
	}
	AddSystemMessage("-------------------------")
}