package types

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestSocketInfo_JSONExport(t *testing.T) {
	port := 8080
	host := "example.com"

	socket := SocketInfo{
		Type:            TrafficTypeIngress,
		Protocol:        ProtocolHTTP,
		ProcessName:     "test-service",
		SourceFile:      "/path/to/file.go",
		SourceLine:      42,
		FunctionName:    "main",
		ListenPort:      &port,
		ListenInterface: "0.0.0.0",
		DestinationHost: &host,
		IsResolved:      true,
		RawValue:        ":8080",
		PatternMatch:    "http.ListenAndServe",
	}

	data, err := json.Marshal(socket)
	if err != nil {
		t.Fatalf("Failed to marshal socket: %v", err)
	}

	var unmarshaled SocketInfo
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal socket: %v", err)
	}

	if unmarshaled.Type != socket.Type {
		t.Errorf("Expected type %s, got %s", socket.Type, unmarshaled.Type)
	}
	if *unmarshaled.ListenPort != *socket.ListenPort {
		t.Errorf("Expected port %d, got %d", *socket.ListenPort, *unmarshaled.ListenPort)
	}
}

func TestAnalysisResults_ExportJSON(t *testing.T) {
	port := 3000
	results := AnalysisResults{
		Sockets: []SocketInfo{
			{
				Type:            TrafficTypeIngress,
				Protocol:        ProtocolHTTP,
				ProcessName:     "web-server",
				SourceFile:      "main.go",
				SourceLine:      10,
				ListenPort:      &port,
				ListenInterface: "0.0.0.0",
				IsResolved:      true,
				RawValue:        ":3000",
				PatternMatch:    "http.ListenAndServe",
			},
		},
		TotalCount:   1,
		IngressCount: 1,
		EgressCount:  0,
		ProcessName:  "web-server",
	}

	var buf bytes.Buffer
	err := results.Export(&buf, "json")
	if err != nil {
		t.Fatalf("Failed to export JSON: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, `"total_count": 1`) {
		t.Error("JSON output missing total_count")
	}
	if !strings.Contains(output, `"ingress_count": 1`) {
		t.Error("JSON output missing ingress_count")
	}
	if !strings.Contains(output, `"listen_port": 3000`) {
		t.Error("JSON output missing listen_port")
	}
}

func TestAnalysisResults_ExportCSV(t *testing.T) {
	port := 8080
	host := "api.example.com"

	results := AnalysisResults{
		Sockets: []SocketInfo{
			{
				Type:            TrafficTypeEgress,
				Protocol:        ProtocolHTTPS,
				ProcessName:     "client",
				SourceFile:      "client.go",
				SourceLine:      25,
				DestinationHost: &host,
				DestinationPort: &port,
				IsResolved:      true,
				RawValue:        "https://api.example.com:8080",
				PatternMatch:    "http.Get",
			},
		},
		TotalCount:  1,
		EgressCount: 1,
	}

	var buf bytes.Buffer
	err := results.Export(&buf, "csv")
	if err != nil {
		t.Fatalf("Failed to export CSV: %v", err)
	}

	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) != 2 {
		t.Errorf("Expected 2 lines (header + data), got %d", len(lines))
	}

	header := lines[0]
	if !strings.Contains(header, "Type,Protocol,ProcessName") {
		t.Error("CSV header missing expected columns")
	}

	data := lines[1]
	if !strings.Contains(data, "egress,https,client") {
		t.Error("CSV data missing expected values")
	}
}

func TestAnalysisResults_ExportYAML(t *testing.T) {
	port := 9090
	results := AnalysisResults{
		Sockets: []SocketInfo{
			{
				Type:            TrafficTypeIngress,
				Protocol:        ProtocolTCP,
				ProcessName:     "tcp-server",
				SourceFile:      "server.go",
				SourceLine:      15,
				ListenPort:      &port,
				ListenInterface: "127.0.0.1",
				IsResolved:      true,
				RawValue:        "127.0.0.1:9090",
				PatternMatch:    "net.Listen",
			},
		},
		TotalCount:   1,
		IngressCount: 1,
	}

	var buf bytes.Buffer
	err := results.Export(&buf, "yaml")
	if err != nil {
		t.Fatalf("Failed to export YAML: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "total_count: 1") {
		t.Error("YAML output missing total_count")
	}
	if !strings.Contains(output, "listen_port: 9090") {
		t.Error("YAML output missing listen_port")
	}
	if !strings.Contains(output, "protocol: tcp") {
		t.Error("YAML output missing protocol")
	}
}

func TestAnalysisResults_ExportUnsupportedFormat(t *testing.T) {
	results := AnalysisResults{}
	var buf bytes.Buffer

	err := results.Export(&buf, "xml")
	if err == nil {
		t.Error("Expected error for unsupported format")
	}
	if !strings.Contains(err.Error(), "unsupported format") {
		t.Errorf("Expected 'unsupported format' error, got: %v", err)
	}
}

func TestFormatIntPtr(t *testing.T) {
	tests := []struct {
		input    *int
		expected string
	}{
		{nil, ""},
		{intPtr(0), "0"},
		{intPtr(8080), "8080"},
		{intPtr(-1), "-1"},
	}

	for _, test := range tests {
		result := formatIntPtr(test.input)
		if result != test.expected {
			t.Errorf("formatIntPtr(%v) = %q, expected %q", test.input, result, test.expected)
		}
	}
}

func TestFormatStringPtr(t *testing.T) {
	tests := []struct {
		input    *string
		expected string
	}{
		{nil, ""},
		{stringPtr(""), ""},
		{stringPtr("localhost"), "localhost"},
		{stringPtr("example.com"), "example.com"},
	}

	for _, test := range tests {
		result := formatStringPtr(test.input)
		if result != test.expected {
			t.Errorf("formatStringPtr(%v) = %q, expected %q", test.input, result, test.expected)
		}
	}
}

// Helper functions for test setup
func intPtr(i int) *int {
	return &i
}

func stringPtr(s string) *string {
	return &s
}
