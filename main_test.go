package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/aschereT/ea-gaming-review/db"
	"github.com/gorilla/mux"
)

type expectedResponseCreateBlogPostOrComment struct {
	Data struct {
		ID string `json:"ID"`
	} `json:"Data"`
	Error string `json:"Error"`
}

type expectedResponseIDs struct {
	Data struct {
		IDs        []string `json:"IDs"`
		BlogPostID string   `json:"BlogPostID"`
	} `json:"Data"`
	Error string `json:"Error"`
}

type expectedResponseGetPost struct {
	Data struct {
		ID          string `json:"ID"`
		Title       string `json:"Title"`
		ArticleText string `json:"ArticleText"`
		AuthorName  string `json:"AuthorName"`
	} `json:"Data"`
}

type expectedResponseGetComment struct {
	Data struct {
		ID          string `json:"ID"`
		ArticleID   string `json:"ArticleID"`
		CommentText string `json:"CommentText"`
		AuthorName  string `json:"AuthorName"`
	} `json:"Data"`
}

func Test_HealthCheck(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "/health", nil)
	if err != nil {
		t.Error(err)
	}

	rec := httptest.NewRecorder()
	handler := http.HandlerFunc(healthCheckHandler)

	handler.ServeHTTP(rec, req)

	actual := rec.Result()

	if actual.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, actual.StatusCode)
	}

	expectedBody := "server_up"
	actualBody := rec.Body.String()

	if actualBody != expectedBody {
		t.Errorf("Expected body to be %s, got %s", expectedBody, actualBody)
	}
}

func Test_RespondWithError(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "/error", nil)
	if err != nil {
		t.Error(err)
	}

	expectedStatusCode := http.StatusTeapot
	expectedError := fmt.Errorf("error detected, self-terminating")
	respondWithErrorHandler := func(w http.ResponseWriter, req *http.Request) {
		respondWithError(w, expectedStatusCode, expectedError)
	}

	rec := httptest.NewRecorder()
	handler := http.HandlerFunc(respondWithErrorHandler)

	handler.ServeHTTP(rec, req)

	actual := rec.Result()

	if actual.StatusCode != expectedStatusCode {
		t.Errorf("Expected status code %d, got %d", expectedStatusCode, actual.StatusCode)
	}

	if contentType := actual.Header.Get("Content-Type"); contentType != "application/json" {
		t.Errorf("Expected Content-Type to be application/json, got %s", contentType)
	}

	expectedBody := "{\"Error\":\"error detected, self-terminating\"}"
	actualBody := rec.Body.String()

	if actualBody != expectedBody {
		t.Errorf("Expected body to be %s, got %s", expectedBody, actualBody)
	}
}

func Test_CreateBlogPost(t *testing.T) {
	//set up in-mem db, and tear down after
	inMemDB = setupDB()
	defer func() {
		inMemDB = nil
	}()

	req, err := http.NewRequest(http.MethodPost, "/blog", strings.NewReader("{\"Title\":\"I've come to make an announcement\",\"ArticleText\":\"walnut moon\",\"AuthorName\":\"Dr. Eggman\"}"))
	if err != nil {
		t.Error(err)
	}

	rec := httptest.NewRecorder()
	handler := http.HandlerFunc(createBlogPostHandler)

	handler.ServeHTTP(rec, req)

	actual := rec.Result()

	if actual.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, actual.StatusCode)
	}

	var actualResponse expectedResponseCreateBlogPostOrComment
	err = json.Unmarshal(rec.Body.Bytes(), &actualResponse)
	if err != nil {
		t.Error(err)
	}

	id := actualResponse.Data.ID

	txn := inMemDB.Txn(false)
	defer txn.Abort()

	actualPost, err := txn.First(db.BlogPostTable, "id", id)
	if err != nil {
		t.Error(err)
	}
	if actualPost == nil {
		t.Error("Expected post to make it into the DB, got nil")
	}
	post := actualPost.(db.BlogPost)

	if post.AuthorName != "Dr. Eggman" {
		t.Errorf("Expected author name to be Dr. Eggman, got %s", post.AuthorName)
	}

	if post.Title != "I've come to make an announcement" {
		t.Errorf("Expected title to be I've come to make an announcement, got %s", post.Title)
	}

	if post.ArticleText != "walnut moon" {
		t.Errorf("Expected article text to be walnut moon, got %s", post.ArticleText)
	}

	if post.ID != id {
		t.Errorf("Expected ID to be %s, got %s", id, post.ID)
	}
}

