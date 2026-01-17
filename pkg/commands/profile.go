package commands

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/spf13/cobra"

	"InferenceProfiler/pkg/collectors"
	"InferenceProfiler/pkg/config"
	"InferenceProfiler/pkg/exporting"
	"InferenceProfiler/pkg/graphing"
)

// NewProfileCmd creates the profile subcommand.
func NewProfileCmd() *cobra.Command {
	cmd := &cobra.Command{
		Aliases: []string{"p"},
		Use:     "profile [flags] -- <command> [args...]",
		Short:   "Profile a command until completion",
		Long: `Run a command and profile it until completion.
Specify the command to profile after '--'.

Example:
  infprofiler profile -- python train.py --epochs 10
  infprofiler profile -g -- ./benchmark`,
		RunE: runProfile,
	}

	Cfg.AddCollectionFlags(cmd)
	Cfg.AddOutputFlags(cmd)
	Cfg.AddGraphFlags(cmd)
	Cfg.AddSystemFlags(cmd)

	return cmd
}

func runProfile(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("no command specified\nUsage: infprofiler profile [flags] -- <command> [args...]")
	}

	Cfg.ApplyDefaults()
	if err := Cfg.Validate(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	// Initialize collector manager
	manager := collectors.NewManager(Cfg)
	defer manager.Close()

	// Collect static metrics
	baseStatic := &collectors.BaseStatic{
		UUID:     Cfg.UUID,
		VMID:     Cfg.VMID,
		Hostname: Cfg.Hostname,
		BootTime: config.GetBootTime(),
	}
	manager.CollectStatic(baseStatic)

	// Initialize exporter
	outputPath := Cfg.OutputName
	if outputPath == "" {
		outputPath = Cfg.GenerateOutputPath("profile")
	}

	exp, err := exporting.NewExporter(outputPath, Cfg.OutputFormat)
	if err != nil {
		return fmt.Errorf("failed to create exporter: %w", err)
	}

	// Create target command
	targetCmd := exec.Command(args[0], args[1:]...)
	targetCmd.Stdout = os.Stdout
	targetCmd.Stderr = os.Stderr
	targetCmd.Stdin = os.Stdin

	log.Printf("Profiling command: %v", args)
	log.Printf("Output: %s", outputPath)

	// Start the target command
	if err := targetCmd.Start(); err != nil {
		exp.Close()
		return fmt.Errorf("failed to start command: %w", err)
	}

	// Start collection in background
	ctx, cancel := context.WithCancel(context.Background())
	collectorDone := make(chan struct{})

	go func() {
		defer close(collectorDone)
		runProfileCollection(ctx, manager, exp)
	}()

	// Wait for command to finish
	startTime := time.Now()
	cmdErr := targetCmd.Wait()
	duration := time.Since(startTime)

	// Stop collection
	cancel()
	<-collectorDone

	log.Printf("Command completed in %v", duration)

	// Finalize
	if err := exp.Close(); err != nil {
		log.Printf("Warning: failed to close exporter: %v", err)
	}

	log.Printf("Data written to: %s", exp.Path())

	// Generate graphs if requested
	if Cfg.GenerateGraphs {
		graphPath := Cfg.GenerateGraphPath(exp.Path())
		log.Printf("Generating graphs: %s", graphPath)

		gen, err := graphing.NewGenerator(exp.Path(), graphPath, Cfg.GraphFormat)
		if err != nil {
			log.Printf("Warning: failed to create graph generator: %v", err)
		} else if err := gen.Generate(); err != nil {
			log.Printf("Warning: failed to generate graphs: %v", err)
		}
	}

	return cmdErr
}

func runProfileCollection(ctx context.Context, manager *collectors.Manager, exp *exporting.Exporter) {
	ticker := time.NewTicker(Cfg.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			baseDynamic := &collectors.BaseDynamic{}
			record := manager.CollectDynamic(baseDynamic)
			if err := exp.Write(record); err != nil {
				log.Printf("Error writing record: %v", err)
			}
		}
	}
}
