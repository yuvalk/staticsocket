# Contributing to StaticSocket

Thank you for your interest in contributing to StaticSocket! This document provides guidelines and information for contributors.

## 🚀 Getting Started

### Prerequisites
- Go 1.19 or later
- Git
- Make (optional but recommended)

### Development Setup
```bash
# Fork and clone the repository
git clone https://github.com/yourusername/staticsocket.git
cd staticsocket

# Install dependencies
go mod download

# Verify everything works
make test
make build
```

## 📋 How to Contribute

### 🐛 Reporting Bugs

Before creating bug reports, please check if the issue already exists. When creating a bug report, include:

**Required Information:**
- Go version: `go version`
- Operating system and version
- StaticSocket version or commit hash
- Complete error message or unexpected output

**Bug Report Template:**
```markdown
**Describe the bug**
A clear description of what the bug is.

**To Reproduce**
Steps to reproduce the behavior:
1. Create a Go file with '...'
2. Run `staticsocket -path ...`
3. See error

**Expected behavior**
What you expected to happen.

**Sample Code**
```go
// Minimal reproducible example
package main
// ...
```

**Environment:**
- OS: [e.g. Ubuntu 22.04]
- Go version: [e.g. 1.21.0]
- StaticSocket version: [e.g. v1.0.0]

**Additional context**
Any other context about the problem.
```

### 💡 Feature Requests

Feature requests are welcome! Please provide:

1. **Clear description** of the feature
2. **Use case** and motivation
3. **Example** of how it would work
4. **Impact** on existing functionality

### 🔧 Code Contributions

#### Pull Request Process

1. **Fork** the repository
2. **Create** a feature branch from `main`
3. **Make** your changes
4. **Add** tests for your changes
5. **Update** documentation if needed
6. **Ensure** all tests pass
7. **Submit** a pull request

#### Pull Request Guidelines

**Before Submitting:**
- [ ] All tests pass (`make test`)
- [ ] Code follows Go conventions (`make lint`)
- [ ] Documentation updated if needed
- [ ] Commit messages are clear and descriptive

**PR Title Format:**
- `feat: add support for gRPC client detection`
- `fix: resolve variable parsing in nested functions`
- `docs: update installation instructions`
- `test: add coverage for HTTP framework patterns`

**PR Description Template:**
```markdown
## Description
Brief description of changes

## Type of Change
- [ ] Bug fix
- [ ] New feature
- [ ] Breaking change
- [ ] Documentation update

## Testing
- [ ] Added unit tests
- [ ] Added integration tests
- [ ] Tested manually with sample projects

## Related Issues
Fixes #123
```

## 🏗️ Development Guidelines

### Code Style

**Follow Go Conventions:**
- Use `gofmt` for formatting
- Follow effective Go practices
- Use meaningful variable and function names
- Add comments for exported functions

**Project-Specific Guidelines:**
- Keep functions focused and small
- Use descriptive test names
- Prefer explicit error handling
- Document complex algorithms

### Testing Requirements

**Test Coverage:**
- Unit tests for all new functionality
- Integration tests for complex features
- Maintain >85% code coverage

**Test Organization:**
```
pkg/analyzer/
├── analyzer.go
├── analyzer_test.go        # Unit tests
└── integration_test.go     # Integration tests

testdata/
├── samples/               # Test input files
└── expected/             # Expected outputs
```

**Test Naming:**
```go
func TestAnalyzer_DetectHTTPServer(t *testing.T) { ... }
func TestPatternMatcher_ParseURL_WithPort(t *testing.T) { ... }
```

### Architecture Guidelines

#### Adding New Languages

1. **Create Language Analyzer**
   ```go
   // pkg/analyzers/python/python.go
   type PythonAnalyzer struct {
       patterns *PythonPatterns
   }
   
   func (p *PythonAnalyzer) Analyze(path string) (*types.AnalysisResults, error) {
       // Python AST parsing implementation
   }
   ```

#### Adding New Socket Patterns (Go)

1. **Update Pattern Definitions**
   ```go
   // internal/parser/patterns/patterns.go
   pm.egressPatterns["grpc.Dial"] = EgressPattern{
       Protocol: types.ProtocolGRPC, 
       AddressArg: 0,
   }
   ```

2. **Add Test Cases**
   ```go
   // internal/parser/patterns/patterns_test.go
   {
       name: "gRPC client connection",
       code: `grpc.Dial("service.local:9090", opts...)`,
       expected: &types.SocketInfo{...},
   }
   ```

3. **Update Documentation**
   - Add example to README.md
   - Update supported patterns list

#### Adding New Resolvers

