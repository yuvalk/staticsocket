package resolver

import (
	"go/ast"
	"strconv"
	"strings"

	socketTypes "github.com/yuvalk/staticsocket/pkg/types"
)

type ValueResolver struct {
	// Future: add support for type checking and constant resolution
}

func New() *ValueResolver {
	return &ValueResolver{}
}

func (r *ValueResolver) ResolveValues(socket *socketTypes.SocketInfo, callExpr *ast.CallExpr, file *ast.File) {
	// If already resolved from string literals, no need to do more
	if socket.IsResolved {
		return
	}

	// Get the URL/address argument based on the pattern
	var urlArg ast.Expr
	if socket.PatternMatch == "http.Get" || socket.PatternMatch == "http.Post" || socket.PatternMatch == "http.PostForm" {
		if len(callExpr.Args) > 0 {
			urlArg = callExpr.Args[0]
		}
	} else {
		// For net.Dial patterns, get the address argument (usually index 1)
		if len(callExpr.Args) > 1 {
			urlArg = callExpr.Args[1]
		}
	}

	if urlArg == nil {
		return
	}

	// Try different resolution strategies
	if r.tryResolveArgument(socket, urlArg, file) {
		return
	}
}

func (r *ValueResolver) tryResolveArgument(socket *socketTypes.SocketInfo, arg ast.Expr, file *ast.File) bool {
	switch expr := arg.(type) {
	case *ast.Ident:
		// Simple identifier (variable or constant)
		if value := r.resolveIdentifier(expr, file); value != "" {
			r.updateSocketWithResolvedValue(socket, value)
			return true
		}
		
		// Check for common patterns like httptest server
		if host, port, resolved := r.analyzeVariablePattern(expr.Name); resolved {
			socket.IsResolved = true
			socket.DestinationHost = &host
			if port > 0 {
				socket.DestinationPort = &port
			}
			socket.RawValue = expr.Name
			return true
		}
		
	case *ast.SelectorExpr:
		// Field access like server.URL, os.Getenv(), etc.
		varName := r.extractSelectorName(expr)
		if host, port, resolved := r.analyzeVariablePattern(varName); resolved {
			socket.IsResolved = true
			socket.DestinationHost = &host
			if port > 0 {
				socket.DestinationPort = &port
			}
			socket.RawValue = varName
			return true
		}
		
	case *ast.BinaryExpr:
		// String concatenation like baseURL + endpoint
		if r.tryResolveBinaryExpr(socket, expr, file) {
			return true
		}
		
	case *ast.CallExpr:
		// Function calls like url.Parse().String(), getServiceURL()
		if r.tryResolveCallExpr(socket, expr, file) {
			return true
		}
	}
	
	return false
}

func (r *ValueResolver) resolveIdentifier(ident *ast.Ident, file *ast.File) string {
	// Look for constant declarations in the file
	for _, decl := range file.Decls {
		if genDecl, ok := decl.(*ast.GenDecl); ok {
			for _, spec := range genDecl.Specs {
				if valueSpec, ok := spec.(*ast.ValueSpec); ok {
					for i, name := range valueSpec.Names {
						if name.Name == ident.Name && i < len(valueSpec.Values) {
							if lit, ok := valueSpec.Values[i].(*ast.BasicLit); ok {
								if lit.Kind.String() == "STRING" {
									if value, err := strconv.Unquote(lit.Value); err == nil {
										return value
									}
								}
							}
						}
					}
				}
			}
		}
	}
	return ""
}

func (r *ValueResolver) updateSocketWithResolvedValue(socket *socketTypes.SocketInfo, value string) {
	socket.RawValue = value
	socket.IsResolved = true

	switch socket.Type {
	case socketTypes.TrafficTypeIngress:
		r.parseIngressValue(socket, value)
	case socketTypes.TrafficTypeEgress:
		r.parseEgressValue(socket, value)
	}
}

func (r *ValueResolver) parseIngressValue(socket *socketTypes.SocketInfo, value string) {
	// Reuse parsing logic from patterns package
	// This is simplified - in practice, you'd factor out the parsing logic
	if value != "" && value[0] == ':' {
		if port, err := strconv.Atoi(value[1:]); err == nil {
			socket.ListenPort = &port
			socket.ListenInterface = "0.0.0.0"
		}
	}
}

func (r *ValueResolver) parseEgressValue(socket *socketTypes.SocketInfo, value string) {
	// Parse egress addresses (host:port format)
	if strings.Contains(value, "://") {
		// This looks like a URL, but we only handle simple host:port here
		// URL parsing is handled by the patterns package
		return
	}
	
	// Parse simple host:port format
	parts := strings.Split(value, ":")
	if len(parts) == 2 {
		host := parts[0]
		socket.DestinationHost = &host
		
		if port, err := strconv.Atoi(parts[1]); err == nil {
			socket.DestinationPort = &port
		}
	}
}

