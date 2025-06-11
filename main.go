package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
)

var (
	activeConns = make(map[net.Conn]struct{}) // to track connections
	mu          sync.Mutex                    // to protect map
)

type header map[string]string
type Request struct {
	Method  string
	Path    string
	Version string
	Headers header
	// QueryParams map[string][]string
	Body []byte
}

type ResponseWriter struct {
	Conn    net.Conn
	Headers header
	Status  int
	Body    []byte
}

func NewResponseWriter(c net.Conn) *ResponseWriter {
	return &ResponseWriter{
		Conn:    c,
		Headers: header{},
	}
}

func (w *ResponseWriter) Header(key, val string) {
	w.Headers[key] = val
}

func (w *ResponseWriter) WriteHeader(status int) {
	w.Status = status
}

func (w *ResponseWriter) Write(body []byte) int {
	w.Body = append(w.Body, body...)
	return len(body)
}

func (w *ResponseWriter) Send() error {
	// 1. Ensure status code, default to 200 OK
	if w.Status == 0 {
		w.Status = 200
	}

	// 2. Textual phrase for status
	statusText := http.StatusText(w.Status)

	// 3. Building the response head status line
	// e.g., "HTTP/1.1 200 OK"
	head := fmt.Sprintf("HTTP/1.1 %d %s\r\n", w.Status, statusText)

	w.Headers["Content-Length"] = strconv.Itoa(len(w.Body))

	// Serialize each header
	for name, value := range w.Headers {
		head += fmt.Sprintf("%s: %s\r\n", name, value)
	}

	// Separate headers from the body
	head += "\r\n"

	// Write header
	if _, err := w.Conn.Write([]byte(head)); err != nil {
		return err
	}

	// Write body
	if len(w.Body) > 0 {
		if _, err := w.Conn.Write([]byte(w.Body)); err != nil {
			return err
		}
	}

	return nil

}

type HandlerFunc func(w *ResponseWriter, r *Request)

var routes = map[string]map[string]HandlerFunc{
	"GET":  {},
	"POST": {},
}

func Handle(method, path string, h HandlerFunc) {
	if routes[method] == nil {
		routes[method] = make(map[string]HandlerFunc)
	}
	routes[method][path] = h
}

func main() {
	Handle("GET", "/", handleIndex)
	Handle("GET", "/hello", handleHello)
	Handle("POST", "/submit", handleSubmit)

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

func ParseRequest(conn net.Conn) (*Request, error) {

	// 1. Buffer to read data
	reader := bufio.NewReader(conn) // buffered input output []bytes

	// 2. Read the request line (first line) "GET / HTTP/1.1\r\n"
	requestLine, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Error reading request line:", err)
		return nil, err
	}

	requestLine = strings.TrimRight(requestLine, "\r\n") // GET / HTTP/1.1

	parts := strings.Split(requestLine, " ") // [GET, /, HTTP/1.1]
	if len(parts) < 3 {
		fmt.Println("Malformed request line:", requestLine)
		return nil, err
	}

	method, path, version := parts[0], parts[1], parts[2]
	fmt.Println("Method: ", method)
	fmt.Println("Path: ", path)
	fmt.Println("Version: ", version)

	// Read Headers
	headers := make(header)
	for {

		line, err := reader.ReadString('\n')
		fmt.Println(line)
		if err != nil {
			fmt.Println("Error reading header:", err)
			return nil, err
		}

		line = strings.TrimRight(line, "\r\n")
		if line == "" {
			break
		}

		// Parse header line: Key: Value
		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			headers[key] = value
			fmt.Printf("Header: %s = %s\n", key, value)
		} else {
			fmt.Println("Malformed header line:", line)
		}

	}
	// Body
	var requestBody string
	if contentLegnthStr, ok := headers["Content-Length"]; ok {
		contentLegnth, err := strconv.Atoi(contentLegnthStr)
		if err != nil {
			fmt.Println("Error parsing Content-Length:", err)
			// might send 400 bad request
			return nil, err
		}

		if contentLegnth > 0 {
			bodyBytes := make([]byte, contentLegnth)
			n, err := io.ReadFull(reader, bodyBytes) // Read exactly content length
			if err != nil {
				fmt.Println("Error reading request body:", err)
				return nil, err
			}

			if n != contentLegnth {
				fmt.Println("Did not read full body:", n, "byte read, expected", contentLegnth)
				return nil, err
			}

			requestBody = string(bodyBytes)
			fmt.Println("Request body:", requestBody)
		}
	}

	return &Request{
		Method:  method,
		Path:    path,
		Version: version,
		Headers: headers,
		Body:    []byte(requestBody),
	}, nil
}

func handleIndex(w *ResponseWriter, r *Request) {
	w.Header("Content-Type", "text/plain")
	w.Write([]byte("Welcome to my Go server!\n"))
}

func handleHello(w *ResponseWriter, r *Request) {
	w.Header("Content-Type", "text/plain")
	w.Write([]byte("Hello"))
}

func handleSubmit(w *ResponseWriter, r *Request) {
	w.Header("Content-Type", "text/plain")
	w.WriteHeader(201)
	w.Write([]byte("Recieved"))
}