func Test_CreateBlogPost_MissingTitle(t *testing.T) {
	//set up in-mem db, and tear down after
	inMemDB = setupDB()
	defer func() {
		inMemDB = nil
	}()

	req, err := http.NewRequest(http.MethodPost, "/blog", strings.NewReader("{\"ArticleText\":\"walnut moon\",\"AuthorName\":\"Dr. Eggman\"}"))
	if err != nil {
		t.Error(err)
	}

	rec := httptest.NewRecorder()
	handler := http.HandlerFunc(createBlogPostHandler)

	handler.ServeHTTP(rec, req)

	actual := rec.Result()

	if actual.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, actual.StatusCode)
	}

	type expectedResponseType struct {
		Data struct {
			ID string `json:"ID"`
		} `json:"Data"`
		Error string `json:"Error"`
	}
	var actualResponse expectedResponseType
	err = json.Unmarshal(rec.Body.Bytes(), &actualResponse)
	if err != nil {
		t.Error(err)
	}

	expectedError := "Title should not be empty"
	if actualResponse.Error != expectedError {
		t.Errorf("Expected error to be %s, got %s", expectedError, actualResponse.Error)
	}
}

func Test_CreateBlogPost_WithID(t *testing.T) {
	//set up in-mem db, and tear down after
	inMemDB = setupDB()
	defer func() {
		inMemDB = nil
	}()

	req, err := http.NewRequest(http.MethodPost, "/blog", strings.NewReader("{\"Title\":\"I've come to make an announcement\",\"ArticleText\":\"walnut moon\",\"AuthorName\":\"Dr. Eggman\",\"ID\":\"idshouldntbehere\"}"))
	if err != nil {
		t.Error(err)
	}

	rec := httptest.NewRecorder()
	handler := http.HandlerFunc(createBlogPostHandler)

	handler.ServeHTTP(rec, req)

	actual := rec.Result()

	if actual.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, actual.StatusCode)
	}

	type expectedResponseType struct {
		Data struct {
			ID string `json:"ID"`
		} `json:"Data"`
		Error string `json:"Error"`
	}
	var actualResponse expectedResponseType
	err = json.Unmarshal(rec.Body.Bytes(), &actualResponse)
	if err != nil {
		t.Error(err)
	}

	expectedError := "ID should not be defined in new post requests"
	if actualResponse.Error != expectedError {
		t.Errorf("Expected error to be %s, got %s", expectedError, actualResponse.Error)
	}
}

func Test_CreateBlogPost_MissingArticleText(t *testing.T) {
	//set up in-mem db, and tear down after
	inMemDB = setupDB()
	defer func() {
		inMemDB = nil
	}()

	req, err := http.NewRequest(http.MethodPost, "/blog", strings.NewReader("{\"Title\":\"I've come to make an announcement\",\"AuthorName\":\"Dr. Eggman\"}"))
	if err != nil {
		t.Error(err)
	}

	rec := httptest.NewRecorder()
	handler := http.HandlerFunc(createBlogPostHandler)

	handler.ServeHTTP(rec, req)

	actual := rec.Result()

	if actual.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, actual.StatusCode)
	}

	type expectedResponseType struct {
		Data struct {
			ID string `json:"ID"`
		} `json:"Data"`
		Error string `json:"Error"`
	}
	var actualResponse expectedResponseType
	err = json.Unmarshal(rec.Body.Bytes(), &actualResponse)
	if err != nil {
		t.Error(err)
	}

	expectedError := "ArticleText should not be empty"
	if actualResponse.Error != expectedError {
		t.Errorf("Expected error to be %s, got %s", expectedError, actualResponse.Error)
	}
}

