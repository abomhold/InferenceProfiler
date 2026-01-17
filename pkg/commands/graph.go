package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"InferenceProfiler/pkg/graphing"
)

var (
	graphFormat string
	graphOutput string
)

// NewGraphCmd creates the graph subcommand.
func NewGraphCmd() *cobra.Command {
	cmd := &cobra.Command{
		Aliases: []string{"g"},
		Use:     "graph <input-file>",
		Short:   "Generate visualization graphs from data",
		Long: `Generate HTML visualization graphs from profiler data files.

Supported input formats: parquet, jsonl, csv, tsv

Example:
  infprofiler graph profile-20240101-120000.parquet
  infprofiler graph data.jsonl -o report.html`,
		Args: cobra.ExactArgs(1),
		RunE: runGraph,
	}

	cmd.Flags().StringVar(&graphFormat, "format", "html", "Output format (html)")
	cmd.Flags().StringVarP(&graphOutput, "output", "o", "", "Output file (auto-generated if empty)")

	return cmd
}

func runGraph(cmd *cobra.Command, args []string) error {
	inputPath := args[0]

	// Validate input exists
	if _, err := os.Stat(inputPath); err != nil {
		return fmt.Errorf("input file not found: %s", inputPath)
	}

	// Generate output path if not specified
	outputPath := graphOutput
	if outputPath == "" {
		Cfg.GraphFormat = graphFormat
		outputPath = Cfg.GenerateGraphPath(inputPath)
	}

	// Generate graphs
	gen, err := graphing.NewGenerator(inputPath, outputPath, graphFormat)
	if err != nil {
		return fmt.Errorf("failed to create generator: %w", err)
	}

	if err := gen.Generate(); err != nil {
		return fmt.Errorf("failed to generate graphs: %w", err)
	}

	fmt.Printf("Generated: %s\n", outputPath)
	return nil
}
