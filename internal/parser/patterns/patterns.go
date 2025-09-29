package patterns

import (
	"go/ast"
	"strconv"
	"strings"

	"github.com/yuvalk/staticsocket/pkg/types"
)

const (
	dialAddressArg     = 2
	hostPortPartsCount = 2
)

type PatternMatcher struct {
	ingressPatterns map[string]IngressPattern
	egressPatterns  map[string]EgressPattern
}

type IngressPattern struct {
	Protocol   types.Protocol
	AddressArg int  // argument index for address
	PortOnly   bool // true if address is just port (e.g., ":8080")
}

type EgressPattern struct {
	Protocol   types.Protocol
	AddressArg int // argument index for address
	URLArg     int // argument index for URL (for HTTP patterns)
}

func NewPatternMatcher() *PatternMatcher {
	pm := &PatternMatcher{
		ingressPatterns: make(map[string]IngressPattern),
		egressPatterns:  make(map[string]EgressPattern),
	}
	pm.initializePatterns()
	return pm
}

func (pm *PatternMatcher) initializePatterns() {
	// Ingress patterns (listeners)
	pm.ingressPatterns["net.Listen"] = IngressPattern{Protocol: types.ProtocolTCP, AddressArg: 1}
	pm.ingressPatterns["net.ListenTCP"] = IngressPattern{Protocol: types.ProtocolTCP, AddressArg: 1}
	pm.ingressPatterns["net.ListenUDP"] = IngressPattern{Protocol: types.ProtocolUDP, AddressArg: 1}
	pm.ingressPatterns["net.ListenUnix"] = IngressPattern{Protocol: types.ProtocolUnix, AddressArg: 1}
	pm.ingressPatterns["http.ListenAndServe"] = IngressPattern{Protocol: types.ProtocolHTTP, AddressArg: 0, PortOnly: true}
	pm.ingressPatterns["http.ListenAndServeTLS"] = IngressPattern{
		Protocol: types.ProtocolHTTPS, AddressArg: 0, PortOnly: true,
	}

	// Egress patterns (outbound connections)
	pm.egressPatterns["net.Dial"] = EgressPattern{Protocol: types.ProtocolTCP, AddressArg: 1}
	pm.egressPatterns["net.DialTCP"] = EgressPattern{Protocol: types.ProtocolTCP, AddressArg: dialAddressArg}
	pm.egressPatterns["net.DialUDP"] = EgressPattern{Protocol: types.ProtocolUDP, AddressArg: dialAddressArg}
	pm.egressPatterns["net.DialTimeout"] = EgressPattern{Protocol: types.ProtocolTCP, AddressArg: 1}
	pm.egressPatterns["http.Get"] = EgressPattern{Protocol: types.ProtocolHTTP, URLArg: 0}
	pm.egressPatterns["http.Post"] = EgressPattern{Protocol: types.ProtocolHTTP, URLArg: 0}
	pm.egressPatterns["http.PostForm"] = EgressPattern{Protocol: types.ProtocolHTTP, URLArg: 0}
}

func (pm *PatternMatcher) MatchSocketPattern(callExpr *ast.CallExpr, file *ast.File) *types.SocketInfo {
	funcName := pm.extractFunctionName(callExpr)
	if funcName == "" {
		return nil
	}

	// Check for ingress patterns
	if pattern, exists := pm.ingressPatterns[funcName]; exists {
		return pm.matchIngressPattern(callExpr, pattern, funcName)
	}

	// Check for egress patterns
	if pattern, exists := pm.egressPatterns[funcName]; exists {
		return pm.matchEgressPattern(callExpr, pattern, funcName)
	}

	return nil
}

func (pm *PatternMatcher) matchIngressPattern(
	callExpr *ast.CallExpr,
	pattern IngressPattern,
	funcName string,
) *types.SocketInfo {
	if len(callExpr.Args) <= pattern.AddressArg {
		return nil
	}

	addressArg := callExpr.Args[pattern.AddressArg]
	rawValue := pm.extractStringLiteral(addressArg)

	socket := &types.SocketInfo{
		Type:         types.TrafficTypeIngress,
		Protocol:     pattern.Protocol,
		RawValue:     rawValue,
		PatternMatch: funcName,
		FunctionName: pm.extractContainingFunction(callExpr),
	}

	if rawValue != "" {
		pm.parseIngressAddress(socket, rawValue, pattern.PortOnly)
	}

	return socket
}

