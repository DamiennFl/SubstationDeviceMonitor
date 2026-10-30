package main

import (
	"log"
	"net/http"
)

// When localhost:8080/health is accessed with a GET request, it will print OK.
// TODO: Update this for Kubernetes
func health(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", health)        // Only allow GET requests to /health
	log.Fatal(http.ListenAndServe(":8080", mux)) // TODO: Use Kubernetes environment variable port
}
