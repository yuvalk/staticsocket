package analyzer

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"github.com/yuvalk/staticsocket/internal/parser/patterns"
	"github.com/yuvalk/staticsocket/internal/resolver"
	"github.com/yuvalk/staticsocket/pkg/types"
)

type Analyzer struct {
	fileSet  *token.FileSet
	patterns *patterns.PatternMatcher
	resolver *resolver.ValueResolver
	results  *types.AnalysisResults
}

func New() *Analyzer {
	return &Analyzer{
		fileSet:  token.NewFileSet(),
		patterns: patterns.NewPatternMatcher(),
		resolver: resolver.New(),
		results: &types.AnalysisResults{
			Sockets: make([]types.SocketInfo, 0),
		},
	}
}

func (a *Analyzer) Analyze(targetPath string) (*types.AnalysisResults, error) {
	info, err := os.Stat(targetPath)
	if err != nil {
		return nil, err
	}

	if info.IsDir() {
		return a.analyzeDirectory(targetPath)
	}
	return a.analyzeFile(targetPath)
}

func (a *Analyzer) analyzeDirectory(dirPath string) (*types.AnalysisResults, error) {
	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !strings.HasSuffix(path, ".go") || strings.Contains(path, "vendor/") {
			return nil
		}

		_, err = a.analyzeFile(path)
		return err
	})

	if err != nil {
		return nil, err
	}

	a.updateCounts()
	return a.results, nil
}

func (a *Analyzer) analyzeFile(filePath string) (*types.AnalysisResults, error) {
	src, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	file, err := parser.ParseFile(a.fileSet, filePath, src, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	visitor := &astVisitor{
		analyzer: a,
		file:     file,
		filePath: filePath,
	}

	ast.Walk(visitor, file)

	a.updateCounts()
	return a.results, nil
}

func (a *Analyzer) updateCounts() {
	a.results.TotalCount = len(a.results.Sockets)
	a.results.IngressCount = 0
	a.results.EgressCount = 0

	for i := range a.results.Sockets {
		switch a.results.Sockets[i].Type {
		case types.TrafficTypeIngress:
			a.results.IngressCount++
		case types.TrafficTypeEgress:
			a.results.EgressCount++
		}
	}
}

type astVisitor struct {
	analyzer *Analyzer
	file     *ast.File
	filePath string
}

func (v *astVisitor) Visit(node ast.Node) ast.Visitor {
	callExpr, ok := node.(*ast.CallExpr)
	if !ok {
		return v
	}

	position := v.analyzer.fileSet.Position(callExpr.Pos())

	if socket := v.analyzer.patterns.MatchSocketPattern(callExpr, v.file); socket != nil {
		socket.SourceFile = v.filePath
		socket.SourceLine = position.Line

		if socket.ProcessName == "" {
			socket.ProcessName = v.deriveProcessName()
		}

		v.analyzer.resolver.ResolveValues(socket, callExpr, v.file)
		v.analyzer.results.Sockets = append(v.analyzer.results.Sockets, *socket)
	}

	return v
}

func (v *astVisitor) deriveProcessName() string {
	packageName := v.file.Name.Name
	if packageName == "main" {
		return filepath.Base(filepath.Dir(v.filePath))
	}
	return packageName
}
