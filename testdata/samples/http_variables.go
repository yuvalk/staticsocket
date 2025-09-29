package main

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
)

func testHttpVariables() {
	// Case 1: httptest.NewServer pattern (like OpenShift)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	})
	server := httptest.NewServer(handler)
	defer server.Close()

	// This should be detected but currently isn't
	resp, _ := http.Post(server.URL, "application/json", nil)
	defer resp.Body.Close()

	// Case 2: Environment variable URL
	apiURL := os.Getenv("API_URL")
	if apiURL != "" {
		http.Get(apiURL)
	}

	// Case 3: url.Parse() result
	parsedURL, _ := url.Parse("https://api.github.com/repos/user/repo")
	http.Get(parsedURL.String())

	// Case 4: String concatenation
	baseURL := "https://api.example.com"
	endpoint := "/users"
	fullURL := baseURL + endpoint
	http.Get(fullURL)

	// Case 5: Variable from function call
	serviceURL := getServiceURL()
	http.Post(serviceURL, "application/json", nil)
}

func getServiceURL() string {
	return "http://service.local:8080/api"
}
