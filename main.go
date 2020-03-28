package main

import (
	"fmt"
	"net/http"
)

func healthCheck(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "server_up\n")
}

func main() {
	http.HandleFunc("/health", healthCheck)

	http.ListenAndServe(":8080", nil)
}
