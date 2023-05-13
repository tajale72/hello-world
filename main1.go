package main

import (
	"io"
	"net/http"
	"os"
)

func main1() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "900001" // Default port if not specified
	}
	http.HandleFunc("/hello", Hello)
	http.ListenAndServe(port, nil)
}

func Hello1(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "Hello World")

}
