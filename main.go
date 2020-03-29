package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

func healthCheck(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "server_up\n")
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/health", healthCheck)

	http.Handle("/", r)

	http.ListenAndServe(":8080", nil)
}
