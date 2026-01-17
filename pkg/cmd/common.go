package cmd

import (
	"InferenceProfiler/pkg/utils"
	"flag"
)

type CommonFlags struct {
	// Collection
	Concurrent      bool
	EnableVM        bool
	EnableContainer bool
	EnableProcess   bool
	EnableNvidia    bool
	EnableVLLM      bool
	CollectGPUProcs bool

	// Output
	Format     string
	OutputFile string
	Flatten    bool

	// Mode
	Batch  bool
	Stream bool

	// Profiling
	Interval int

	// Cleanup
	Cleanup bool

	// Graphs
	Graphs   bool
	GraphDir string

	// Server
	Port int
}

func addCollectionFlags(fs *flag.FlagSet, f *CommonFlags) {
	fs.BoolVar(&f.Concurrent, "concurrent", false, "Enable concurrent collection")
	fs.BoolVar(&f.EnableVM, "vm", true, "Collect VM metrics")
	fs.BoolVar(&f.EnableContainer, "container", false, "Collect container metrics")
	fs.BoolVar(&f.EnableProcess, "process", false, "Collect process metrics")
	fs.BoolVar(&f.EnableNvidia, "nvidia", false, "Collect NVIDIA GPU metrics")
	fs.BoolVar(&f.EnableVLLM, "vllm", false, "Collect vLLM metrics")
	fs.BoolVar(&f.CollectGPUProcs, "gpu-procs", false, "Collect GPU process info")
}

func addOutputFlags(fs *flag.FlagSet, f *CommonFlags, defaultFormat string) {
	fs.StringVar(&f.Format, "format", defaultFormat, "Output format (jsonl, parquet, csv, tsv)")
	fs.StringVar(&f.OutputFile, "output", "", "Output file path")
	fs.BoolVar(&f.Flatten, "flatten", false, "Flatten nested structures")
}

func addModeFlags(fs *flag.FlagSet, f *CommonFlags) {
	fs.BoolVar(&f.Batch, "batch", false, "Batch mode: collect in memory, write at end")
	fs.BoolVar(&f.Stream, "stream", false, "Stream mode: write directly to file")
}

func addProfilingFlags(fs *flag.FlagSet, f *CommonFlags) {
	fs.IntVar(&f.Interval, "interval", 100, "Collection interval in milliseconds")
	fs.BoolVar(&f.Cleanup, "cleanup", false, "Remove intermediate JSON files")
}

func addGraphFlags(fs *flag.FlagSet, f *CommonFlags) {
	fs.BoolVar(&f.Graphs, "graphs", false, "Generate graphs after collection")
	fs.StringVar(&f.GraphDir, "graph-dir", "", "Graph output directory")
}

func addServerFlags(fs *flag.FlagSet, f *CommonFlags) {
	fs.IntVar(&f.Port, "port", 8080, "HTTP server port")
}

func (f *CommonFlags) ToConfig() *utils.Config {
	cfg := utils.NewConfig()
	cfg.Concurrent = f.Concurrent
	cfg.EnableVM = f.EnableVM
	cfg.EnableContainer = f.EnableContainer
	cfg.EnableProcess = f.EnableProcess
	cfg.EnableNvidia = f.EnableNvidia
	cfg.EnableVLLM = f.EnableVLLM
	cfg.CollectGPUProcesses = f.CollectGPUProcs
	cfg.Format = f.Format
	cfg.OutputFile = f.OutputFile
	cfg.Flatten = f.Flatten
	cfg.Interval = f.Interval
	cfg.Port = f.Port
	return cfg
}