func Test_CreateBlogPost_MissingAuthorName(t *testing.T) {
	//set up in-mem db, and tear down after
	inMemDB = setupDB()
	defer func() {
		inMemDB = nil
	}()

	req, err := http.NewRequest(http.MethodPost, "/blog", strings.NewReader("{\"Title\":\"I've come to make an announcement\",\"ArticleText\":\"walnut moon\"}"))
	if err != nil {
		t.Error(err)
	}

	rec := httptest.NewRecorder()
	handler := http.HandlerFunc(createBlogPostHandler)

	handler.ServeHTTP(rec, req)

	actual := rec.Result()

	if actual.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, actual.StatusCode)
	}

	type expectedResponseType struct {
		Data struct {
			ID string `json:"ID"`
		} `json:"Data"`
		Error string `json:"Error"`
	}
	var actualResponse expectedResponseType
	err = json.Unmarshal(rec.Body.Bytes(), &actualResponse)
	if err != nil {
		t.Error(err)
	}

	expectedError := "AuthorName should not be empty"
	if actualResponse.Error != expectedError {
		t.Errorf("Expected error to be %s, got %s", expectedError, actualResponse.Error)
	}
}

func Test_GetBlogPostIDs_Empty(t *testing.T) {
	inMemDB = setupDB()
	defer func() {
		inMemDB = nil
	}()

	req, err := http.NewRequest(http.MethodGet, "/blog", nil)
	if err != nil {
		t.Error(err)
	}

	rec := httptest.NewRecorder()
	handler := http.HandlerFunc(getBlogPostsIDsHandler)

	handler.ServeHTTP(rec, req)

	actual := rec.Result()

	if actual.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, actual.StatusCode)
	}

	expectedBody := "{\"Data\":{\"IDs\":null}}"
	actualBody := rec.Body.String()

	if actualBody != expectedBody {
		t.Errorf("Expected body to be %s, got %s", expectedBody, actualBody)
	}
}

func Test_GetSingleBlogPost(t *testing.T) {
	//set up in-mem db, and tear down after
	inMemDB = setupDB()
	defer func() {
		inMemDB = nil
	}()

	//add a post
	req, err := http.NewRequest(http.MethodPost, "/blog", strings.NewReader("{\"Title\":\"I've come to make an announcement\",\"ArticleText\":\"walnut moon\",\"AuthorName\":\"Dr. Eggman\"}"))
	if err != nil {
		t.Error(err)
	}

	rec := httptest.NewRecorder()
	handler := http.HandlerFunc(createBlogPostHandler)
	handler.ServeHTTP(rec, req)
	actual := rec.Result()

	if actual.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, actual.StatusCode)
	}

	var actualResponse expectedResponseCreateBlogPostOrComment
	err = json.Unmarshal(rec.Body.Bytes(), &actualResponse)
	if err != nil {
		t.Error(err)
	}

	id := actualResponse.Data.ID

	//try to get it
	r := mux.NewRouter()
	r.HandleFunc("/blog/{id}", getSingleBlogPostHandler)

	ts := httptest.NewServer(r)
	defer ts.Close()

	req, err = http.NewRequest(http.MethodGet, ts.URL+"/blog/"+id, nil)
	if err != nil {
		t.Error(err)
	}

	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	actual = rec.Result()
	returnedBody := rec.Body.String()

	expectedBody := "{\"Data\":{\"ID\":\"" + id + "\",\"Title\":\"I've come to make an announcement\",\"ArticleText\":\"walnut moon\",\"AuthorName\":\"Dr. Eggman\"}}"

	if returnedBody != expectedBody {
		t.Errorf("Expected actual body to match expected body, but differs: \nexpected: %s\nactual:   %s", expectedBody, returnedBody)
	}
}

