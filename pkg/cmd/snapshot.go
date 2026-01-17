package cmd

import (
	"InferenceProfiler/pkg/collecting"
	"InferenceProfiler/pkg/exporting"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
)

var snapshotCmd = &cobra.Command{
	Use:     "snapshot",
	Aliases: []string{"ss", "snap"},
	Short:   "Collect a single snapshot of metrics",
	Run:     runSnapshot,
}

func init() {
	rootCmd.AddCommand(snapshotCmd)

	// Snapshot-specific flags
	snapshotCmd.Flags().StringVarP(&cfg.OutputFormat, "format", "f", "json", "Output format (json, csv, tsv)")
	snapshotCmd.Flags().StringVarP(&cfg.OutputFile, "output", "o", "", "Output file (default: stdout)")
	snapshotCmd.Flags().BoolVar(&cfg.Flatten, "flatten", false, "Flatten nested structures")
}

func runSnapshot(cmd *cobra.Command, args []string) {
	applyNegativeFlags(cmd)

	// Initialize manager
	manager := collecting.NewManager(cfg)
	defer manager.Close()

	// Collect static metrics
	static := &collecting.StaticMetrics{
		UUID:     cfg.UUID,
		Hostname: cfg.Hostname,
	}
	manager.CollectStatic(static)

	// Collect dynamic metrics
	dynamic := &collecting.DynamicMetrics{}
	record := manager.CollectDynamic(dynamic)

	if cfg.Flatten {
		record = exporting.FlattenRecord(record)
	}

	// Output
	var output []byte
	var err error

	switch cfg.OutputFormat {
	case "json":
		output, err = json.MarshalIndent(record, "", "  ")
	case "csv":
		output, err = exporting.RecordToCSV(record)
	case "tsv":
		output, err = exporting.RecordToTSV(record)
	default:
		log.Fatalf("Unknown format: %s", cfg.OutputFormat)
	}

	if err != nil {
		log.Fatalf("Failed to marshal output: %v", err)
	}

	if cfg.OutputFile != "" {
		if err := os.WriteFile(cfg.OutputFile, output, 0644); err != nil {
			log.Fatalf("Failed to write file: %v", err)
		}
		log.Printf("Wrote snapshot to %s", cfg.OutputFile)
	} else {
		fmt.Println(string(output))
	}
}
