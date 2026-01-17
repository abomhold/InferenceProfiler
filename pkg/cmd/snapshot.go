package cmd

import (
	"InferenceProfiler/pkg/collecting"
	"InferenceProfiler/pkg/exporting"
	"InferenceProfiler/pkg/utils"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
)

func Snapshot(args []string) {
	fs := flag.NewFlagSet("snapshot", flag.ExitOnError)
	cfg := utils.NewConfig()
	applyFlags := utils.GetFlags(fs, cfg)

	parseArgs := args
	if len(args) > 0 && args[0] == "ss" {
		parseArgs = args[1:]
	}
	fs.Parse(parseArgs)
	applyFlags()

	manager := collecting.NewManager(cfg)
	defer manager.Close()

	record := make(map[string]interface{})

	if cfg.CollectStatic {
		static := &collecting.StaticMetrics{
			UUID:     cfg.UUID,
			Hostname: cfg.Hostname,
		}
		manager.CollectStatic(static)
		if s := manager.GetStaticRecord(); s != nil {
			record["static"] = s
		}
	}

	if cfg.CollectDynamic {
		dynamic := &collecting.DynamicMetrics{}
		dRecord := manager.CollectDynamic(dynamic)

		// Check DisableFlatten (inverted)
		if !cfg.DisableFlatten {
			dRecord = exporting.FlattenRecord(dRecord)
		}

		for k, v := range dRecord {
			record[k] = v
		}
	}

	output, err := json.MarshalIndent(record, "", "  ")
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