func Test_DeleteBlogPost(t *testing.T) {
	//set up in-mem db, and tear down after
	inMemDB = setupDB()
	defer func() {
		inMemDB = nil
	}()

	//add a post
	req, err := http.NewRequest(http.MethodPost, "/blog", strings.NewReader("{\"Title\":\"I've come to make an announcement\",\"ArticleText\":\"walnut moon\",\"AuthorName\":\"Dr. Eggman\"}"))
	if err != nil {
		t.Error(err)
	}

	rec := httptest.NewRecorder()
	handler := http.HandlerFunc(createBlogPostHandler)
	handler.ServeHTTP(rec, req)
	actual := rec.Result()

	if actual.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, actual.StatusCode)
	}

	var actualResponse expectedResponseCreateBlogPostOrComment
	err = json.Unmarshal(rec.Body.Bytes(), &actualResponse)
	if err != nil {
		t.Error(err)
	}

	id := actualResponse.Data.ID

	//try to delete it
	r := mux.NewRouter()
	r.HandleFunc("/blog/{id}", deleteBlogPostHandler)

	ts := httptest.NewServer(r)
	defer ts.Close()

	req, err = http.NewRequest(http.MethodDelete, ts.URL+"/blog/"+id, nil)
	if err != nil {
		t.Error(err)
	}

	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	actual = rec.Result()
	returnedBody := rec.Body.String()

	expectedBody := "{\"Data\":\"OK\"}"

	if returnedBody != expectedBody {
		t.Errorf("Expected actual body to match expected body, but differs: \nexpected: %s\nactual:   %s", expectedBody, returnedBody)
	}

	txn := inMemDB.Txn(false)
	defer txn.Abort()

	post, err := txn.First(db.BlogPostTable, "id", id)
	if err != nil {
		t.Error(err)
	}
	if post != nil {
		coercedPost := post.(db.BlogPost)
		t.Errorf("Expected post to be deleted, got %#v", coercedPost)
	}
}

func Test_DeleteBlogPost_DoesntExist(t *testing.T) {
	//set up in-mem db, and tear down after
	inMemDB = setupDB()
	defer func() {
		inMemDB = nil
	}()

	id := "McDoesntExist"

	//try to delete it
	r := mux.NewRouter()
	r.HandleFunc("/blog/{id}", deleteBlogPostHandler)

	ts := httptest.NewServer(r)
	defer ts.Close()

	req, err := http.NewRequest(http.MethodDelete, ts.URL+"/blog/"+id, nil)
	if err != nil {
		t.Error(err)
	}

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	returnedBody := rec.Body.String()
	expectedBody := "{\"Error\":\"No post found with ID " + id + "\"}"

	if returnedBody != expectedBody {
		t.Errorf("Expected actual body to match expected body, but differs: \nexpected: %s\nactual:   %s", expectedBody, returnedBody)
	}

	txn := inMemDB.Txn(false)
	defer txn.Abort()

	post, err := txn.First(db.BlogPostTable, "id", id)
	if err != nil {
		t.Error(err)
	}
	if post != nil {
		coercedPost := post.(db.BlogPost)
		t.Errorf("Expected post to be deleted, got %#v", coercedPost)
	}
}

