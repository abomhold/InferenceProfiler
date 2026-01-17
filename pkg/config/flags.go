package config

import (
	"github.com/spf13/cobra"
)

// AddCollectionFlags adds common collection flags to a command.
func (c *Config) AddCollectionFlags(cmd *cobra.Command) {
	flags := cmd.Flags()
	flags.DurationVar(&c.Interval, "interval", c.Interval, "Collection interval")
	flags.BoolVar(&c.EnableVM, "vm", c.EnableVM, "Enable VM metrics collection")
	flags.BoolVar(&c.EnableContainer, "container", c.EnableContainer, "Enable container metrics")
	flags.BoolVar(&c.EnableProcess, "process", c.EnableProcess, "Enable process metrics")
	flags.BoolVar(&c.EnableNvidia, "nvidia", c.EnableNvidia, "Enable NVIDIA GPU metrics")
	flags.BoolVar(&c.EnableVLLM, "vllm", c.EnableVLLM, "Enable vLLM metrics")
	flags.BoolVar(&c.CollectGPUProcesses, "gpu-procs", c.CollectGPUProcesses, "Collect GPU process info")
}

// AddOutputFlags adds common output flags to a command.
func (c *Config) AddOutputFlags(cmd *cobra.Command) {
	flags := cmd.Flags()
	flags.StringVarP(&c.OutputDir, "output-dir", "o", c.OutputDir, "Output directory")
	flags.StringVarP(&c.OutputFormat, "format", "f", c.OutputFormat, "Output format (parquet, jsonl, csv, tsv)")
	flags.StringVar(&c.OutputName, "output", c.OutputName, "Output filename (auto-generated if empty)")
}

// AddGraphFlags adds graph generation flags to a command.
func (c *Config) AddGraphFlags(cmd *cobra.Command) {
	flags := cmd.Flags()
	flags.BoolVarP(&c.GenerateGraphs, "graphs", "g", c.GenerateGraphs, "Generate visualization graphs")
	flags.StringVar(&c.GraphFormat, "graph-format", c.GraphFormat, "Graph format (html, png, svg)")
	flags.StringVar(&c.GraphOutput, "graph-output", c.GraphOutput, "Graph output file (auto-generated if empty)")
}

// AddSystemFlags adds system identification flags to a command.
func (c *Config) AddSystemFlags(cmd *cobra.Command) {
	flags := cmd.Flags()
	flags.StringVar(&c.UUID, "uuid", c.UUID, "System UUID (auto-generated if empty)")
	flags.StringVar(&c.VMID, "vmid", c.VMID, "VM ID")
	flags.StringVar(&c.Hostname, "hostname", c.Hostname, "Hostname override")
}

// AddAllFlags adds all common flags to a command.
func (c *Config) AddAllFlags(cmd *cobra.Command) {
	c.AddCollectionFlags(cmd)
	c.AddOutputFlags(cmd)
	c.AddGraphFlags(cmd)
	c.AddSystemFlags(cmd)
}
