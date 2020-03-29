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
	Error string      `json:"Error,omitempty"`
}

type CreateBlogPostResponse struct {
	ID string `json:"ID"`
}

type GetBlogPostIDsResponse struct {
	IDs []string `json:"IDs"`
}

var (
	inMemDB *memdb.MemDB
)

func healthCheckHandler(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "server_up")
}

func respondWithError(w http.ResponseWriter, statusCode int, err error) {
	const funcname = "respondWithError"
	w.Header().Set("Content-Type", "application/json")

	fmt.Println(err)
	w.WriteHeader(statusCode)
	resp, jsonErr := json.Marshal(Response{Data: nil, Error: fmt.Sprint(err)})
	if jsonErr != nil {
		fmt.Println(fmt.Errorf("%s : Error marshalling error response: %w", funcname, jsonErr))
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(nil)
	} else {
		w.Write(resp)
	}
}

func getBlogPostsIDsHandler(w http.ResponseWriter, req *http.Request) {
	const funcname = "getBlogPosts"
	w.Header().Set("Content-Type", "application/json")

	ids, err := db.GetBlogIDs(inMemDB)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, fmt.Errorf("%s : Error getting blog IDs: %w", funcname, err))
		return
	}

	fmt.Println(funcname, ": Got blog IDs", len(ids))
	w.WriteHeader(http.StatusOK)
	resp, err := json.Marshal(Response{Data: GetBlogPostIDsResponse{IDs: ids}})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, fmt.Errorf("%s : Error marshalling error response: %w", funcname, err))
	} else {
		w.Write(resp)
	}
}

func getSingleBlogPostHandler(w http.ResponseWriter, req *http.Request) {
	const funcname = "getABlogPost"
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(req)
	id := vars["id"]
	post, err := db.GetBlogPost(inMemDB, id)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err)
		return
	}
	if post == nil {
		respondWithError(w, http.StatusNotFound, fmt.Errorf("%s : No such post found, %s", funcname, id))
		return
	}

	fmt.Println(funcname, ": Got blog post", id)
	w.WriteHeader(http.StatusOK)
	resp, err := json.Marshal(Response{Data: *post})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, fmt.Errorf("%s : Error marshalling error response: %w", funcname, err))
	} else {
		w.Write(resp)
	}
}

func createBlogPostHandler(w http.ResponseWriter, req *http.Request) {
	const funcname = "createBlogPost"
	w.Header().Set("Content-Type", "application/json")

	defer req.Body.Close()
	dec := json.NewDecoder(req.Body)
	dec.DisallowUnknownFields()
	var newPost db.BlogPost
	err := dec.Decode(&newPost)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, fmt.Errorf("%s : Error decoding body: %w", funcname, err))
		return
	}

	id, err := db.CreateBlogPost(inMemDB, newPost)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, fmt.Errorf("%s : Error creating new blog post: %w", funcname, err))
		return
	}

	fmt.Println(funcname, ": Created new blog post", id)
	w.WriteHeader(http.StatusOK)
	resp, err := json.Marshal(Response{Data: CreateBlogPostResponse{ID: id}})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, fmt.Errorf("%s : Error marshalling error response: %w", funcname, err))
	} else {
		w.Write(resp)
	}
}

func deleteBlogPostHandler(w http.ResponseWriter, req *http.Request) {

}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/health", healthCheckHandler).Methods(http.MethodGet)

	r.HandleFunc("/blog", getBlogPostsIDsHandler).Methods(http.MethodGet)
	r.HandleFunc("/blog", createBlogPostHandler).Methods(http.MethodPost)
	
	r.HandleFunc("/blog/{id}", getSingleBlogPostHandler).Methods(http.MethodGet)
	r.HandleFunc("/blog/{id}", deleteBlogPostHandler).Methods(http.MethodDelete)

	http.Handle("/", r)

	newDB, err := db.CreateDB()
	if err != nil {
		panic(err)
	}
	inMemDB = newDB

	fmt.Println("main: server up, listening at :8080")
	http.ListenAndServe(":8080", nil)

}