func Test_CreateBlogComment(t *testing.T) {
	//set up in-mem db, and tear down after
	inMemDB = setupDB()
	defer func() {
		inMemDB = nil
	}()

	req, err := http.NewRequest(http.MethodPost, "/blog", strings.NewReader("{\"Title\":\"I've come to make an announcement\",\"ArticleText\":\"walnut moon\",\"AuthorName\":\"Dr. Eggman\"}"))
	if err != nil {
		t.Error(err)
	}

	rec := httptest.NewRecorder()
	handler := http.HandlerFunc(createBlogPostHandler)

	handler.ServeHTTP(rec, req)

	actual := rec.Result()

	if actual.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, actual.StatusCode)
	}

	var createBlogResponse expectedResponseCreateBlogPostOrComment
	err = json.Unmarshal(rec.Body.Bytes(), &createBlogResponse)
	if err != nil {
		t.Error(err)
	}

	id := createBlogResponse.Data.ID

	//try to comment on the post
	r := mux.NewRouter()
	r.HandleFunc("/blog/{id}/comment", createBlogCommentHandler)

	ts := httptest.NewServer(r)
	defer ts.Close()

	rec = httptest.NewRecorder()
	req, err = http.NewRequest(http.MethodPost, ts.URL+"/blog/"+id+"/comment", strings.NewReader("{\"AuthorName\": \"Anony Mouse\",\"CommentText\": \"this review sucks\"}"))
	if err != nil {
		t.Error(err)
	}

	r.ServeHTTP(rec, req)

	actual = rec.Result()

	if actual.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, actual.StatusCode)
	}

	var actualResponse expectedResponseCreateBlogPostOrComment
	err = json.Unmarshal(rec.Body.Bytes(), &actualResponse)
	if err != nil {
		t.Error(err)
	}

	commentID := actualResponse.Data.ID

	//check if the comment made it into the DB
	txn := inMemDB.Txn(false)
	defer txn.Abort()

	post, err := txn.First(db.CommentsTable, "id", commentID)
	if err != nil {
		t.Error(err)
	}
	if post == nil {
		t.Error("Expected comment to be made, got nil")
	}
}

func Test_CreateBlogComment_MissingCommentText(t *testing.T) {
	//set up in-mem db, and tear down after
	inMemDB = setupDB()
	defer func() {
		inMemDB = nil
	}()

	req, err := http.NewRequest(http.MethodPost, "/blog", strings.NewReader("{\"Title\":\"I've come to make an announcement\",\"ArticleText\":\"walnut moon\",\"AuthorName\":\"Dr. Eggman\"}"))
	if err != nil {
		t.Error(err)
	}

	rec := httptest.NewRecorder()
	handler := http.HandlerFunc(createBlogPostHandler)

	handler.ServeHTTP(rec, req)

	actual := rec.Result()

	if actual.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, actual.StatusCode)
	}

	var createBlogResponse expectedResponseCreateBlogPostOrComment
	err = json.Unmarshal(rec.Body.Bytes(), &createBlogResponse)
	if err != nil {
		t.Error(err)
	}

	id := createBlogResponse.Data.ID

	//try to comment on the post
	r := mux.NewRouter()
	r.HandleFunc("/blog/{id}/comment", createBlogCommentHandler)

	ts := httptest.NewServer(r)
	defer ts.Close()

	rec = httptest.NewRecorder()
	req, err = http.NewRequest(http.MethodPost, ts.URL+"/blog/"+id+"/comment", strings.NewReader("{\"AuthorName\": \"Anony Mouse\"}"))
	if err != nil {
		t.Error(err)
	}

	r.ServeHTTP(rec, req)

	actual = rec.Result()

	if actual.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, actual.StatusCode)
	}

	var actualResponse expectedResponseCreateBlogPostOrComment
	err = json.Unmarshal(rec.Body.Bytes(), &actualResponse)
	if err != nil {
		t.Error(err)
	}

	expectedError := "CommentText should not be empty"
	if actualResponse.Error != expectedError {
		t.Errorf("Expected error to be %s, got %s", expectedError, actualResponse.Error)
	}
}

