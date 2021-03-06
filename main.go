package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/aschereT/ea-gaming-review/db"
	"github.com/gorilla/mux"
	"github.com/hashicorp/go-memdb"
)

type Response struct {
	Data  interface{} `json:"Data,omitempty"`
	Error string      `json:"Error,omitempty"`
}

type CreateBlogPostOrCommentResponse struct {
	ID string `json:"ID"`
}

type GetBlogPostIDsResponse struct {
	IDs []string `json:"IDs"`
}

type GetBlogCommentsIDsResponse struct {
	BlogPostID string   `json:"BlogPostID"`
	IDs        []string `json:"IDs"`
}

var (
	inMemDB *memdb.MemDB
)

func healthCheckHandler(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "server_up")
}

func logError(funcname string, err error) {
	fmt.Printf("[%s] %s: %s\n", time.Now().String(), funcname, err)
}

func log(funcname string, a ...interface{}) {
	fmt.Println(fmt.Sprintf("[%s] %s:", time.Now().String(), funcname), a)
}

func setupDB() *memdb.MemDB {
	newDB, err := db.CreateDB()
	if err != nil {
		panic(err)
	}
	return newDB
}

//immediately respond with Data nil: and Error: err
func respondWithError(w http.ResponseWriter, statusCode int, err error) {
	const funcname = "respondWithError"
	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(statusCode)
	resp, jsonErr := json.Marshal(Response{Data: nil, Error: fmt.Sprint(err)})
	if jsonErr != nil {
		logError(funcname, fmt.Errorf("Error marshalling error response: %w", jsonErr))
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(nil)
	} else {
		w.Write(resp)
	}
}

func getBlogPostsIDsHandler(w http.ResponseWriter, req *http.Request) {
	const funcname = "getBlogPostsIDsHandler"
	w.Header().Set("Content-Type", "application/json")

	ids, err := db.GetBlogIDs(inMemDB)
	if err != nil {
		logError(funcname, err)
		respondWithError(w, http.StatusInternalServerError, fmt.Errorf("Error getting blog IDs"))
		return
	}

	log(funcname, "Got", len(ids), "blog IDs")
	w.WriteHeader(http.StatusOK)
	resp, err := json.Marshal(Response{Data: GetBlogPostIDsResponse{IDs: ids}})
	if err != nil {
		logError(funcname, err)
		respondWithError(w, http.StatusInternalServerError, fmt.Errorf("Error marshalling error response"))
	} else {
		w.Write(resp)
	}
}

func getSingleBlogPostHandler(w http.ResponseWriter, req *http.Request) {
	const funcname = "getSingleBlogPostHandler"
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(req)
	id := vars["id"]
	post, err := db.GetBlogPost(inMemDB, id)
	if err != nil {
		logError(funcname, err)
		respondWithError(w, http.StatusInternalServerError, fmt.Errorf("Error getting blog post"))
		return
	}
	if post == nil {
		err = fmt.Errorf("No post found with ID %s", id)
		logError(funcname, err)
		respondWithError(w, http.StatusNotFound, err)
		return
	}

	log(funcname, "Got blog post", id)
	w.WriteHeader(http.StatusOK)
	resp, err := json.Marshal(Response{Data: *post})
	if err != nil {
		logError(funcname, err)
		respondWithError(w, http.StatusInternalServerError, fmt.Errorf("Error marshalling response"))
	} else {
		w.Write(resp)
	}
}

