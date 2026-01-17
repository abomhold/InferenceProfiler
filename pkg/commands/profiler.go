package commands

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"InferenceProfiler/pkg/collectors"
	"InferenceProfiler/pkg/config"
	"InferenceProfiler/pkg/exporting"
	"InferenceProfiler/pkg/graphing"
)

// NewProfilerCmd creates the profiler subcommand.
func NewProfilerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Aliases: []string{"pr"},
		Use:     "profiler",
		Short:   "Continuous profiling until interrupted",
		Long: `Run continuous system profiling, collecting metrics at regular intervals
until interrupted with Ctrl+C.

Example:
  infprofiler profiler --interval 100ms --format parquet
  infprofiler profiler --nvidia=false --vllm=false`,
		RunE: runProfiler,
	}

	Cfg.AddCollectionFlags(cmd)
	Cfg.AddOutputFlags(cmd)
	Cfg.AddGraphFlags(cmd)
	Cfg.AddSystemFlags(cmd)

	return cmd
}

func runProfiler(cmd *cobra.Command, args []string) error {
	Cfg.ApplyDefaults()
	if err := Cfg.Validate(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	manager := collectors.NewManager(Cfg)
	defer manager.Close()

	static := &collectors.StaticMetrics{
		UUID:     Cfg.UUID,
		VMID:     Cfg.VMID,
		Hostname: Cfg.Hostname,
		BootTime: config.GetBootTime(),
	}
	manager.CollectStatic(static)

	outputPath := Cfg.OutputName
	if outputPath == "" {
		outputPath = Cfg.GenerateOutputPath("profiler")
	}

	exp, err := exporting.NewExporter(outputPath, Cfg.OutputFormat)
	if err != nil {
		return fmt.Errorf("failed to create exporter: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	log.Printf("Starting profiler with %v interval", Cfg.Interval)
	log.Printf("Output: %s", outputPath)

	ticker := time.NewTicker(Cfg.Interval)
	defer ticker.Stop()

	sampleCount := 0
	startTime := time.Now()

	for {
		select {
		case <-ctx.Done():
			return finalizeProfiler(exp, sampleCount, startTime)

		case <-sigChan:
			log.Println("Received interrupt signal")
			return finalizeProfiler(exp, sampleCount, startTime)

		case <-ticker.C:
			dynamic := &collectors.DynamicMetrics{}
			record := manager.CollectDynamic(dynamic)
			if err := exp.Write(record); err != nil {
				log.Printf("Error writing record: %v", err)
				continue
			}

			sampleCount++
			if sampleCount%100 == 0 {
				log.Printf("Collected %d samples", sampleCount)
			}
		}
	}
}

func finalizeProfiler(exp *exporting.Exporter, sampleCount int, startTime time.Time) error {
	log.Printf("Collection complete: %d samples in %v", sampleCount, time.Since(startTime))

	if err := exp.Close(); err != nil {
		return fmt.Errorf("failed to close exporter: %w", err)
	}

	log.Printf("Data written to: %s", exp.Path())

	if Cfg.GenerateGraphs {
		graphPath := Cfg.GenerateGraphPath(exp.Path())
		log.Printf("Generating graphs: %s", graphPath)

		gen, err := graphing.NewGenerator(exp.Path(), graphPath, Cfg.GraphFormat)
		if err != nil {
			log.Printf("Warning: failed to create graph generator: %v", err)
			return nil
		}
		if err := gen.Generate(); err != nil {
			log.Printf("Warning: failed to generate graphs: %v", err)
		}
	}

	return nil
}