func Test_CreateBlogComment_AddedID(t *testing.T) {
	//set up in-mem db, and tear down after
	inMemDB = setupDB()
	defer func() {
		inMemDB = nil
	}()

	req, err := http.NewRequest(http.MethodPost, "/blog", strings.NewReader("{\"Title\":\"I've come to make an announcement\",\"ArticleText\":\"walnut moon\",\"AuthorName\":\"Dr. Eggman\"}"))
	if err != nil {
		t.Error(err)
	}

	rec := httptest.NewRecorder()
	handler := http.HandlerFunc(createBlogPostHandler)

	handler.ServeHTTP(rec, req)

	actual := rec.Result()

	if actual.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, actual.StatusCode)
	}

	var createBlogResponse expectedResponseCreateBlogPostOrComment
	err = json.Unmarshal(rec.Body.Bytes(), &createBlogResponse)
	if err != nil {
		t.Error(err)
	}

	id := createBlogResponse.Data.ID

	//try to comment on the post
	r := mux.NewRouter()
	r.HandleFunc("/blog/{id}/comment", createBlogCommentHandler)

	ts := httptest.NewServer(r)
	defer ts.Close()

	rec = httptest.NewRecorder()
	req, err = http.NewRequest(http.MethodPost, ts.URL+"/blog/"+id+"/comment", strings.NewReader("{\"AuthorName\": \"Anony Mouse\",\"CommentText\": \"this review sucks\",\"ID\":\"qwer\"}"))
	if err != nil {
		t.Error(err)
	}

	r.ServeHTTP(rec, req)

	actual = rec.Result()

	if actual.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, actual.StatusCode)
	}

	var actualResponse expectedResponseCreateBlogPostOrComment
	err = json.Unmarshal(rec.Body.Bytes(), &actualResponse)
	if err != nil {
		t.Error(err)
	}

	expectedError := "ID should not be defined in new post requests"
	if actualResponse.Error != expectedError {
		t.Errorf("Expected error to be %s, got %s", expectedError, actualResponse.Error)
	}
}

func Test_CreateBlogComment_AddedArticleID(t *testing.T) {
	//set up in-mem db, and tear down after
	inMemDB = setupDB()
	defer func() {
		inMemDB = nil
	}()

	req, err := http.NewRequest(http.MethodPost, "/blog", strings.NewReader("{\"Title\":\"I've come to make an announcement\",\"ArticleText\":\"walnut moon\",\"AuthorName\":\"Dr. Eggman\"}"))
	if err != nil {
		t.Error(err)
	}

	rec := httptest.NewRecorder()
	handler := http.HandlerFunc(createBlogPostHandler)

	handler.ServeHTTP(rec, req)

	actual := rec.Result()

	if actual.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, actual.StatusCode)
	}

	var createBlogResponse expectedResponseCreateBlogPostOrComment
	err = json.Unmarshal(rec.Body.Bytes(), &createBlogResponse)
	if err != nil {
		t.Error(err)
	}

	id := createBlogResponse.Data.ID

	//try to comment on the post
	r := mux.NewRouter()
	r.HandleFunc("/blog/{id}/comment", createBlogCommentHandler)

	ts := httptest.NewServer(r)
	defer ts.Close()

	rec = httptest.NewRecorder()
	req, err = http.NewRequest(http.MethodPost, ts.URL+"/blog/"+id+"/comment", strings.NewReader("{\"AuthorName\": \"Anony Mouse\",\"CommentText\": \"this review sucks\",\"ArticleID\":\"qwer\"}"))
	if err != nil {
		t.Error(err)
	}

	r.ServeHTTP(rec, req)

	actual = rec.Result()

	if actual.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, actual.StatusCode)
	}

	var actualResponse expectedResponseCreateBlogPostOrComment
	err = json.Unmarshal(rec.Body.Bytes(), &actualResponse)
	if err != nil {
		t.Error(err)
	}

	expectedError := "ArticleID should not be defined in new post requests"
	if actualResponse.Error != expectedError {
		t.Errorf("Expected error to be %s, got %s", expectedError, actualResponse.Error)
	}
}

