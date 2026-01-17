package commands

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"InferenceProfiler/pkg/collectors"
	"InferenceProfiler/pkg/config"
	"InferenceProfiler/pkg/formatting"
)

var (
	snapshotStatic  bool
	snapshotDynamic bool
	snapshotOutput  string
)

// NewSnapshotCmd creates the snapshot subcommand.
func NewSnapshotCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "snapshot",
		Aliases: []string{"ss"},
		Short:   "Capture a single metrics snapshot",
		Long: `Capture a single snapshot of system metrics and output as JSON.

Modes:
  --static   Only static metrics (system info, GPU info)
  --dynamic  Only dynamic metrics (CPU, memory, utilization)
  (default)  Both static and dynamic metrics

Example:
  infprofiler snapshot --static
  infprofiler snapshot --dynamic -o metrics.json
  infprofiler snapshot`,
		RunE: runSnapshot,
	}

	Cfg.AddCollectionFlags(cmd)
	Cfg.AddSystemFlags(cmd)

	cmd.Flags().BoolVar(&snapshotStatic, "static", false, "Collect only static metrics")
	cmd.Flags().BoolVar(&snapshotDynamic, "dynamic", false, "Collect only dynamic metrics")
	cmd.Flags().StringVarP(&snapshotOutput, "output", "o", "", "Output file (default: stdout)")

	return cmd
}

func runSnapshot(cmd *cobra.Command, args []string) error {
	Cfg.ApplyDefaults()

	manager := collectors.NewManager(Cfg)
	defer manager.Close()

	result := make(formatting.Record)

	collectStatic := !snapshotDynamic || snapshotStatic
	collectDynamic := !snapshotStatic || snapshotDynamic

	if !snapshotStatic && !snapshotDynamic {
		collectStatic = true
		collectDynamic = true
	}

	if collectStatic {
		static := &collectors.StaticMetrics{
			UUID:     Cfg.UUID,
			VMID:     Cfg.VMID,
			Hostname: Cfg.Hostname,
			BootTime: config.GetBootTime(),
		}
		manager.CollectStatic(static)

		for k, v := range manager.GetStaticRecord() {
			result[k] = v
		}
	}

	if collectDynamic {
		dynamic := &collectors.DynamicMetrics{}
		dynamicRecord := manager.CollectDynamic(dynamic)
		dynamicRecord = formatting.FlattenRecord(dynamicRecord)
		for k, v := range dynamicRecord {
			result[k] = v
		}
	}

	return writeSnapshotOutput(result)
}

func writeSnapshotOutput(data formatting.Record) error {
	output, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	if snapshotOutput == "" {
		fmt.Println(string(output))
		return nil
	}

	if err := os.WriteFile(snapshotOutput, output, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	fmt.Fprintf(os.Stderr, "Written to: %s\n", snapshotOutput)
	return nil
}
