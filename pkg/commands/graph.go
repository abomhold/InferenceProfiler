package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"InferenceProfiler/pkg/graphing"
)

var (
	graphOutput string
)

// NewGraphCmd creates the graph subcommand.
func NewGraphCmd() *cobra.Command {
	cmd := &cobra.Command{
		Aliases: []string{"g"},
		Use:     "graph <input-file>",
		Short:   "Generate visualization graphs from data",
		Long: `Generate PNG visualization graphs from profiler data files.

Output is a directory containing PNG files for each metric.

Supported input formats: parquet, jsonl, csv, tsv

Example:
  infprofiler graph profile-20240101-120000.parquet
  infprofiler graph data.jsonl -o ./graphs`,
		Args: cobra.ExactArgs(1),
		RunE: runGraph,
	}

	cmd.Flags().StringVarP(&graphOutput, "output", "o", "", "Output directory (auto-generated if empty)")

	return cmd
}

func runGraph(cmd *cobra.Command, args []string) error {
	inputPath := args[0]
	if _, err := os.Stat(inputPath); err != nil {
		return fmt.Errorf("input file not found: %s", inputPath)
	}
	outputDir := graphOutput
	if outputDir == "" {
		outputDir = Cfg.GenerateGraphPath(inputPath)
	}

	gen, err := graphing.NewGenerator(inputPath, outputDir, "png")
	if err != nil {
		return fmt.Errorf("failed to create generator: %w", err)
	}
	if err := gen.Generate(); err != nil {
		return fmt.Errorf("failed to generate graphs: %w", err)
	}
	fmt.Printf("Generated graphs in: %s\n", outputDir)
	return nil
}
