package resolver

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"

	"github.com/yuvalk/staticsocket/pkg/types"
)

func TestValueResolver_ResolveHttpTestServer(t *testing.T) {
	code := `package main

import (
	"net/http"
	"net/http/httptest"
)

func testHandler() {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	server := httptest.NewServer(handler)
	defer server.Close()
	
	http.Post(server.URL, "application/json", nil)
}`

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", code, 0)
	if err != nil {
		t.Fatalf("Failed to parse code: %v", err)
	}

	resolver := New()
	
	// Find the http.Post call
	var callExpr *ast.CallExpr
	ast.Inspect(file, func(n ast.Node) bool {
		if call, ok := n.(*ast.CallExpr); ok {
			if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
				if ident, ok := sel.X.(*ast.Ident); ok && ident.Name == "http" && sel.Sel.Name == "Post" {
					callExpr = call
					return false
				}
			}
		}
		return true
	})

	if callExpr == nil {
		t.Fatal("Could not find http.Post call")
	}

	socket := &types.SocketInfo{
		Type:         types.TrafficTypeEgress,
		Protocol:     types.ProtocolHTTP,
		PatternMatch: "http.Post",
	}

	// Test resolution
	resolver.ResolveValues(socket, callExpr, file)

	// Should detect that this is likely a local test server
	if !socket.IsResolved {
		t.Error("Expected socket to be resolved for httptest server pattern")
	}
}

func TestValueResolver_ResolveConstantURL(t *testing.T) {
	code := `package main

import "net/http"

const baseURL = "https://api.example.com"

func makeRequest() {
	http.Get(baseURL + "/users")
}`

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", code, 0)
	if err != nil {
		t.Fatalf("Failed to parse code: %v", err)
	}

	resolver := New()
	
	// Find the http.Get call
	var callExpr *ast.CallExpr
	ast.Inspect(file, func(n ast.Node) bool {
		if call, ok := n.(*ast.CallExpr); ok {
			if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
				if ident, ok := sel.X.(*ast.Ident); ok && ident.Name == "http" && sel.Sel.Name == "Get" {
					callExpr = call
					return false
				}
			}
		}
		return true
	})

	if callExpr == nil {
		t.Fatal("Could not find http.Get call")
	}

	socket := &types.SocketInfo{
		Type:         types.TrafficTypeEgress,
		Protocol:     types.ProtocolHTTP,
		PatternMatch: "http.Get",
	}

	// Test resolution
	resolver.ResolveValues(socket, callExpr, file)

	// Should resolve the base URL part
	if !socket.IsResolved {
		t.Error("Expected socket to be resolved for constant URL")
	}

	if socket.DestinationHost == nil || *socket.DestinationHost != "api.example.com" {
		t.Errorf("Expected host to be api.example.com, got %v", socket.DestinationHost)
	}
}

func TestValueResolver_DetectCommonPatterns(t *testing.T) {
	tests := []struct {
		name         string
		varName      string
		expectedHost string
		expectedPort int
	}{
		{
			name:         "httptest server",
			varName:      "server.URL",
			expectedHost: "localhost",
			expectedPort: 0, // Dynamic port
		},
		{
			name:         "environment variable",
			varName:      "os.Getenv(\"API_URL\")",
			expectedHost: "", // Can't resolve
			expectedPort: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolver := New()
			
			// Test pattern detection
			host, port, resolved := resolver.analyzeVariablePattern(tt.varName)
			
			if tt.expectedHost != "" && (!resolved || host != tt.expectedHost) {
				t.Errorf("Expected host %s, got %s (resolved: %t)", tt.expectedHost, host, resolved)
			}
			
			if tt.expectedPort > 0 && port != tt.expectedPort {
				t.Errorf("Expected port %d, got %d", tt.expectedPort, port)
			}
		})
	}
}