func Test_CreateBlogComment_MissingAuthorName(t *testing.T) {
	//set up in-mem db, and tear down after
	inMemDB = setupDB()
	defer func() {
		inMemDB = nil
	}()

	req, err := http.NewRequest(http.MethodPost, "/blog", strings.NewReader("{\"Title\":\"I've come to make an announcement\",\"ArticleText\":\"walnut moon\",\"AuthorName\":\"Dr. Eggman\"}"))
	if err != nil {
		t.Error(err)
	}

	rec := httptest.NewRecorder()
	handler := http.HandlerFunc(createBlogPostHandler)

	handler.ServeHTTP(rec, req)

	actual := rec.Result()

	if actual.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, actual.StatusCode)
	}

	var createBlogResponse expectedResponseCreateBlogPostOrComment
	err = json.Unmarshal(rec.Body.Bytes(), &createBlogResponse)
	if err != nil {
		t.Error(err)
	}

	id := createBlogResponse.Data.ID

	//try to comment on the post
	r := mux.NewRouter()
	r.HandleFunc("/blog/{id}/comment", createBlogCommentHandler)

	ts := httptest.NewServer(r)
	defer ts.Close()

	rec = httptest.NewRecorder()
	req, err = http.NewRequest(http.MethodPost, ts.URL+"/blog/"+id+"/comment", strings.NewReader("{\"CommentText\": \"this review sucks\"}"))
	if err != nil {
		t.Error(err)
	}

	r.ServeHTTP(rec, req)

	actual = rec.Result()

	if actual.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, actual.StatusCode)
	}

	var actualResponse expectedResponseCreateBlogPostOrComment
	err = json.Unmarshal(rec.Body.Bytes(), &actualResponse)
	if err != nil {
		t.Error(err)
	}

	expectedError := "AuthorName should not be empty"
	if actualResponse.Error != expectedError {
		t.Errorf("Expected error to be %s, got %s", expectedError, actualResponse.Error)
	}
}

func Test_GetBlogCommentsIDs_Empty(t *testing.T) {
	//set up in-mem db, and tear down after
	inMemDB = setupDB()
	defer func() {
		inMemDB = nil
	}()

	req, err := http.NewRequest(http.MethodPost, "/blog", strings.NewReader("{\"Title\":\"I've come to make an announcement\",\"ArticleText\":\"walnut moon\",\"AuthorName\":\"Dr. Eggman\"}"))
	if err != nil {
		t.Error(err)
	}

	rec := httptest.NewRecorder()
	handler := http.HandlerFunc(createBlogPostHandler)

	handler.ServeHTTP(rec, req)

	commentResponse := rec.Result()

	if commentResponse.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, commentResponse.StatusCode)
	}

	var createBlogResponse expectedResponseCreateBlogPostOrComment
	err = json.Unmarshal(rec.Body.Bytes(), &createBlogResponse)
	if err != nil {
		t.Error(err)
	}

	id := createBlogResponse.Data.ID

	//try to comment on the post
	r := mux.NewRouter()
	r.HandleFunc("/blog/{id}/comment", createBlogCommentHandler)

	ts := httptest.NewServer(r)
	defer ts.Close()

	rec = httptest.NewRecorder()
	req, err = http.NewRequest(http.MethodPost, ts.URL+"/blog/"+id+"/comment", strings.NewReader("{\"AuthorName\": \"Anony Mouse\",\"CommentText\": \"this review sucks\"}"))
	if err != nil {
		t.Error(err)
	}

	r.ServeHTTP(rec, req)

	commentResponse = rec.Result()

	if commentResponse.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, commentResponse.StatusCode)
	}

	var actualResponse expectedResponseCreateBlogPostOrComment
	err = json.Unmarshal(rec.Body.Bytes(), &actualResponse)
	if err != nil {
		t.Error(err)
	}

	commentID := actualResponse.Data.ID

	getIDsRouter := mux.NewRouter()
	getIDsRouter.HandleFunc("/blog/{id}/comment", getBlogCommentsIDsHandler)

	ts2 := httptest.NewServer(getIDsRouter)
	defer ts2.Close()

	rec = httptest.NewRecorder()
	req, err = http.NewRequest(http.MethodGet, ts2.URL+"/blog/"+id+"/comment", nil)
	if err != nil {
		t.Error(err)
	}

	getIDsRouter.ServeHTTP(rec, req)

	actual := rec.Result()

	if actual.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, actual.StatusCode)
	}

	var getIDsResponse expectedResponseIDs
	err = json.Unmarshal(rec.Body.Bytes(), &getIDsResponse)
	if err != nil {
		t.Error(err)
	}

	if getIDsResponse.Error != "" {
		t.Errorf("Expected no errors, got %s", getIDsResponse.Error)
	}
	if getIDsResponse.Data.BlogPostID != id {
		t.Errorf("Expected comment parent to be %s, got %s", id, getIDsResponse.Data.BlogPostID)
	}
	if len(getIDsResponse.Data.IDs) != 1 {
		t.Errorf("Expected number of comments on blog to be 1, got %d", len(getIDsResponse.Data.IDs))
	}
	if getIDsResponse.Data.IDs[0] != commentID {
		t.Errorf("Expected comment ID to be %s, got %s", commentID, getIDsResponse.Data.IDs[0])
	}
}

