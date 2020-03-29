package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/aschereT/ea-gaming-review/db"
	"github.com/gorilla/mux"
	"github.com/hashicorp/go-memdb"
)

type Response struct {
	Data  interface{} `json:"Data,omitempty"`
	Error error       `json:"Error,omitempty"`
}

type CreateBlogPostResponse struct {
	ID string `json:"ID,omitempty"`
}

type GetBlogPostResponse struct {
	IDs []string `json:"IDs,omitempty"`
}

var (
	inMemDB *memdb.MemDB
)

func healthCheck(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "server_up\n")
}

func getBlogPosts(w http.ResponseWriter, req *http.Request) {
	const funcname = "getBlogPosts"
	ids, err := db.GetBlogIDs(inMemDB)
	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		fmt.Println(fmt.Errorf("%s : Error getting blog IDs: %w", funcname, err))
		w.WriteHeader(500)
		resp, jsonErr := json.Marshal(Response{Data: nil, Error: err})
		if jsonErr != nil {
			fmt.Println(fmt.Errorf("%s : Error marshalling error response: %w", funcname, jsonErr))
		} else {
			w.Write(resp)
		}
		return
	}

	fmt.Println(funcname, ": Got blog IDs", len(ids))
	w.WriteHeader(200)
	resp, err := json.Marshal(Response{Data: GetBlogPostResponse{IDs: ids}})
	if err != nil {
		fmt.Println(fmt.Errorf("%s : Error marshalling error response: %w", funcname, err))
	} else {
		w.Write(resp)
	}
	return
}

func createBlogPost(w http.ResponseWriter, req *http.Request) {
	const funcname = "createBlogPost"
	id, err := db.CreateBlogPost(inMemDB, db.BlogPost{ArticleText: "testtext", AuthorName: "testname", Title: "testtitle"})
	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		fmt.Println(fmt.Errorf("%s : Error creating new blog post: %w", funcname, err))
		w.WriteHeader(500)
		resp, jsonErr := json.Marshal(Response{Data: nil, Error: err})
		if jsonErr != nil {
			fmt.Println(fmt.Errorf("%s : Error marshalling error response: %w", funcname, jsonErr))
		} else {
			w.Write(resp)
		}
		return
	}

	fmt.Println(funcname, ": Created new blog post", id)
	w.WriteHeader(200)
	resp, err := json.Marshal(Response{Data: CreateBlogPostResponse{ID: id}})
	if err != nil {
		fmt.Println(fmt.Errorf("%s : Error marshalling error response: %w", funcname, err))
	} else {
		w.Write(resp)
	}
	return
}

func deleteBlogPost(w http.ResponseWriter, req *http.Request) {

}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/health", healthCheck).Methods(http.MethodGet)

	r.HandleFunc("/blog", getBlogPosts).Methods(http.MethodGet)
	r.HandleFunc("/blog", createBlogPost).Methods(http.MethodPost)
	r.HandleFunc("/blog", deleteBlogPost).Methods(http.MethodDelete)

	http.Handle("/", r)

	newDB, err := db.CreateDB()
	if err != nil {
		panic(err)
	}
	inMemDB = newDB

	fmt.Println("main: server up, listening at :8080")
	http.ListenAndServe(":8080", nil)

}
