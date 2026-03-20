// internal/transfer/transfer.go
package transfer

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
)

const (
	FilePort  = ":9998"
	ChunkSize = 32 * 1024 // 32KB chunks
)

// =====================================================================
// PART 1: LEGACY MANUAL IP TRANSFERS (Used by 'p2p send' and 'p2p listen')
// =====================================================================

func StartFileServer() {
	listener, err := net.Listen("tcp", FilePort)
	if err != nil {
		fmt.Println("Failed to start the File server:", err)
		return
	}
	defer listener.Close()
	fmt.Println("Listening for file transfers on port 9998...")

	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		go handleIncomingFile(conn)
	}
}

func handleIncomingFile(conn net.Conn) {
	defer conn.Close()

	var nameLen int64
	binary.Read(conn, binary.LittleEndian, &nameLen)

	nameBytes := make([]byte, nameLen)
	io.ReadFull(conn, nameBytes)
	filename := string(nameBytes)

	var fileSize int64
	binary.Read(conn, binary.LittleEndian, &fileSize)

	fmt.Printf("\nReceiving file: %s (%d bytes)...\n", filename, fileSize)

	outFile, err := os.Create("received_" + filename)
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	defer outFile.Close()

	receivedBytes := int64(0)
	buffer := make([]byte, ChunkSize)

	for receivedBytes < fileSize {
		n, err := conn.Read(buffer)
		if err != nil && err != io.EOF {
			fmt.Println("\nError reading chunk:", err)
			break
		}
		if n == 0 {
			break
		}

		outFile.Write(buffer[:n])
		receivedBytes += int64(n)

		progress := float64(receivedBytes) / float64(fileSize) * 100
		fmt.Printf("\rProgress: %.2f%%", progress)
	}
	fmt.Println("\nFile received successfully!")
}

func SendFile(targetIP string, filePath string) {
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Println("File not found:", err)
		return
	}
	defer file.Close()

	stat, _ := file.Stat()
	fileSize := stat.Size()
	filename := filepath.Base(filePath)

	conn, err := net.Dial("tcp", targetIP+FilePort)
	if err != nil {
		fmt.Println("Failed to connect to peer:", err)
		return
	}
	defer conn.Close()

	fmt.Printf("Sending %s to %s...\n", filename, targetIP)

	binary.Write(conn, binary.LittleEndian, int64(len(filename)))
	conn.Write([]byte(filename))
	binary.Write(conn, binary.LittleEndian, fileSize)

	sentBytes := int64(0)
	buffer := make([]byte, ChunkSize)

	for {
		n, err := file.Read(buffer)
		if err != nil && err != io.EOF {
			fmt.Println("\nError reading file:", err)
			break
		}
		if n == 0 {
			break
		}

		conn.Write(buffer[:n])
		sentBytes += int64(n)

		progress := float64(sentBytes) / float64(fileSize) * 100
		fmt.Printf("\rProgress: %.2f%%", progress)
	}
	fmt.Println("\nTransfer complete!")
}

// =====================================================================
// PART 2: NEW MESH NETWORK TRANSFERS (Used by the UI /send command)
// =====================================================================

// ServeFile opens a dynamic port for the mesh room to pull a file securely
func ServeFile(filePath string, onProgress func(int)) (int, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return 0, err
	}

	stat, _ := file.Stat()
	fileSize := stat.Size()

	listener, err := net.Listen("tcp", ":0") // :0 assigns a random dynamic port
	if err != nil {
		return 0, err
	}
	port := listener.Addr().(*net.TCPAddr).Port

	go func() {
		defer file.Close()
		defer listener.Close()

		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer conn.Close()

		sendWithProgress(file, conn, fileSize, onProgress)
	}()

	return port, nil
}

// FetchFile connects to the dynamic port offered in the mesh room
func FetchFile(addr string, saveName string, fileSize int64, onProgress func(int)) error {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return err
	}
	defer conn.Close()

	os.MkdirAll("downloads", os.ModePerm)
	outFile, err := os.Create(filepath.Join("downloads", saveName))
	if err != nil {
		return err
	}
	defer outFile.Close()

	sendWithProgress(conn, outFile, fileSize, onProgress)
	return nil
}

// sendWithProgress is a helper for the Mesh transfers to update the UI progress bar
func sendWithProgress(src io.Reader, dst io.Writer, totalSize int64, update func(int)) {
	buf := make([]byte, ChunkSize)
	var totalSent int64

	for {
		n, err := src.Read(buf)
		if n > 0 {
			dst.Write(buf[:n])
			totalSent += int64(n)
			if totalSize > 0 {
				percentage := int((float64(totalSent) / float64(totalSize)) * 100)
				update(percentage)
			}
		}
		if err == io.EOF {
			update(100)
			break
		}
		if err != nil {
			break
		}
	}
}