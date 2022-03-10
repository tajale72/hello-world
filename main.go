package main

import (
	"io"
	"net/http"
	"os"
)

func main() {
	port := os.Getenv("PORT")
	http.HandleFunc("/", Hello)
	http.ListenAndServe(port, nil)
}

func Hello(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "Hello World")
}
