package main

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"os/signal"
	"strings"
	"sync"
	"syscall"
)

var (
	activeConns = make(map[net.Conn]struct{}) // to track connections
	mu          sync.Mutex                    // to protect map
)

func main() {
	// Listen for incoming connection
	listener, err := net.Listen("tcp", "localhost:8080")
	// Telling the os that I only want handle tcp traffic. and listen on port 8080, this socket is passive, listener socket
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer listener.Close()

	fmt.Println("Server is listening on port 8080")

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {

		<-ctx.Done()

		fmt.Println("\nShutting down gracefully...")

		// Stop accepting new connections
		listener.Close()

		mu.Lock()
		for conn := range activeConns {
			conn.Close()
		}
		mu.Unlock()
	}()

	for {
		// Accept incoming connections
		// Creates a new socket, different from listening socket from above
		// OS creates a new connection socket for that client.
		conn, err := listener.Accept()
		if err != nil {
			select {
			case <-ctx.Done():
				// Context was cancel, exit
				fmt.Println("Server shutdown complete")
				return

			default:
				fmt.Println("Accept error:", err)
				continue
			}
		}

		mu.Lock()
		activeConns[conn] = struct{}{}
		mu.Unlock()

		// Handle client connection in goroutine
		go handleClient(conn)
	}
}

func handleClient(conn net.Conn) {
	defer func() {
		conn.Close()
		mu.Lock()
		delete(activeConns, conn)
		mu.Unlock()
	}()

	// 1. Buffer to read data
	reader := bufio.NewReader(conn) // buffered input output []bytes

	// 2. Read the request line (first line) "GET / HTTP/1.1\r\n"
	requestLine, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Error reading request line:", err)
		return
	}

	requestLine = strings.TrimRight(requestLine, "\r\n")

	parts := strings.Split(requestLine, " ")
	if len(parts) < 3 {
		fmt.Println("Malformed request line:", requestLine)
		return
	}

	method, path, version := parts[0], parts[1], parts[2]
	fmt.Println("Method: ", method)
	fmt.Println("Path: ", path)
	fmt.Println("Version: ", version)

	// Read Headers
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading header:", err)
			return
		}

		line = strings.TrimRight(line, "\r\n")
		if line == "" {
			break
		}

	}

	response := "HTTP/1.1 200 OK\r\n"
	conn.Write([]byte(response))
}