func (r *ValueResolver) analyzeVariablePattern(varName string) (host string, port int, resolved bool) {
	// Common patterns we can make educated guesses about
	switch {
	case strings.Contains(varName, "server.URL") || strings.Contains(varName, "httptest"):
		// httptest.NewServer() typically binds to localhost with random port
		return "localhost", 0, true
		
	case strings.Contains(varName, "localhost") || strings.Contains(varName, "127.0.0.1"):
		// Variables with localhost in name likely target localhost
		return "localhost", 0, true
		
	case strings.Contains(varName, "URL") && (strings.Contains(varName, "api") || strings.Contains(varName, "service")):
		// API/service URLs - we can mark as external but don't know specifics
		return "external-service", 0, true
		
	default:
		return "", 0, false
	}
}

func (r *ValueResolver) extractSelectorName(expr *ast.SelectorExpr) string {
	// Extract the full selector expression as a string
	var parts []string
	
	// Walk the selector chain
	current := expr
	for current != nil {
		parts = append([]string{current.Sel.Name}, parts...)
		
		if ident, ok := current.X.(*ast.Ident); ok {
			parts = append([]string{ident.Name}, parts...)
			break
		} else if sel, ok := current.X.(*ast.SelectorExpr); ok {
			current = sel
		} else {
			break
		}
	}
	
	return strings.Join(parts, ".")
}

func (r *ValueResolver) tryResolveBinaryExpr(socket *socketTypes.SocketInfo, expr *ast.BinaryExpr, file *ast.File) bool {
	// Handle string concatenation like baseURL + endpoint
	if expr.Op.String() == "+" {
		// Try to resolve the left side (usually the base URL)
		if ident, ok := expr.X.(*ast.Ident); ok {
			if baseValue := r.resolveIdentifier(ident, file); baseValue != "" {
				// Mark as partially resolved with the base URL
				socket.IsResolved = true
				socket.RawValue = baseValue + " + ..."
				
				// Try to parse the base URL
				if strings.Contains(baseValue, "://") {
					// Parse as URL
					r.parseURLForSocket(socket, baseValue)
				} else {
					// Parse as host:port
					r.parseEgressValue(socket, baseValue)
				}
				return true
			}
		}
	}
	return false
}

func (r *ValueResolver) tryResolveCallExpr(socket *socketTypes.SocketInfo, expr *ast.CallExpr, file *ast.File) bool {
	// Handle function calls that return URLs
	if sel, ok := expr.Fun.(*ast.SelectorExpr); ok {
		funcName := r.extractSelectorName(sel)
		
		// Common patterns
		switch {
		case strings.Contains(funcName, "String") && strings.Contains(funcName, "URL"):
			// url.Parse().String() pattern
			socket.IsResolved = true
			socket.RawValue = "parsed-url"
			socket.DestinationHost = stringPtr("parsed-url-host")
			return true
			
		case strings.Contains(funcName, "getURL") || strings.Contains(funcName, "GetURL"):
			// Functions that return URLs
			socket.IsResolved = true
			socket.RawValue = funcName + "()"
			socket.DestinationHost = stringPtr("dynamic-url")
			return true
		}
	}
	
	return false
}

func (r *ValueResolver) parseURLForSocket(socket *socketTypes.SocketInfo, url string) {
	// Simple URL parsing to extract host/port
	if strings.HasPrefix(url, "https://") {
		socket.Protocol = socketTypes.ProtocolHTTPS
		url = url[8:]
		port := 443
		socket.DestinationPort = &port
	} else if strings.HasPrefix(url, "http://") {
		socket.Protocol = socketTypes.ProtocolHTTP
		url = url[7:]
		port := 80
		socket.DestinationPort = &port
	}
	
	// Extract host
	parts := strings.Split(url, "/")
	if len(parts) > 0 && parts[0] != "" {
		hostPort := parts[0]
		if strings.Contains(hostPort, ":") {
			hostPortParts := strings.Split(hostPort, ":")
			if len(hostPortParts) >= 2 {
				host := hostPortParts[0]
				socket.DestinationHost = &host
				if port, err := strconv.Atoi(hostPortParts[1]); err == nil {
					socket.DestinationPort = &port
				}
			}
		} else {
			socket.DestinationHost = &hostPort
		}
	}
}

func stringPtr(s string) *string {
	return &s
}