// Package commands provides CLI command implementations.
package commands

import (
	"os"

	"github.com/spf13/cobra"

	"InferenceProfiler/pkg/config"
)

// Cfg is the shared configuration instance.
var Cfg = config.New()

// NewRootCmd creates the root command with all subcommands.
func NewRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "infprofiler",
		Short: "System profiler for ML inference workloads",
		Long: `InferenceProfiler collects system metrics including CPU, memory,
GPU utilization, and ML inference statistics.

Commands:
  profiler   Continuous profiling until interrupted (Ctrl+C)
  profile    Profile a command until completion
  snapshot   Capture a single metrics snapshot
  serve      Run HTTP server exposing metrics endpoints
  graph      Generate visualization graphs from data files`,
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	root.AddCommand(
		NewProfilerCmd(),
		NewProfileCmd(),
		NewSnapshotCmd(),
		NewServeCmd(),
		NewGraphCmd(),
	)

	return root
}

// Execute runs the root command.
func Execute() {
	if err := NewRootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}