func createBlogPostHandler(w http.ResponseWriter, req *http.Request) {
	const funcname = "createBlogPostHandler"
	w.Header().Set("Content-Type", "application/json")

	defer req.Body.Close()
	dec := json.NewDecoder(req.Body)
	dec.DisallowUnknownFields()
	var newPost db.BlogPost
	err := dec.Decode(&newPost)
	if err != nil {
		logError(funcname, err)
		respondWithError(w, http.StatusInternalServerError, fmt.Errorf("Error decoding request body"))
		return
	}

	log(funcname, "Received request to create new blog post", fmt.Sprintf("%#v", newPost))
	if newPost.Title == "" {
		err := fmt.Errorf("Title should not be empty")
		logError(funcname, err)
		respondWithError(w, http.StatusBadRequest, err)
		return
	}
	if newPost.ArticleText == "" {
		err := fmt.Errorf("ArticleText should not be empty")
		logError(funcname, err)
		respondWithError(w, http.StatusBadRequest, err)
		return
	}
	if newPost.AuthorName == "" {
		err := fmt.Errorf("AuthorName should not be empty")
		logError(funcname, err)
		respondWithError(w, http.StatusBadRequest, err)
		return
	}
	if newPost.ID != "" {
		//should be empty
		err := fmt.Errorf("ID should not be defined in new post requests")
		logError(funcname, err)
		respondWithError(w, http.StatusBadRequest, err)
		return
	}
	log(funcname, "Request looks legit", fmt.Sprintf("%#v", newPost))

	id, err := db.CreateBlogPost(inMemDB, newPost)
	if err != nil {
		logError(funcname, err)
		respondWithError(w, http.StatusInternalServerError, fmt.Errorf("Error creating new blog post"))
		return
	}

	log(funcname, "Created new blog post", id)
	w.WriteHeader(http.StatusOK)
	resp, err := json.Marshal(Response{Data: CreateBlogPostOrCommentResponse{ID: id}})
	if err != nil {
		logError(funcname, err)
		respondWithError(w, http.StatusInternalServerError, fmt.Errorf("Error marshalling response"))
	} else {
		w.Write(resp)
	}
}

func deleteBlogPostHandler(w http.ResponseWriter, req *http.Request) {
	const funcname = "deleteBlogPostHandler"
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(req)
	id := vars["id"]
	exists, err := db.DeleteBlogPost(inMemDB, id)
	if err != nil {
		logError(funcname, err)
		respondWithError(w, http.StatusInternalServerError, fmt.Errorf("Error deleting blog post"))
		return
	}
	if !exists {
		err = fmt.Errorf("No post found with ID %s", id)
		logError(funcname, err)
		respondWithError(w, http.StatusNotFound, err)
		return
	}

	log(funcname, "Deleted blog post", id)
	w.WriteHeader(http.StatusOK)
	resp, err := json.Marshal(Response{Data: "OK"})
	if err != nil {
		logError(funcname, err)
		respondWithError(w, http.StatusInternalServerError, fmt.Errorf("Error marshalling response"))
	} else {
		w.Write(resp)
	}
}

func getBlogCommentsIDsHandler(w http.ResponseWriter, req *http.Request) {
	const funcname = "getBlogCommentsIDsHandler"
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(req)
	id := vars["id"]
	ids, err := db.GetCommentIDs(inMemDB, id)
	if err != nil {
		logError(funcname, err)
		respondWithError(w, http.StatusInternalServerError, fmt.Errorf("Error getting comment IDs"))
		return
	}

	log(funcname, "Got", len(ids), "comment IDs")
	w.WriteHeader(http.StatusOK)
	resp, err := json.Marshal(Response{Data: GetBlogCommentsIDsResponse{IDs: ids, BlogPostID: id}})
	if err != nil {
		logError(funcname, err)
		respondWithError(w, http.StatusInternalServerError, fmt.Errorf("Error marshalling error response"))
	} else {
		w.Write(resp)
	}
}

func getSingleBlogCommentHandler(w http.ResponseWriter, req *http.Request) {
	const funcname = "getSingleBlogCommentHandler"
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(req)
	id := vars["id"]
	commentID := vars["commentID"]
	comment, err := db.GetBlogComment(inMemDB, id, commentID)
	if err != nil {
		logError(funcname, err)
		respondWithError(w, http.StatusInternalServerError, fmt.Errorf("Error getting blog post"))
		return
	}
	if comment == nil {
		err = fmt.Errorf("No post found with ID %s", id)
		logError(funcname, err)
		respondWithError(w, http.StatusNotFound, err)
		return
	}

	log(funcname, "Got blog post", id)
	w.WriteHeader(http.StatusOK)
	resp, err := json.Marshal(Response{Data: *comment})
	if err != nil {
		logError(funcname, err)
		respondWithError(w, http.StatusInternalServerError, fmt.Errorf("Error marshalling response"))
	} else {
		w.Write(resp)
	}
}

