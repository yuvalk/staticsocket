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
		fmt.Fprintf(os.Stderr, "Error analyzing path %s: %v\n", *targetPath, err)
		os.Exit(1)
	}

	output := os.Stdout
	if *outputFile != "" {
		file, err := os.Create(*outputFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating output file: %v\n", err)
			os.Exit(1)
		}
		defer file.Close()
		output = file
	}

	if err := results.Export(output, *format); err != nil {
		fmt.Fprintf(os.Stderr, "Error exporting results: %v\n", err)
		os.Exit(1)
	}
}