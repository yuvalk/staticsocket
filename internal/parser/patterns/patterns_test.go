package patterns

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"

	"github.com/yuvalk/staticsocket/pkg/types"
)

func TestPatternMatcher_MatchIngressPatterns(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected *types.SocketInfo
	}{
		{
			name: "HTTP ListenAndServe with port",
			code: `package main
import "net/http"
func main() {
	http.ListenAndServe(":8080", nil)
}`,
			expected: &types.SocketInfo{
				Type:            types.TrafficTypeIngress,
				Protocol:        types.ProtocolHTTP,
				RawValue:        ":8080",
				PatternMatch:    "http.ListenAndServe",
				IsResolved:      true,
				ListenPort:      intPtr(8080),
				ListenInterface: "0.0.0.0",
			},
		},
		{
			name: "HTTPS ListenAndServeTLS",
			code: `package main
import "net/http"
func main() {
	http.ListenAndServeTLS(":8443", "cert.pem", "key.pem", nil)
}`,
			expected: &types.SocketInfo{
				Type:            types.TrafficTypeIngress,
				Protocol:        types.ProtocolHTTPS,
				RawValue:        ":8443",
				PatternMatch:    "http.ListenAndServeTLS",
				IsResolved:      true,
				ListenPort:      intPtr(8443),
				ListenInterface: "0.0.0.0",
			},
		},
		{
			name: "TCP net.Listen",
			code: `package main
import "net"
func main() {
	net.Listen("tcp", "localhost:9090")
}`,
			expected: &types.SocketInfo{
				Type:            types.TrafficTypeIngress,
				Protocol:        types.ProtocolTCP,
				RawValue:        "localhost:9090",
				PatternMatch:    "net.Listen",
				IsResolved:      true,
				ListenPort:      intPtr(9090),
				ListenInterface: "localhost",
			},
		},
		{
			name: "UDP net.ListenUDP",
			code: `package main
import "net"
func main() {
	net.ListenUDP("udp", &net.UDPAddr{Port: 5353})
}`,
			expected: &types.SocketInfo{
				Type:         types.TrafficTypeIngress,
				Protocol:     types.ProtocolUDP,
				RawValue:     "",
				PatternMatch: "net.ListenUDP",
				IsResolved:   false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, "test.go", tt.code, 0)
			if err != nil {
				t.Fatalf("Failed to parse code: %v", err)
			}

			pm := NewPatternMatcher()
			var result *types.SocketInfo

			ast.Inspect(file, func(n ast.Node) bool {
				if call, ok := n.(*ast.CallExpr); ok {
					if socket := pm.MatchSocketPattern(call, file); socket != nil {
						result = socket
						return false
					}
				}
				return true
			})

			if result == nil {
				t.Fatal("Expected to find a socket pattern, but found none")
			}

			if result.Type != tt.expected.Type {
				t.Errorf("Type: expected %s, got %s", tt.expected.Type, result.Type)
			}
			if result.Protocol != tt.expected.Protocol {
				t.Errorf("Protocol: expected %s, got %s", tt.expected.Protocol, result.Protocol)
			}
			if result.PatternMatch != tt.expected.PatternMatch {
				t.Errorf("PatternMatch: expected %s, got %s", tt.expected.PatternMatch, result.PatternMatch)
			}
			if result.RawValue != tt.expected.RawValue {
				t.Errorf("RawValue: expected %s, got %s", tt.expected.RawValue, result.RawValue)
			}
			if result.IsResolved != tt.expected.IsResolved {
				t.Errorf("IsResolved: expected %t, got %t", tt.expected.IsResolved, result.IsResolved)
			}

			if tt.expected.ListenPort != nil {
				if result.ListenPort == nil {
					t.Error("Expected ListenPort to be set, but it was nil")
				} else if *result.ListenPort != *tt.expected.ListenPort {
					t.Errorf("ListenPort: expected %d, got %d", *tt.expected.ListenPort, *result.ListenPort)
				}
			}

			if tt.expected.ListenInterface != "" {
				if result.ListenInterface != tt.expected.ListenInterface {
					t.Errorf("ListenInterface: expected %s, got %s", tt.expected.ListenInterface, result.ListenInterface)
				}
			}
		})
	}
}

