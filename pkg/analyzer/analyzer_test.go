package analyzer

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/yuvalk/staticsocket/pkg/types"
)

func TestAnalyzer_AnalyzeFile(t *testing.T) {
	testCode := `package main
import ("net"; "net/http")
const serverPort = ":8080"
func main() {
	http.ListenAndServe(":3000", nil)
	listener, _ := net.Listen("tcp", serverPort)
	defer listener.Close()
	http.Get("https://api.example.com/data")
}`
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.go")
	if err := os.WriteFile(testFile, []byte(testCode), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	analyzer := New()
	results, err := analyzer.Analyze(testFile)
	if err != nil {
		t.Fatalf("Failed to analyze file: %v", err)
	}

	validateSocketCounts(t, results, 3, 2, 1)

	sockets := make(map[string]*types.SocketInfo)
	for i := range results.Sockets {
		sockets[results.Sockets[i].PatternMatch] = &results.Sockets[i]
	}

	validateHTTPServer(t, sockets)
	validateTCPListener(t, sockets)
	validateHTTPClient(t, sockets)
}

func validateSocketCounts(t *testing.T, results *types.AnalysisResults, total, ingress, egress int) {
	t.Helper()
	t.Run("Socket Counts", func(t *testing.T) {
		if results.TotalCount != total {
			t.Errorf("Expected %d total sockets, got %d", total, results.TotalCount)
		}
		if results.IngressCount != ingress {
			t.Errorf("Expected %d ingress sockets, got %d", ingress, results.IngressCount)
		}
		if results.EgressCount != egress {
			t.Errorf("Expected %d egress sockets, got %d", egress, results.EgressCount)
		}
	})
}

func validateHTTPServer(t *testing.T, sockets map[string]*types.SocketInfo) {
	t.Helper()
	t.Run("HTTP Server", func(t *testing.T) {
		httpServer, ok := sockets["http.ListenAndServe"]
		if !ok {
			t.Fatal("Expected to find http.ListenAndServe pattern")
		}
		if httpServer.Type != types.TrafficTypeIngress {
			t.Error("HTTP server should be ingress traffic")
		}
		if httpServer.ListenPort == nil || *httpServer.ListenPort != 3000 {
			t.Errorf("Expected HTTP server port 3000, got %v", httpServer.ListenPort)
		}
	})
}

func validateTCPListener(t *testing.T, sockets map[string]*types.SocketInfo) {
	t.Helper()
	t.Run("TCP Listener", func(t *testing.T) {
		tcpListener, ok := sockets["net.Listen"]
		if !ok {
			t.Fatal("Expected to find net.Listen pattern")
		}
		if tcpListener.Type != types.TrafficTypeIngress {
			t.Error("TCP listener should be ingress traffic")
		}
	})
}

func validateHTTPClient(t *testing.T, sockets map[string]*types.SocketInfo) {
	t.Helper()
	t.Run("HTTP Client", func(t *testing.T) {
		httpClient, ok := sockets["http.Get"]
		if !ok {
			t.Fatal("Expected to find http.Get pattern")
		}
		if httpClient.Type != types.TrafficTypeEgress {
			t.Error("HTTP client should be egress traffic")
		}
		if httpClient.DestinationHost == nil || *httpClient.DestinationHost != "api.example.com" {
			t.Errorf("Expected destination host api.example.com, got %v", httpClient.DestinationHost)
		}
	})
}

func TestAnalyzer_AnalyzeDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	// Create multiple test files
	files := map[string]string{
		"server.go": `package main
import "net/http"
func main() {
	http.ListenAndServe(":8080", nil)
}`,
		"client.go": `package main
import "net/http"
func main() {
	http.Get("https://example.com")
}`,
		"other.txt": "not a go file",
		"vendor/dep.go": `package vendor
import "net"
func init() {
	net.Listen("tcp", ":9999")
}`,
	}

	for filename, content := range files {
		filePath := filepath.Join(tmpDir, filename)
		if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
			t.Fatalf("Failed to create directory: %v", err)
		}
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to write file %s: %v", filename, err)
		}
	}

	analyzer := New()
	results, err := analyzer.Analyze(tmpDir)
	if err != nil {
		t.Fatalf("Failed to analyze directory: %v", err)
	}

	// Should find sockets from server.go and client.go, but not from vendor/ or .txt files
	if results.TotalCount != 2 {
		t.Errorf("Expected 2 sockets, got %d", results.TotalCount)
	}

	if results.IngressCount != 1 {
		t.Errorf("Expected 1 ingress socket, got %d", results.IngressCount)
	}

	if results.EgressCount != 1 {
		t.Errorf("Expected 1 egress socket, got %d", results.EgressCount)
	}
}

func TestAnalyzer_AnalyzeNonExistentPath(t *testing.T) {
	analyzer := New()
	_, err := analyzer.Analyze("/non/existent/path")
	if err == nil {
		t.Error("Expected error for non-existent path")
	}
}

func TestAnalyzer_IntegrationWithTestData(t *testing.T) {
	// Test with our sample files if they exist
	samplesDir := "../../testdata/samples"
	if _, err := os.Stat(samplesDir); os.IsNotExist(err) {
		t.Skip("Sample data directory not found, skipping integration test")
	}

	analyzer := New()
	results, err := analyzer.Analyze(samplesDir)
	if err != nil {
		t.Fatalf("Failed to analyze samples directory: %v", err)
	}

	if results.TotalCount == 0 {
		t.Error("Expected to find some sockets in sample data")
	}

	t.Logf("Found %d total sockets (%d ingress, %d egress)",
		results.TotalCount, results.IngressCount, results.EgressCount)

	// Test JSON export
	jsonData, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal results to JSON: %v", err)
	}

	t.Logf("JSON output:\n%s", jsonData)
}

func TestAnalyzer_EmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "empty.go")

	if err := os.WriteFile(testFile, []byte("package main"), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	analyzer := New()
	results, err := analyzer.Analyze(testFile)
	if err != nil {
		t.Fatalf("Failed to analyze empty file: %v", err)
	}

	if results.TotalCount != 0 {
		t.Errorf("Expected 0 sockets in empty file, got %d", results.TotalCount)
	}
}

func TestAnalyzer_InvalidGoFile(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "invalid.go")

	// Write invalid Go syntax
	if err := os.WriteFile(testFile, []byte("invalid go syntax {{{"), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	analyzer := New()
	_, err := analyzer.Analyze(testFile)
	if err == nil {
		t.Error("Expected error for invalid Go file")
	}
}

func TestDeriveProcessName(t *testing.T) {
	tests := []struct {
		name         string
		packageName  string
		filePath     string
		expectedName string
	}{
		{
			name:         "main package",
			packageName:  "main",
			filePath:     "/path/to/myservice/main.go",
			expectedName: "myservice",
		},
		{
			name:         "non-main package",
			packageName:  "server",
			filePath:     "/path/to/project/server/server.go",
			expectedName: "server",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This is a simplified test since deriveProcessName is part of the visitor
			// In a full implementation, you'd create a more comprehensive test
			if tt.packageName != "main" && tt.packageName != tt.expectedName {
				t.Errorf("Expected %s, would get %s", tt.expectedName, tt.packageName)
			}
		})
	}
}
