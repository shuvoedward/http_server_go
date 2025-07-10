package main

import (
	"fmt"
	"net"
	"net/http"
	"strconv"
)

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