func Test_GetBlogComment(t *testing.T) {
	//set up in-mem db, and tear down after
	inMemDB = setupDB()
	defer func() {
		inMemDB = nil
	}()

	req, err := http.NewRequest(http.MethodPost, "/blog", strings.NewReader("{\"Title\":\"I've come to make an announcement\",\"ArticleText\":\"walnut moon\",\"AuthorName\":\"Dr. Eggman\"}"))
	if err != nil {
		t.Error(err)
	}

	rec := httptest.NewRecorder()
	handler := http.HandlerFunc(createBlogPostHandler)

	handler.ServeHTTP(rec, req)

	actual := rec.Result()

	if actual.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, actual.StatusCode)
	}

	var createBlogResponse expectedResponseCreateBlogPostOrComment
	err = json.Unmarshal(rec.Body.Bytes(), &createBlogResponse)
	if err != nil {
		t.Error(err)
	}

	id := createBlogResponse.Data.ID

	//try to comment on the post
	r := mux.NewRouter()
	r.HandleFunc("/blog/{id}/comment", createBlogCommentHandler)
	r.HandleFunc("/blog/{id}/comment/{commentID}", getSingleBlogCommentHandler)

	ts := httptest.NewServer(r)
	defer ts.Close()

	rec = httptest.NewRecorder()
	req, err = http.NewRequest(http.MethodPost, ts.URL+"/blog/"+id+"/comment", strings.NewReader("{\"AuthorName\": \"Anony Mouse\",\"CommentText\": \"this review sucks\"}"))
	if err != nil {
		t.Error(err)
	}

	r.ServeHTTP(rec, req)

	actual = rec.Result()

	if actual.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, actual.StatusCode)
	}

	var createCommentResponse expectedResponseCreateBlogPostOrComment
	err = json.Unmarshal(rec.Body.Bytes(), &createCommentResponse)
	if err != nil {
		t.Error(err)
	}

	commentID := createCommentResponse.Data.ID

	rec = httptest.NewRecorder()
	req, err = http.NewRequest(http.MethodPost, ts.URL+"/blog/"+id+"/comment/" + commentID, nil)
	if err != nil {
		t.Error(err)
	}

	r.ServeHTTP(rec, req)

	actual = rec.Result()

	if actual.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, actual.StatusCode)
	}

	var actualResponse expectedResponseGetComment
	err = json.Unmarshal(rec.Body.Bytes(), &actualResponse)
	if err != nil {
		t.Error(err)
	}

	expectedComment := db.BlogComment{ArticleID: id, AuthorName: "Anony Mouse", CommentText: "this review sucks", ID: commentID}

	if actualResponse.Data != expectedComment {
		t.Errorf("Expected response to be %#v, got %#v", expectedComment, actualResponse)
	}
}

//TODO: Test deleteBlogCommentHandler