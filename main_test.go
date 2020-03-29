package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSample(t *testing.T) {
	return
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

	fmt.Println(len(expectedBody), len(actualBody))
	if actualBody != expectedBody {
		t.Errorf("Expected body to be %s, got %s", expectedBody, actualBody)
	}
}
