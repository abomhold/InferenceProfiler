package cmd

import (
	"InferenceProfiler/pkg/exporting"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func Graph(args []string) {
	fs := flag.NewFlagSet("graph", flag.ExitOnError)
	var outputDir string
	fs.StringVar(&outputDir, "output", "", "Output directory for graphs")
	fs.Parse(args)

	if fs.NArg() < 1 {
		log.Fatal("Input file required. Usage: infpro graph [flags] <input-file>")
	}

	inputFile := fs.Arg(0)
	if _, err := os.Stat(inputFile); err != nil {
		log.Fatalf("Input file not found: %s", inputFile)
	}

	if outputDir == "" {
		base := filepath.Base(inputFile)
		ext := filepath.Ext(base)
		name := strings.TrimSuffix(base, ext)
		outputDir = name + "_graphs"
	}

	log.Printf("Generating graphs from %s", inputFile)
	log.Printf("Output directory: %s", outputDir)

	if err := generateGraphs(inputFile, outputDir); err != nil {
		log.Fatalf("Failed to generate graphs: %v", err)
	}

	log.Printf("Successfully generated graphs in %s", outputDir)
}

func generateGraphs(inputPath, outputDir string) error {
	if outputDir == "" {
		base := filepath.Base(inputPath)
		ext := filepath.Ext(base)
		name := strings.TrimSuffix(base, ext)
		outputDir = name + "_graphs"
	}

	gen, err := exporting.NewGenerator(inputPath, outputDir, "png")
	if err != nil {
		return fmt.Errorf("failed to create generator: %w", err)
	}

	return gen.Generate()
}
