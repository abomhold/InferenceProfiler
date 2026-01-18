package main

import (
	"InferenceProfiler/pkg/cmd"
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]
	args := os.Args[2:]

	switch command {
	case "profiler", "pr":
		cmd.Profiler(args)
	case "profile", "p":
		cmd.Profile(args)
	case "snapshot", "ss":
		cmd.Snapshot(args)
	case "serve", "s":
		cmd.Serve(args)
	case "graph", "g":
		cmd.Graph(args)
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Print(`InferenceProfiler - System metrics collector for ML inference workloads

Usage:
  infpro <command> [flags]

Commands:
  profiler, pr    Continuous profiling until Ctrl+C
  profile, p      Profile a command until completion
  snapshot, ss    Capture a single metrics snapshot
  serve, s        Start HTTP metrics server
  graph, g        Generate graphs from data files

Collection Flags:
  -concurrent          Enable concurrent collection
  -vm                  Collect VM metrics (default: true)
  -container           Collect container metrics
  -process             Collect process metrics
  -nvidia              Collect NVIDIA GPU metrics
  -vllm                Collect vLLM metrics
  -gpu-procs           Collect GPU process info

Output Flags:
  -format string       Output format: jsonl, parquet, csv, tsv (default: parquet)
  -output string       Output file path
  -flatten             Flatten nested structures
  -batch               Batch mode: collect in memory, write at end
  -stream              Stream mode: write directly to file (default)

Graph Flags:
  -graphs              Generate graphs after collection
  -graph-dir string    Graph output directory

Profiler/Profile Flags:
  -interval int        Collection interval in ms (default: 100)
  -cleanup             Remove intermediate JSON files

Serve Flags:
  -port int            HTTP server port (default: 8080)

Examples:
  # Continuous profiling
  infpro profiler -process -nvidia -interval 500 -output metrics.parquet

  # Profile a command
  infpro profile -concurrent -nvidia -- python train.py

  # Single snapshot
  infpro snapshot -process -nvidia -flatten -output snapshot.json

  # HTTP server
  infpro serve -port 9090 -process -nvidia

  # Generate graphs
  infpro graph -output graphs/ metrics.parquet
`)
}
