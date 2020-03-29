package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gogo/protobuf/io"
)

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

	fmt.Println(len(expectedBody), len(actualBody))
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
	io.NewFullReader()
	req, err := http.NewRequest(http.MethodPost, "/blog", nil)
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

	fmt.Println(len(expectedBody), len(actualBody))
	if actualBody != expectedBody {
		t.Errorf("Expected body to be %s, got %s", expectedBody, actualBody)
	}
}

func Test_GetBlogPostIDs(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "/blog", strings.NewReader("{\"Title\":\"I've come to make an announcement\",\"ArticleText\":\"Test ertyertyretyrt\",\"AuthorName\":\"Dr. Eggman\"}"))
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

	expectedBody := "server_up"
	actualBody := rec.Body.String()

	fmt.Println(len(expectedBody), len(actualBody))
	if actualBody != expectedBody {
		t.Errorf("Expected body to be %s, got %s", expectedBody, actualBody)
	}
}