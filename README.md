# StaticSocket

A powerful static analysis tool for identifying socket creation patterns in source code. StaticSocket analyzes your codebase to detect network connections, classify them as ingress (incoming) or egress (outgoing) traffic, and extract detailed metadata including hosts, ports, and protocols.

**Current Implementation**: Go language support  
**Roadmap**: Multi-language support for Python, Java, C++, Rust, and more

[![CI](https://github.com/yuvalk/staticsocket/workflows/CI/badge.svg)](https://github.com/yuvalk/staticsocket/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/yuvalk/staticsocket)](https://goreportcard.com/report/github.com/yuvalk/staticsocket)
[![License](https://img.shields.io/github/license/yuvalk/staticsocket)](LICENSE)

## Features

### 🔍 **Comprehensive Socket Detection**
- **HTTP/HTTPS servers**: `http.ListenAndServe`, `http.ListenAndServeTLS` (Go)
- **TCP/UDP listeners**: `net.Listen`, `net.ListenTCP`, `net.ListenUDP` (Go)
- **Outbound connections**: `net.Dial`, `net.DialTimeout`, `http.Get`, `http.Post` (Go)
- **Framework support**: Detects patterns across popular Go networking libraries
- **Future languages**: Python (socket, requests), Java (ServerSocket, HttpClient), C++ (Boost.Asio), Rust (tokio)

### 📊 **Traffic Classification**
- **Ingress Traffic**: Servers, listeners, and services accepting connections
- **Egress Traffic**: Outbound HTTP requests, database connections, API calls

### 🧠 **Intelligent Resolution**
- **String literals**: Direct parsing of hardcoded URLs and addresses
- **Constants**: Resolves `const` declarations throughout the codebase (Go)
- **Variables**: Smart pattern recognition for common variable types (Go)
- **Dynamic patterns**: httptest servers, API URLs, environment variables (Go)
- **Language-specific**: Adapts resolution strategies per language

### 📋 **Multiple Output Formats**
- **JSON**: Structured data for programmatic consumption
- **YAML**: Human-readable configuration format
- **CSV**: Spreadsheet-compatible tabular output

## Installation

### Download Binary
```bash
# Download latest release
curl -L https://github.com/yuvalk/staticsocket/releases/latest/download/staticsocket-linux-amd64.tar.gz | tar xz
sudo mv staticsocket /usr/local/bin/
```

### Build from Source
```bash
git clone https://github.com/yuvalk/staticsocket.git
cd staticsocket
go build -o staticsocket .
```

### Docker
```bash
docker run --rm -v $(pwd):/workspace staticsocket/staticsocket -path /workspace
```

## Quick Start

### Analyze a Go File
```bash
staticsocket -path main.go -format json
```

### Analyze a Go Project
```bash
staticsocket -path ./src -format yaml
```

### Save Results to File
```bash
staticsocket -path . -format csv -output results.csv
```

## Usage Examples

### Basic Analysis
```go
// example.go
package main

import (
    "net/http"
    "net"
)

func main() {
    // HTTP server (ingress)
    http.ListenAndServe(":8080", nil)
    
    // Database connection (egress)
    conn, _ := net.Dial("tcp", "database.internal:5432")
    defer conn.Close()
    
    // API call (egress)
    http.Get("https://api.github.com/user")
}
```

### Analysis Output
```bash
$ staticsocket -path example.go -format json
```

```json
{
  "sockets": [
    {
      "type": "ingress",
      "protocol": "http",
      "process_name": "main",
      "source_file": "example.go",
      "source_line": 10,
      "listen_port": 8080,
      "listen_interface": "0.0.0.0",
      "is_resolved": true,
      "raw_value": ":8080",
      "pattern_match": "http.ListenAndServe"
    },
    {
      "type": "egress",
      "protocol": "tcp",
      "process_name": "main",
      "source_file": "example.go",
      "source_line": 13,
      "destination_host": "database.internal",
      "destination_port": 5432,
      "is_resolved": true,
      "raw_value": "database.internal:5432",
      "pattern_match": "net.Dial"
    },
    {
      "type": "egress",
      "protocol": "https",
      "process_name": "main",
      "source_file": "example.go",
      "source_line": 17,
      "destination_host": "api.github.com",
      "destination_port": 443,
      "is_resolved": true,
      "raw_value": "https://api.github.com/user",
      "pattern_match": "http.Get"
    }
  ],
  "total_count": 3,
  "ingress_count": 1,
  "egress_count": 2,
  "process_name": "main"
}
```

## Advanced Features

### Variable Resolution
StaticSocket can resolve various types of dynamic values:

```go
// Constants
const apiURL = "https://api.service.com"
http.Get(apiURL)  // ✅ Resolves to api.service.com:443

// httptest servers
server := httptest.NewServer(handler)
http.Post(server.URL, "application/json", nil)  // ✅ Resolves to localhost

// Environment variables (pattern-based)
apiURL := os.Getenv("API_URL")
http.Get(apiURL)  // ✅ Resolves to external-service
```

### Framework Detection (Go)
Automatically detects socket usage in popular Go frameworks:
- **HTTP frameworks**: Gin, Echo, Fiber
- **gRPC**: Server and client connections
- **Database drivers**: SQL, MongoDB connections

*Framework detection for other languages planned in future releases*

## Command Line Options

```bash
staticsocket [options]

Options:
  -path string        Path to analyze (file or directory) (default ".")
  -format string      Output format: json, yaml, csv (default "json")
  -output string      Output file (default: stdout)
  -verbose           Enable verbose output
  -help              Show help message

Note: Currently supports Go files (.go). Other languages coming soon.
```

## Use Cases

### 🛡️ **Security Analysis**
- Audit network connections in microservices
- Identify external dependencies and API calls
- Map ingress/egress traffic for security policies

### 🏗️ **Architecture Documentation**
- Generate network topology diagrams
- Document service dependencies
- Create deployment configurations

### 📋 **Compliance & Governance**
- SBOM (Software Bill of Materials) generation
- Network security assessments
- Regulatory compliance reporting

### 🚀 **DevOps Integration**
- CI/CD pipeline integration
- Infrastructure as Code validation
- Container security scanning

## Development

### Prerequisites
- Go 1.19 or later
- make (optional, for convenience)

### Building
```bash
# Build binary
make build

# Run tests
make test

# Run linting
make lint

# Build for all platforms
make build-all
```

### Testing
```bash
# Run all tests
go test ./...

# Run tests with coverage
make coverage

# Run integration tests
make demo
```

### Project Structure
```
├── main.go                           # CLI entry point
├── pkg/
│   ├── analyzer/                     # Main analysis engine
│   └── types/                        # Data structures & export
├── internal/
│   ├── parser/patterns/              # Socket pattern matching
│   └── resolver/                     # Variable resolution
├── testdata/                         # Test fixtures
├── .github/workflows/                # CI/CD pipelines
└── docs/                            # Documentation
```

## Contributing

We welcome contributions! Please see our [Contributing Guide](CONTRIBUTING.md) for details.

### 🐛 **Reporting Issues**

Found a bug or have a feature request? Please [open an issue](https://github.com/yuvalk/staticsocket/issues/new) with:

**For Bugs:**
- Go version and OS
- Sample code that reproduces the issue
- Expected vs actual behavior
- StaticSocket version

**For Feature Requests:**
- Clear description of the proposed feature
- Use case and motivation
- Example of how it would work

### 🔧 **Pull Requests**

1. **Fork** the repository
2. **Create** a feature branch: `git checkout -b feature/my-feature`
3. **Add tests** for your changes
4. **Ensure** all tests pass: `make test`
5. **Run** linting: `make lint`
6. **Commit** with clear messages
7. **Push** to your fork
8. **Submit** a pull request

#### Pull Request Guidelines:
- **Include tests** for new functionality
- **Update documentation** if needed
- **Follow Go conventions** and existing code style
- **Keep commits focused** and atomic
- **Add examples** for new socket patterns

#### Development Workflow:
```bash
# Setup development environment
git clone https://github.com/yuvalk/staticsocket.git
cd staticsocket
go mod download

# Make your changes
git checkout -b feature/my-feature

# Test your changes
make test
make lint

# Test with real codebases
./staticsocket -path /path/to/go/project

# Submit pull request
git push origin feature/my-feature
```

### 🎯 **Good First Issues**

New contributors can look for issues labeled:
- `good-first-issue`: Perfect for newcomers
- `help-wanted`: Community input needed
- `documentation`: Improve docs and examples

### 📝 **Areas for Contribution**

- **Multi-language support**: Add Python, Java, C++, Rust analyzers
- **New socket patterns**: Add support for more networking libraries per language
- **Enhanced resolution**: Improve variable and constant resolution across languages
- **Output formats**: Add new export formats (XML, Prometheus, etc.)
- **Performance**: Optimize analysis for large codebases
- **Documentation**: Examples, tutorials, and guides

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- Inspired by network security analysis needs in cloud-native environments
- Built using Go's powerful AST parsing capabilities
- Thanks to the Go community for excellent static analysis tools

---

**Made with ❤️ for the Go community**

For more information, visit our [documentation](https://github.com/yuvalk/staticsocket/wiki) or join the discussion in [GitHub Discussions](https://github.com/yuvalk/staticsocket/discussions).