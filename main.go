package main

import (
	"fmt"
	"net"
)

func main() {
	// Listen for incoming connection
	listener, err := net.Listen("tcp", "localhost:8080")
	// Telling the os that I only want handle tcp traffic. and listen on port 8080
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer listener.Close()

	fmt.Println("Server is listening on port 8080")

	for {
		// Accept incoming connections
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error:", err)
			continue
		}

		// Handle client connection in goroutine
		go handleClient(conn)
	}
}

func handleClient(conn net.Conn) {
	defer conn.Close()

	// Buffer to read data
	buffer := make([]byte, 1024)

	for {
		// Read data from the client
		// n = number of bytes read into the buffer- not the size of a tcp packet.
		n, err := conn.Read(buffer)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		// Process and use the data
		message := buffer[:n]
		fmt.Printf("Recieved: %s\n", message)

		// Echo back the received message
		_, err = conn.Write(message)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
	}
}
