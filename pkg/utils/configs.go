package utils

import (
	"crypto/rand"
	"flag"
	"fmt"
	"os"
)

const (
	CMDSeparator        = "--"
	VLLMEnvVar          = "VLLM_METRICS_ENDPOINT"
	DefaultVLLMEndpoint = "http://127.0.0.1:8000/metrics"
)

type Config struct {
	DisableVM           bool
	DisableContainer    bool
	DisableProcess      bool
	DisableNvidia       bool
	DisableVLLM         bool
	DisableGPUProcesses bool
	Concurrent          bool
	CollectStatic       bool
	CollectDynamic      bool
	Format              string
	OutputFile          string
	NoJsonString        bool
	Interval            int
	Duration            int
	Delta               bool
	Batch               bool
	Stream              bool
	Cleanup             bool
	Graphs              bool
	GraphDir            string
	Port                int
	UUID                string
	Hostname            string
}

func NewConfig() *Config {
	return &Config{
		Interval: 1000,
		Port:     8080,
		Format:   "json",
		UUID:     GenerateUUID(),
		Hostname: GetHostname(),
	}
}

func GetFlags(fs *flag.FlagSet, cfg *Config) func() {
	fs.BoolVar(&cfg.Concurrent, "concurrent", false, "Enable concurrent collection")
	fs.BoolVar(&cfg.DisableVM, "no-vm", false, "Collect VM metrics")
	fs.BoolVar(&cfg.DisableContainer, "no-container", false, "Disable container metrics")
	fs.BoolVar(&cfg.DisableProcess, "no-procs", false, "Disable process metrics")
	fs.BoolVar(&cfg.DisableNvidia, "no-nvidia", false, "Disable NVIDIA GPU metrics")
	fs.BoolVar(&cfg.DisableVLLM, "no-vllm", false, "Disable vLLM metrics")
	fs.BoolVar(&cfg.DisableGPUProcesses, "no-gpu-procs", false, "Disable GPU process info")
	fs.BoolVar(&cfg.NoJsonString, "no-json-string", false, "Expand all arrays to top-level fields (procs, disks, static GPU info)")
	fs.StringVar(&cfg.Format, "format", cfg.Format, "Output format")
	fs.StringVar(&cfg.OutputFile, "output", "", "Output file path")
	fs.IntVar(&cfg.Interval, "interval", cfg.Interval, "Collection interval (ms)")
	fs.IntVar(&cfg.Duration, "duration", 0, "Duration in ms for delta mode (0 = continuous or wait for command)")
	fs.BoolVar(&cfg.Delta, "delta", false, "Delta mode: capture initial and final snapshots, output difference")
	fs.IntVar(&cfg.Port, "port", cfg.Port, "HTTP server port")
	fs.BoolVar(&cfg.Batch, "batch", false, "Batch mode")
	fs.BoolVar(&cfg.Stream, "stream", false, "Stream mode")
	fs.BoolVar(&cfg.Cleanup, "cleanup", false, "Cleanup temp files")
	fs.BoolVar(&cfg.Graphs, "graphs", false, "Generate graphs")
	fs.StringVar(&cfg.GraphDir, "graph-dir", "", "Graph output directory")

	var static, dynamic bool
	fs.BoolVar(&static, "static", false, "Collect static metrics")
	fs.BoolVar(&dynamic, "dynamic", false, "Collect dynamic metrics")

	return func() {
		if !static && !dynamic {
			cfg.CollectStatic = true
			cfg.CollectDynamic = true
		} else {
			cfg.CollectStatic = static
			cfg.CollectDynamic = dynamic
		}
	}
}

func GetHostname() string {
	if h, err := os.Hostname(); err == nil {
		return h
	}
	return "unknown"
}

func GenerateUUID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}
