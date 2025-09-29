package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// This mimics the OpenShift webhook_test.go pattern
func TestParseUrlError(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(400)
	})
	server := httptest.NewServer(handler)
	defer server.Close()

	// This is the exact pattern from OpenShift that wasn't being detected
	_, err := http.Post(server.URL, "application/json", nil)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}
