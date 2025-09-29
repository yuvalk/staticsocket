package types

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"gopkg.in/yaml.v3"
)

type TrafficType string

const (
	TrafficTypeIngress TrafficType = "ingress"
	TrafficTypeEgress  TrafficType = "egress"
)

type Protocol string

const (
	ProtocolTCP   Protocol = "tcp"
	ProtocolUDP   Protocol = "udp"
	ProtocolHTTP  Protocol = "http"
	ProtocolHTTPS Protocol = "https"
	ProtocolGRPC  Protocol = "grpc"
	ProtocolUnix  Protocol = "unix"
)

type SocketInfo struct {
	Type         TrafficType `json:"type" yaml:"type"`
	Protocol     Protocol    `json:"protocol" yaml:"protocol"`
	ProcessName  string      `json:"process_name" yaml:"process_name"`
	SourceFile   string      `json:"source_file" yaml:"source_file"`
	SourceLine   int         `json:"source_line" yaml:"source_line"`
	FunctionName string      `json:"function_name" yaml:"function_name"`
	
	// Ingress-specific fields
	ListenPort      *int    `json:"listen_port,omitempty" yaml:"listen_port,omitempty"`
	ListenInterface string  `json:"listen_interface,omitempty" yaml:"listen_interface,omitempty"`
	
	// Egress-specific fields
	DestinationHost *string `json:"destination_host,omitempty" yaml:"destination_host,omitempty"`
	DestinationPort *int    `json:"destination_port,omitempty" yaml:"destination_port,omitempty"`
	
	// Additional metadata
	IsResolved   bool   `json:"is_resolved" yaml:"is_resolved"`
	RawValue     string `json:"raw_value" yaml:"raw_value"`
	PatternMatch string `json:"pattern_match" yaml:"pattern_match"`
}

type AnalysisResults struct {
	Sockets     []SocketInfo `json:"sockets" yaml:"sockets"`
	TotalCount  int          `json:"total_count" yaml:"total_count"`
	IngressCount int         `json:"ingress_count" yaml:"ingress_count"`
	EgressCount  int         `json:"egress_count" yaml:"egress_count"`
	ProcessName  string      `json:"process_name" yaml:"process_name"`
}

func (r *AnalysisResults) Export(writer io.Writer, format string) error {
	switch strings.ToLower(format) {
	case "json":
		encoder := json.NewEncoder(writer)
		encoder.SetIndent("", "  ")
		return encoder.Encode(r)
	case "yaml":
		encoder := yaml.NewEncoder(writer)
		defer encoder.Close()
		return encoder.Encode(r)
	case "csv":
		return r.exportCSV(writer)
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}
}

func (r *AnalysisResults) exportCSV(writer io.Writer) error {
	csvWriter := csv.NewWriter(writer)
	defer csvWriter.Flush()

	headers := []string{
		"Type", "Protocol", "ProcessName", "SourceFile", "SourceLine", "FunctionName",
		"ListenPort", "ListenInterface", "DestinationHost", "DestinationPort",
		"IsResolved", "RawValue", "PatternMatch",
	}

	if err := csvWriter.Write(headers); err != nil {
		return err
	}

	for _, socket := range r.Sockets {
		record := []string{
			string(socket.Type),
			string(socket.Protocol),
			socket.ProcessName,
			socket.SourceFile,
			fmt.Sprintf("%d", socket.SourceLine),
			socket.FunctionName,
			formatIntPtr(socket.ListenPort),
			socket.ListenInterface,
			formatStringPtr(socket.DestinationHost),
			formatIntPtr(socket.DestinationPort),
			fmt.Sprintf("%t", socket.IsResolved),
			socket.RawValue,
			socket.PatternMatch,
		}
		if err := csvWriter.Write(record); err != nil {
			return err
		}
	}

	return nil
}

func formatIntPtr(ptr *int) string {
	if ptr == nil {
		return ""
	}
	return fmt.Sprintf("%d", *ptr)
}

func formatStringPtr(ptr *string) string {
	if ptr == nil {
		return ""
	}
	return *ptr
}