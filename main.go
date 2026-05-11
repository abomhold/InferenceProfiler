package main

import (
	"fmt"
	"os"
	"slices"

	"InferenceProfiler/pkg/cmd"
)

func main() {
	args := os.Args[1:]

	if hasFlag(args, "-h", "--help", "-help") {
		printUsage()
		return
	}

	cmd.Run(args)
}

func hasFlag(args []string, flags ...string) bool {
	for _, flag := range flags {
		if slices.Contains(args, flag) {
			return true
		}
	}
	return false
}

func printUsage() {
	fmt.Print(`InferenceProfiler - System metrics collector for ML inference workloads

Usage:
  infpro [command] [flags]

Commands:
  continuous, c   Collect on a fixed interval until Ctrl+C / SIGTERM (default)
  snapshot, s     Single collection pass, then exit
  server, ser     HTTP API server for remote control

Output:
  Each run writes a static line (config + system info) followed by dynamic
  records. Output is JSONL whether to stdout or to a file.

    stdout (default)    static line, then dynamic line(s)
    -output DIR         writes to DIR/{uuid}.jsonl

  Server mode endpoints:
    GET    /health           Health check
    GET    /snapshot         Live state: {"static": {...}, "tick": {...}}
                             (tick is non-empty only while collecting)
    GET    /collect          Current state and run info
    PUT    /collect          Start a continuous run (body: {"uuid": "..."})
    DELETE /collect          Stop and flush
    GET    /files            List output files (optional ?uuid=xxx)
    GET    /files/{uuid}     Stream the file whose name starts with {uuid}
                             (supports Range header)

Collection flags:
  -no-vm           Disable VM metrics (cpu, mem, disk, net)
  -no-container    Disable container/cgroup metrics
  -no-procs        Disable process metrics
  -no-nvidia       Disable NVIDIA GPU metrics
  -no-vllm         Disable vLLM metrics
  -no-vllm-hist    Disable vLLM histogram collection
  -vllm-endpoint URL    vLLM metrics endpoint (default: http://localhost:8000/metrics)
  -disabled LIST   Comma-separated collectors to disable
                   (vm,container,process,nvidia,vllm,vllm-hist)

Output flags:
  -output DIR      Output directory (default: stdout)
  -flatten         Flatten nested structs to top-level keys
  -uuid ID         Set run UUID (default: random)

Timing flags:
  -interval MS     Collection interval in milliseconds (default: 1000)

Server flags:
  -port PORT       HTTP port (default: 8888, bound on 0.0.0.0)

Debug flags:
  -debug           Verbose debug logging to stderr
  -poll-stats      Show per-collector poller statistics on exit
  -pprof ADDR      Enable pprof server (e.g. localhost:6060)

Environment overrides:
  Every flag has an INFPRO_<NAME> environment variable equivalent, where
  <NAME> is the upper-cased flag with dashes replaced by underscores.
  Examples: INFPRO_INTERVAL=500, INFPRO_OUTPUT=/tmp/out, INFPRO_DEBUG=1.
  Command-line flags take precedence over environment variables.

Examples:
  infpro                                    Continuous to stdout at 1s
  infpro s                                  Snapshot to stdout
  infpro snapshot -output ./metrics         Snapshot to file
  infpro c -output ./metrics                Continuous to file
  infpro -interval 100                      Continuous at 100ms
  infpro -no-nvidia -no-vllm                Skip GPU and vLLM collectors
  infpro -disabled vm,process               Same idea via -disabled
  infpro -poll-stats                        Show timing stats on Ctrl+C
  infpro server -output ./data              Server mode on default port
  infpro ser -port 9090                     Server on a custom port
  infpro -debug -interval 500               Verbose debug output
`)
}
