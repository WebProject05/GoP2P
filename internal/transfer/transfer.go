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

// This function listens for incoming file transfers
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

	// Track progress during receipt
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

		// Calculate and print percentage
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

	// Send in chunks and track progress
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