package transfer

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
)

const FilePort = ":9998"

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

	// Reading the length of the filename
	var nameLen int64
	binary.Read(conn, binary.LittleEndian, &nameLen)

	// Read the actual filename
	nameBytes := make([]byte, nameLen)
	conn.Read(nameBytes)
	filename := string(nameBytes)

	// Read the file size
	var fileSize int64
	binary.Read(conn, binary.LittleEndian, &fileSize)


	// Create the file and copy the data
	outFile, err := os.Create("received_"+filename)
	if err != nil {
		fmt.Println("Error Creating the file:", err)
		return
	}

	defer outFile.Close()

	io.CopyN(outFile, conn, fileSize)
	fmt.Println("File Received Successfully.")
}


//This function connects to a specfic peer and then sends a file
func SendFile(targetIP string, filePath string) {
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Println("File not found:", err)
		return
	}
	defer file.Close()

	stat, _ := file.Stat()
	filename := filepath.Base(filePath)

	conn, err := net.Dial("tcp", targetIP+FilePort)
	if err != nil {
		fmt.Println("Failed to connect to peer:", err)
		return
	}
	defer conn.Close()

	fmt.Printf("Sending %s to %s...\n", filename, targetIP)

	// Send filename length
	binary.Write(conn, binary.LittleEndian, int64(len(filename)))
	// Send filename
	conn.Write([]byte(filename))
	// Send file size
	binary.Write(conn, binary.LittleEndian, stat.Size())

	// Send file data
	io.Copy(conn, file)
	fmt.Println("Transfer complete!")
}