func (pm *PatternMatcher) matchEgressPattern(
	callExpr *ast.CallExpr,
	pattern EgressPattern,
	funcName string,
) *types.SocketInfo {
	var rawValue string
	var argIndex int
	var isURL bool

	// Check if this pattern uses URLArg (for HTTP methods)
	if pattern.URLArg >= 0 && (funcName == "http.Get" || funcName == "http.Post" || funcName == "http.PostForm") {
		argIndex = pattern.URLArg
		isURL = true
	} else {
		argIndex = pattern.AddressArg
		isURL = false
	}

	if len(callExpr.Args) <= argIndex {
		return nil
	}

	arg := callExpr.Args[argIndex]
	rawValue = pm.extractStringLiteral(arg)

	socket := &types.SocketInfo{
		Type:         types.TrafficTypeEgress,
		Protocol:     pattern.Protocol,
		RawValue:     rawValue,
		PatternMatch: funcName,
		FunctionName: pm.extractContainingFunction(callExpr),
	}

	if rawValue != "" {
		if isURL {
			pm.parseEgressURL(socket, rawValue)
		} else {
			pm.parseEgressAddress(socket, rawValue)
		}
	}

	return socket
}

func (pm *PatternMatcher) extractFunctionName(callExpr *ast.CallExpr) string {
	switch fun := callExpr.Fun.(type) {
	case *ast.SelectorExpr:
		if ident, ok := fun.X.(*ast.Ident); ok {
			return ident.Name + "." + fun.Sel.Name
		}
	case *ast.Ident:
		return fun.Name
	}
	return ""
}

func (pm *PatternMatcher) extractStringLiteral(expr ast.Expr) string {
	if lit, ok := expr.(*ast.BasicLit); ok && lit.Kind.String() == "STRING" {
		value, err := strconv.Unquote(lit.Value)
		if err == nil {
			return value
		}
	}
	return ""
}

func (pm *PatternMatcher) extractContainingFunction(callExpr *ast.CallExpr) string {
	// This is a simplified implementation
	// In a real implementation, you'd walk up the AST to find the containing function
	return "unknown"
}

func (pm *PatternMatcher) parseIngressAddress(socket *types.SocketInfo, address string, portOnly bool) {
	socket.IsResolved = true

	if portOnly && strings.HasPrefix(address, ":") {
		// Format like ":8080"
		if port, err := strconv.Atoi(address[1:]); err == nil {
			socket.ListenPort = &port
			socket.ListenInterface = "0.0.0.0"
		}
		return
	}

	// Parse host:port format
	parts := strings.Split(address, ":")
	if len(parts) == hostPortPartsCount {
		host := parts[0]
		if host == "" {
			host = "0.0.0.0"
		}
		socket.ListenInterface = host

		if port, err := strconv.Atoi(parts[1]); err == nil {
			socket.ListenPort = &port
		}
	}
}

func (pm *PatternMatcher) parseEgressAddress(socket *types.SocketInfo, address string) {
	socket.IsResolved = true

	parts := strings.Split(address, ":")
	if len(parts) == hostPortPartsCount {
		host := parts[0]
		socket.DestinationHost = &host

		if port, err := strconv.Atoi(parts[1]); err == nil {
			socket.DestinationPort = &port
		}
	}
}

func (pm *PatternMatcher) parseEgressURL(socket *types.SocketInfo, url string) {
	socket.IsResolved = true

	// Parse URL to extract scheme, host, and port
	var remainingURL string
	var defaultPort int

	if strings.HasPrefix(url, "https://") {
		socket.Protocol = types.ProtocolHTTPS
		remainingURL = url[8:]
		defaultPort = 443
	} else if strings.HasPrefix(url, "http://") {
		socket.Protocol = types.ProtocolHTTP
		remainingURL = url[7:]
		defaultPort = 80
	} else {
		// No scheme prefix, treat as raw URL
		remainingURL = url
		defaultPort = 80
	}

	// Extract host and port from URL (everything before the first slash)
	parts := strings.Split(remainingURL, "/")
	if len(parts) > 0 && parts[0] != "" {
		hostPort := parts[0]
		if strings.Contains(hostPort, ":") {
			// Host includes explicit port
			hostPortParts := strings.Split(hostPort, ":")
			if len(hostPortParts) >= hostPortPartsCount {
				host := hostPortParts[0]
				socket.DestinationHost = &host
				if port, err := strconv.Atoi(hostPortParts[1]); err == nil {
					socket.DestinationPort = &port
				}
			}
		} else {
			// Host without explicit port, use default
			socket.DestinationHost = &hostPort
			socket.DestinationPort = &defaultPort
		}
	}
}