func TestPatternMatcher_MatchEgressPatterns(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected *types.SocketInfo
	}{
		{
			name: "HTTP GET request",
			code: `package main
import "net/http"
func main() {
	http.Get("https://api.example.com/data")
}`,
			expected: &types.SocketInfo{
				Type:            types.TrafficTypeEgress,
				Protocol:        types.ProtocolHTTPS,
				RawValue:        "https://api.example.com/data",
				PatternMatch:    "http.Get",
				IsResolved:      true,
				DestinationHost: stringPtr("api.example.com"),
				DestinationPort: intPtr(443),
			},
		},
		{
			name: "HTTP POST request",
			code: `package main
import "net/http"
func main() {
	http.Post("http://localhost:8080/api", "application/json", nil)
}`,
			expected: &types.SocketInfo{
				Type:            types.TrafficTypeEgress,
				Protocol:        types.ProtocolHTTP,
				RawValue:        "http://localhost:8080/api",
				PatternMatch:    "http.Post",
				IsResolved:      true,
				DestinationHost: stringPtr("localhost"),
				DestinationPort: intPtr(8080),
			},
		},
		{
			name: "TCP net.Dial",
			code: `package main
import "net"
func main() {
	net.Dial("tcp", "database.internal:5432")
}`,
			expected: &types.SocketInfo{
				Type:            types.TrafficTypeEgress,
				Protocol:        types.ProtocolTCP,
				RawValue:        "database.internal:5432",
				PatternMatch:    "net.Dial",
				IsResolved:      true,
				DestinationHost: stringPtr("database.internal"),
				DestinationPort: intPtr(5432),
			},
		},
		{
			name: "TCP net.DialTimeout",
			code: `package main
import ("net"; "time")
func main() {
	net.DialTimeout("tcp", "example.com:443", 5*time.Second)
}`,
			expected: &types.SocketInfo{
				Type:            types.TrafficTypeEgress,
				Protocol:        types.ProtocolTCP,
				RawValue:        "example.com:443",
				PatternMatch:    "net.DialTimeout",
				IsResolved:      true,
				DestinationHost: stringPtr("example.com"),
				DestinationPort: intPtr(443),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, "test.go", tt.code, 0)
			if err != nil {
				t.Fatalf("Failed to parse code: %v", err)
			}

			pm := NewPatternMatcher()
			var result *types.SocketInfo

			ast.Inspect(file, func(n ast.Node) bool {
				if call, ok := n.(*ast.CallExpr); ok {
					if socket := pm.MatchSocketPattern(call, file); socket != nil {
						result = socket
						return false
					}
				}
				return true
			})

			if result == nil {
				t.Fatal("Expected to find a socket pattern, but found none")
			}

			if result.Type != tt.expected.Type {
				t.Errorf("Type: expected %s, got %s", tt.expected.Type, result.Type)
			}
			if result.Protocol != tt.expected.Protocol {
				t.Errorf("Protocol: expected %s, got %s", tt.expected.Protocol, result.Protocol)
			}
			if result.PatternMatch != tt.expected.PatternMatch {
				t.Errorf("PatternMatch: expected %s, got %s", tt.expected.PatternMatch, result.PatternMatch)
			}

			if tt.expected.DestinationHost != nil {
				if result.DestinationHost == nil {
					t.Error("Expected DestinationHost to be set, but it was nil")
				} else if *result.DestinationHost != *tt.expected.DestinationHost {
					t.Errorf("DestinationHost: expected %s, got %s", *tt.expected.DestinationHost, *result.DestinationHost)
				}
			}

			if tt.expected.DestinationPort != nil {
				if result.DestinationPort == nil {
					t.Error("Expected DestinationPort to be set, but it was nil")
				} else if *result.DestinationPort != *tt.expected.DestinationPort {
					t.Errorf("DestinationPort: expected %d, got %d", *tt.expected.DestinationPort, *result.DestinationPort)
				}
			}
		})
	}
}

func TestPatternMatcher_NoMatch(t *testing.T) {
	code := `package main
import "fmt"
func main() {
	fmt.Println("Hello, World!")
}`

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", code, 0)
	if err != nil {
		t.Fatalf("Failed to parse code: %v", err)
	}

	pm := NewPatternMatcher()
	var result *types.SocketInfo

	ast.Inspect(file, func(n ast.Node) bool {
		if call, ok := n.(*ast.CallExpr); ok {
			if socket := pm.MatchSocketPattern(call, file); socket != nil {
				result = socket
				return false
			}
		}
		return true
	})

	if result != nil {
		t.Errorf("Expected no socket pattern match, but found: %+v", result)
	}
}

func TestExtractFunctionName(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected string
	}{
		{
			name: "Package selector",
			code: `package main
func main() {
	http.Get("url")
}`,
			expected: "http.Get",
		},
		{
			name: "Simple identifier",
			code: `package main
func Get(url string) {}
func main() {
	Get("url")
}`,
			expected: "Get",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, "test.go", tt.code, 0)
			if err != nil {
				t.Fatalf("Failed to parse code: %v", err)
			}

			pm := NewPatternMatcher()
			var result string

			ast.Inspect(file, func(n ast.Node) bool {
				if call, ok := n.(*ast.CallExpr); ok {
					result = pm.extractFunctionName(call)
					if result != "" {
						return false
					}
				}
				return true
			})

			if result != tt.expected {
				t.Errorf("Expected function name %s, got %s", tt.expected, result)
			}
		})
	}
}

// Helper functions
func intPtr(i int) *int {
	return &i
}

func stringPtr(s string) *string {
	return &s
}