func createBlogCommentHandler(w http.ResponseWriter, req *http.Request) {
	const funcname = "createBlogCommentHandler"
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(req)
	articleID := vars["id"]

	defer req.Body.Close()
	dec := json.NewDecoder(req.Body)
	dec.DisallowUnknownFields()
	var newPost db.BlogComment
	err := dec.Decode(&newPost)
	if err != nil {
		logError(funcname, err)
		respondWithError(w, http.StatusInternalServerError, fmt.Errorf("Error decoding request body"))
		return
	}

	log(funcname, "Received request to create new blog post", fmt.Sprintf("%#v", newPost))
	if newPost.CommentText == "" {
		err := fmt.Errorf("CommentText should not be empty")
		logError(funcname, err)
		respondWithError(w, http.StatusBadRequest, err)
		return
	}
	if newPost.AuthorName == "" {
		err := fmt.Errorf("AuthorName should not be empty")
		logError(funcname, err)
		respondWithError(w, http.StatusBadRequest, err)
		return
	}
	if newPost.ArticleID != "" {
		err := fmt.Errorf("ArticleID should not be defined in new post requests")
		logError(funcname, err)
		respondWithError(w, http.StatusBadRequest, err)
		return
	}
	if newPost.ID != "" {
		//should be empty
		err := fmt.Errorf("ID should not be defined in new post requests")
		logError(funcname, err)
		respondWithError(w, http.StatusBadRequest, err)
		return
	}
	newPost.ArticleID = articleID
	log(funcname, "Request looks legit", fmt.Sprintf("%#v", newPost))

	commentID, err := db.CreateBlogComment(inMemDB, newPost)
	if err != nil {
		logError(funcname, err)
		respondWithError(w, http.StatusInternalServerError, fmt.Errorf("Error creating new blog post"))
		return
	}

	log(funcname, "Created new comment on", articleID, commentID)
	w.WriteHeader(http.StatusOK)
	resp, err := json.Marshal(Response{Data: CreateBlogPostOrCommentResponse{ID: commentID}})
	if err != nil {
		logError(funcname, err)
		respondWithError(w, http.StatusInternalServerError, fmt.Errorf("Error marshalling response"))
	} else {
		w.Write(resp)
	}
}

func deleteBlogCommentHandler(w http.ResponseWriter, req *http.Request) {
	const funcname = "deleteBlogCommentHandler"
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(req)
	id := vars["id"]
	commentID := vars["commentID"]
	exists, err := db.DeleteBlogComment(inMemDB, id, commentID)
	if err != nil {
		logError(funcname, err)
		respondWithError(w, http.StatusInternalServerError, fmt.Errorf("Error deleting blog post"))
		return
	}
	if !exists {
		err = fmt.Errorf("No post found with ID %s", id)
		logError(funcname, err)
		respondWithError(w, http.StatusNotFound, err)
		return
	}

	log(funcname, "Deleted blog post", id)
	w.WriteHeader(http.StatusOK)
	resp, err := json.Marshal(Response{Data: "OK"})
	if err != nil {
		logError(funcname, err)
		respondWithError(w, http.StatusInternalServerError, fmt.Errorf("Error marshalling response"))
	} else {
		w.Write(resp)
	}
}

func main() {
	const funcname = "main"
	r := mux.NewRouter()
	r.HandleFunc("/health", healthCheckHandler).Methods(http.MethodGet)

	r.HandleFunc("/blog", getBlogPostsIDsHandler).Methods(http.MethodGet)
	r.HandleFunc("/blog", createBlogPostHandler).Methods(http.MethodPost)

	r.HandleFunc("/blog/{id}", getSingleBlogPostHandler).Methods(http.MethodGet)
	r.HandleFunc("/blog/{id}", deleteBlogPostHandler).Methods(http.MethodDelete)

	r.HandleFunc("/blog/{id}/comment", getBlogCommentsIDsHandler).Methods(http.MethodGet)
	r.HandleFunc("/blog/{id}/comment", createBlogCommentHandler).Methods(http.MethodPost)

	r.HandleFunc("/blog/{id}/comment/{commentID}", getSingleBlogCommentHandler).Methods(http.MethodGet)
	r.HandleFunc("/blog/{id}/comment/{commentID}", deleteBlogCommentHandler).Methods(http.MethodDelete)
	http.Handle("/", r)

	inMemDB = setupDB()

	const DefaultAddr = ":8080"
	log(funcname, "server up, listening at :8080")
	err := http.ListenAndServe(DefaultAddr, nil)
	if err != nil {
		panic(err)
	}
}
