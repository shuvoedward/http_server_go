package main

import (
	"context"
	"fmt"
	"net"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

var (
	activeConns = make(map[net.Conn]struct{}) // to track connections
	mu          sync.Mutex                    // to protect map
)

type header map[string]string

func main() {

	var wg sync.WaitGroup

	globalMW := []Middleware{
		RecoverMiddleware,
		LoggingMiddleware,
	}

	Handle("GET", "/", Chain(globalMW, handleIndex))
	Handle("GET", "/hello", handleHello)
	Handle("POST", "/submit", handleSubmit)
	//routes["GET"]["/"] = handler1
	//routes["GET"]["/"] = handler2

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

		<-ctx.Done() // Wait for SIGINT/SIGTERM

		fmt.Println("\nShutting down gracefully...")

		// 1. Stop accepting new connections
		listener.Close()

		fmt.Println("Waiting for active connections to drain...")
		done := make(chan struct{}) // channel to signal when WaitGroup is done
		go func() {
			wg.Wait()
			close(done)
		}()

		// 3. Implement a timeout for waiting for connections
		select {
		case <-done:
			fmt.Println("All active connection drained")
		case <-time.After(5 * time.Second): // giving 5 seconds for existing requests to complete
			fmt.Println("Timeout waiting for conenctions to drain. Forcibly closing remaining.")
		}

		// 4. Forcefully close any remaining active connections
		mu.Lock()
		for conn := range activeConns {
			conn.Close()
		}
		mu.Unlock()
		fmt.Println("Server shutdown complete.")
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

		wg.Add(1)
		// Handle client connection in goroutine
		go handleClient(conn, &wg)
		// wg is passed as a pointer because wg should be shared and
		// modified by multiple goroutines.
	}
}

func handleClient(conn net.Conn, wg *sync.WaitGroup) {
	defer func() {
		conn.Close()
		mu.Lock()
		delete(activeConns, conn)
		mu.Unlock()
		wg.Done()
	}()

	req, err := ParseRequest(conn)
	if err != nil {
		// Bad request
		rw := NewResponseWriter(conn)
		rw.WriteHeader(400)
		rw.Write([]byte("Bad Request"))
		rw.Send()
		return
	}

	rw := NewResponseWriter(conn)

	// Routing
	if methodMap, ok := routes[req.Method]; ok {
		if handler, ok := methodMap[req.Path]; ok {
			handler(rw, req)
		} else {
			// path not found
			rw.WriteHeader(404)
			rw.Write([]byte("Not Found"))
		}
	} else {
		// method now allowed
		rw.WriteHeader(405)
		rw.Write([]byte("Method Not Allowed"))
	}

	// Send whatever the handler wrote
	rw.Send()

}
