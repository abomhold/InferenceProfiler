package commands

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"InferenceProfiler/pkg/collecting"
	"InferenceProfiler/pkg/config"
	"InferenceProfiler/pkg/metrics"
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

	// Initialize collector manager
	manager := collecting.NewManager(Cfg)
	defer manager.Close()

	result := make(metrics.Record)

	// Determine what to collect
	collectStatic := !snapshotDynamic || snapshotStatic
	collectDynamic := !snapshotStatic || snapshotDynamic

	// If neither flag set, collect both
	if !snapshotStatic && !snapshotDynamic {
		collectStatic = true
		collectDynamic = true
	}

	if collectStatic {
		baseStatic := &metrics.BaseStatic{
			UUID:     Cfg.UUID,
			VMID:     Cfg.VMID,
			Hostname: Cfg.Hostname,
			BootTime: config.GetBootTime(),
		}
		manager.CollectStatic(baseStatic)

		result["uuid"] = baseStatic.UUID
		result["vId"] = baseStatic.VMID
		result["vHostname"] = baseStatic.Hostname
		result["vBootTime"] = baseStatic.BootTime
	}

	if collectDynamic {
		baseDynamic := &metrics.BaseDynamic{}
		dynamicMetrics := manager.CollectDynamic(baseDynamic)
		for k, v := range dynamicMetrics {
			result[k] = v
		}
	}

	return writeSnapshotOutput(result)
}

func writeSnapshotOutput(data metrics.Record) error {
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