1. **Implement Resolution Logic**
   ```go
   // internal/resolver/resolver.go
   func (r *ValueResolver) resolveNewPattern(...) bool {
       // Implementation
   }
   ```

2. **Add Pattern Analysis**
   ```go
   case strings.Contains(varName, "grpc"):
       return "grpc-service", 9090, true
   ```

3. **Test Edge Cases**
   - Empty values
   - Invalid formats
   - Complex expressions

## 🧪 Testing

### Running Tests
```bash
# All tests
make test

# Specific package
go test ./pkg/analyzer

# With coverage
make coverage

# Race condition detection
go test -race ./...

# Verbose output
go test -v ./...
```

### Writing Tests

**Unit Test Example:**
```go
func TestAnalyzer_DetectHTTPServer(t *testing.T) {
    tests := []struct {
        name     string
        code     string
        expected int // number of sockets
    }{
        {
            name: "simple HTTP server",
            code: `http.ListenAndServe(":8080", nil)`,
            expected: 1,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

**Integration Test Example:**
```go
func TestAnalyzer_RealProject(t *testing.T) {
    analyzer := New()
    results, err := analyzer.Analyze("testdata/real-project")
    require.NoError(t, err)
    
    assert.True(t, results.TotalCount > 0)
    assert.True(t, results.IngressCount > 0)
}
```

### Test Data

**Sample Files:**
- Keep samples minimal and focused
- Include comments explaining patterns
- Cover edge cases and error conditions

**Expected Results:**
- Provide expected JSON output for complex samples
- Update when changing detection logic
- Include negative test cases

## 📚 Documentation

### Code Documentation

**Package Comments:**
```go
// Package analyzer provides static analysis capabilities for detecting
// socket creation patterns in Go source code.
package analyzer
```

**Function Comments:**
```go
// AnalyzeDirectory recursively analyzes all Go files in a directory
// and returns aggregated socket information.
//
// The analysis excludes vendor directories and non-Go files.
// Returns error if directory doesn't exist or contains no Go files.
func (a *Analyzer) AnalyzeDirectory(dirPath string) (*types.AnalysisResults, error) {
```

### User Documentation

**README Updates:**
- Add examples for new features
- Update supported patterns list
- Include performance characteristics

**Wiki Contributions:**
- Tutorial content
- Advanced usage examples
- Integration guides

## 🔄 Release Process

### Versioning
We follow [Semantic Versioning](https://semver.org/):
- **MAJOR**: Incompatible API changes
- **MINOR**: New functionality (backwards compatible)
- **PATCH**: Bug fixes (backwards compatible)

### Release Checklist
- [ ] All tests pass
- [ ] Documentation updated
- [ ] CHANGELOG.md updated
- [ ] Version bumped
- [ ] Tagged release created

## 🎯 Contribution Areas

### High-Priority Areas

**Socket Pattern Detection:**
- Multi-language support (Python, Java, C++, Rust)
- gRPC client/server patterns across languages
- WebSocket connections
- Message queue connections (RabbitMQ, Kafka)
- Database-specific patterns (Redis, MongoDB)

**Variable Resolution:**
- Function return value tracking
- Struct field resolution
- Interface method calls
- Cross-package constant resolution

**Output Formats:**
- Prometheus metrics format
- SARIF for security tools
- GraphQL schema for network topology
- OpenAPI specifications

**Performance Optimization:**
- Parallel file processing
- AST caching
- Memory usage optimization
- Large codebase handling

### Good First Issues

**Documentation:**
- Add more usage examples
- Create integration tutorials
- Improve error messages
- Write contributor guides

**Testing:**
- Add framework-specific test cases
- Improve edge case coverage
- Add benchmarking tests
- Create real-world samples

**Tooling:**
- GitHub Actions improvements
- Docker optimizations
- Development scripts
- IDE integration

## 💬 Community

### Getting Help

**Before asking for help:**
1. Check existing documentation
2. Search closed issues
3. Try the latest version

**Where to get help:**
- [GitHub Discussions](https://github.com/yuvalk/staticsocket/discussions)
- [Issues](https://github.com/yuvalk/staticsocket/issues) for bugs
- [Wiki](https://github.com/yuvalk/staticsocket/wiki) for documentation

### Code of Conduct

Be respectful and inclusive. We want everyone to feel welcome to contribute.

**Expected Behavior:**
- Use welcoming and inclusive language
- Respect differing viewpoints and experiences
- Accept constructive criticism gracefully
- Focus on what's best for the community

## 📄 License

By contributing to StaticSocket, you agree that your contributions will be licensed under the Apache License 2.0.

---

Thank you for contributing to StaticSocket! 🚀