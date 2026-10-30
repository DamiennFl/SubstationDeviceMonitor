package main

import (
	"fmt"
	"log"
	"net/http"
)

// handler is type http.HandlerFunc.
// ResponseWriter's value assembles the HTTP server's response, by writing to it we send data to the HTTP client.
// Request is a data structure that represents the client's HTTP request. r.URL.Path is the path component of the URL.
func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hi there, I love %s!", r.URL.Path[1:])
}

// When localhost:8080/health is accessed, it will print OK.
func health(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
}

func main() {
	http.HandleFunc("/", handler)
	http.HandleFunc("/health", health)
	log.Fatal(http.ListenAndServe(":8080", nil))
}