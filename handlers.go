package main

func handleIndex(w *ResponseWriter, r *Request) {
	w.Header("Content-Type", "text/plain")
	w.Write([]byte("Welcome to my Go server!\n"))
}

func handleHello(w *ResponseWriter, r *Request) {
	w.Header("Content-Type", "text/plain")
	w.Write([]byte("Hello\n"))
}

func handleSubmit(w *ResponseWriter, r *Request) {
	w.Header("Content-Type", "text/plain")
	w.WriteHeader(201)
	w.Write([]byte("Recieved\n"))
}
