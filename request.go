package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
)

type Request struct {
	Method  string
	Path    string
	Version string
	Headers header
	// QueryParams map[string][]string
	Body []byte
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

	// rawPath := parts[1]
	// url.Parse()

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
