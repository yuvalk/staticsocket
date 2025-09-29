package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/yuvalk/staticsocket/pkg/analyzer"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	var (
		targetPath = flag.String("path", ".", "Path to analyze (file or directory)")
		outputFile = flag.String("output", "", "Output file (default: stdout)")
		format     = flag.String("format", "json", "Output format: json, yaml, csv")
		verbose    = flag.Bool("verbose", false, "Enable verbose output")
	)
	flag.Parse()

	if *verbose {
		log.SetOutput(os.Stderr)
	} else {
		log.SetOutput(io.Discard)
	}

	analyzer := analyzer.New()
	results, err := analyzer.Analyze(*targetPath)
	if err != nil {
		return fmt.Errorf("analyzing path %s: %w", *targetPath, err)
	}

	output := os.Stdout
	if *outputFile != "" {
		file, err := os.Create(*outputFile)
		if err != nil {
			return fmt.Errorf("creating output file: %w", err)
		}
		defer file.Close()
		output = file
	}

	if err := results.Export(output, *format); err != nil {
		return fmt.Errorf("exporting results: %w", err)
	}

	return nil
}
