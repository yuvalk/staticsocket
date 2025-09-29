package resolver

import (
	"go/ast"
	"strconv"

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

	// Try to resolve variables and constants
	// This is a simplified implementation - a full implementation would use go/types
	for _, arg := range callExpr.Args {
		if ident, ok := arg.(*ast.Ident); ok {
			if value := r.resolveIdentifier(ident, file); value != "" {
				r.updateSocketWithResolvedValue(socket, value)
				return
			}
		}
	}
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
	// Simplified egress parsing
	// In practice, you'd use proper URL parsing and host:port parsing
}