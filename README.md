# Custom Go HTTP Server

This project is a foundational HTTP server implemented from scratch in Go, providing a deep dive into the core mechanics of web serving without relying on Go's standard `net/http` package for the primary server logic. It demonstrates fundamental networking concepts, HTTP protocol understanding, concurrent request handling, and robust server management practices like graceful shutdown.

## Table of Contents

- [Features](#features)
- [Concepts Demonstrated](#concepts-demonstrated)
- [Project Structure](#project-structure)
- [How to Run](#how-to-run)
- [API Endpoints](#api-endpoints)
- [Future Improvements](#future-improvements)
- [License](#license)

## Features

* **Low-Level TCP Server:** Built directly on Go's `net` package for raw TCP socket communication.
* **Custom HTTP Request Parsing:** Manually parses HTTP request lines and headers (e.g., Method, Path, Version, Headers).
* **Custom HTTP Response Writing:** Implements a `ResponseWriter` abstraction for building and sending HTTP responses, including status codes and headers.
* **Request Routing:** Dispatches incoming requests to specific handler functions based on HTTP method and URL path.
* **Middleware Chaining:** Supports a flexible middleware system to add cross-cutting concerns (like logging and panic recovery) to handlers.
* **Concurrent Request Handling:** Utilizes Go goroutines to handle multiple client connections simultaneously.
* **Graceful Shutdown:** Implements a robust mechanism to:
    * Stop accepting new connections upon signal (SIGINT/SIGTERM).
    * Allow active connections/requests to complete within a timeout.
    * Forcefully close any remaining active connections after the timeout.
* **Concurrency Safety:** Employs `sync.Mutex` to protect shared resources (active connections map) from race conditions.
* **Panic Recovery:** Middleware to catch panics in handlers and return a 500 Internal Server Error.
* **Request Logging:** Middleware to log incoming requests, their status, and response time.

## Concepts Demonstrated

This project showcases a strong understanding of various Go and server-side programming concepts:

* **Go's `net` Package:** Direct use of `net.Listen()`, `listener.Accept()`, and `net.Conn` for network programming.
* **Interfaces:** Extensive use of Go interfaces (`io.Reader`, `net.Listener`, `net.Conn`) for abstraction and polymorphism.
* **Goroutines and Concurrency:** Practical application of goroutines for concurrent request processing.
* **`sync.WaitGroup`:** Essential for coordinating and waiting for goroutines during graceful shutdown.
* **`sync.Mutex`:** Protecting shared data structures (maps) from concurrent access.
* **`context.Context`:** Used for signaling cancellation and managing timeouts during shutdown.
* **`defer` statements:** For ensuring resource cleanup (like closing connections) in goroutines.
* **HTTP Protocol Basics:** Understanding request/response structure, methods, paths, headers, and status codes.
* **Server Architecture:** Concepts of listeners, accept loops, request/response cycles, routing, and middleware.

## Project Structure

The code is organized into several files within the `main` package for clarity and modularity: