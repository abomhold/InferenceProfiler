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

//// Package commands provides CLI command implementations.
//package pkg

//
//import (
//	"InferenceProfiler/pkg/utils"
//	"context"
//	"encoding/json"
//	"flag"
//	"fmt"
//	"log"
//	"net/http"
//	"os"
//	"os/exec"
//	"os/signal"
//	"sync"
//	"syscall"
//	"time"
//
//	"InferenceProfiler/pkg/collecting"
//	"InferenceProfiler/pkg/exporting"
//)
//
//var Cfg = utils.New()
//var (
//	staticOnly   bool
//	dynamicOnly  bool
//	concurrent   bool
//	noVM         bool
//	noContainer  bool
//	noProcess    bool
//	noNvidia     bool
//	noVLLM       bool
//	noGPUProcs   bool
//	interval     time.Duration
//	outputDir    string
//	outputFormat string
//	outputFile   string
//	graphs       bool
//	graphOutput  string
//	serveAddr    string
//)
//
//func usage() {
//	fmt.Fprintf(os.Stderr, `InferenceProfiler - System profiler for ML inference workloads
//
//Usage: infprofiler <command> [flags] [args]
//
//Commands:
//  profiler   Continuous profiling until Ctrl+C
//  profile    Profile a command until completion
//  snapshot   Capture a single metrics snapshot
//  serve      Run HTTP server exposing metrics
//  graph      Generate PNG graphs from data files
//
//Flags (all commands):
//  -static         Only static metrics
//  -dynamic        Only dynamic metrics
//  -concurrent     Run collectors concurrently
//  -no-vm          Disable VM metrics
//  -no-container   Disable container metrics
//  -no-procs     Disable OS process metrics
//  -no-nvidia      Disable NVIDIA GPU metrics
//  -no-vllm        Disable vLLM metrics
//  -no-gpu-procs   Disable GPU process collection
//
//Run 'infpro <command> -h' for command-specific help.
//`)
//}
//
//// Execute runs the CLI.
//func Execute() {
//	if len(os.Args) < 2 {
//		usage()
//		os.Exit(1)
//	}
//
//	cmd := os.Args[1]
//	args := os.Args[2:]
//
//	var err error
//	switch cmd {
//	case "profiler", "pr":
//		err = runProfiler(args)
//	case "profile", "p":
//		err = runProfile(args)
//	case "snapshot", "ss":
//		err = runSnapshot(args)
//	case "serve", "s":
//		err = runServe(args)
//	case "graph", "g":
//		err = runGraph(args)
//	case "-h", "--help", "help":
//		usage()
//		return
//	default:
//		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", cmd)
//		usage()
//		os.Exit(1)
//	}
//
//	if err != nil {
//		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
//		os.Exit(1)
//	}
//}
//
//func addCommonFlags(fs *flag.FlagSet) {
//	fs.BoolVar(&staticOnly, "static", false, "Only static metrics")
//	fs.BoolVar(&dynamicOnly, "dynamic", false, "Only dynamic metrics")
//	fs.BoolVar(&concurrent, "concurrent", false, "Run collectors concurrently")
//	fs.BoolVar(&noVM, "no-vm", false, "Disable VM metrics")
//	fs.BoolVar(&noContainer, "no-container", false, "Disable container metrics")
//	fs.BoolVar(&noProcess, "no-procs", false, "Disable OS process metrics")
//	fs.BoolVar(&noNvidia, "no-nvidia", false, "Disable NVIDIA GPU metrics")
//	fs.BoolVar(&noVLLM, "no-vllm", false, "Disable vLLM metrics")
//	fs.BoolVar(&noGPUProcs, "no-gpu-procs", false, "Disable GPU process collection")
//}
//
//func addOutputFlags(fs *flag.FlagSet) {
//	fs.DurationVar(&interval, "interval", 100*time.Millisecond, "Collection interval")
//	fs.StringVar(&outputDir, "dir", ".", "Output directory")
//	fs.StringVar(&outputFormat, "format", "parquet", "Output format (parquet, jsonl, csv, tsv)")
//	fs.StringVar(&outputFile, "o", "", "Output file (auto-generated if empty)")
//	fs.BoolVar(&graphs, "graphs", false, "Generate PNG graphs")
//	fs.StringVar(&graphOutput, "graph-dir", "", "Graph output directory")
//}
//
//func applyConfig() {
//	Cfg.EnableVM = !noVM
//	Cfg.EnableContainer = !noContainer
//	Cfg.EnableProcess = !noProcess
//	Cfg.EnableNvidia = !noNvidia
//	Cfg.EnableVLLM = !noVLLM
//	Cfg.CollectGPUProcesses = !noGPUProcs
//	Cfg.Concurrent = concurrent
//	Cfg.Interval = interval
//	Cfg.OutputDir = outputDir
//	Cfg.OutputFormat = outputFormat
//	Cfg.OutputName = outputFile
//	Cfg.GenerateGraphs = graphs
//	Cfg.GraphOutput = graphOutput
//	Cfg.ApplyDefaults()
//}
//
//func collectBoth() bool {
//	return !staticOnly && !dynamicOnly
//}
//
//// ============================================================================
//// profiler command
//// ============================================================================
//
//func runProfiler(args []string) error {
//	fs := flag.NewFlagSet("profiler", flag.ExitOnError)
//	fs.Usage = func() {
//		fmt.Fprintf(os.Stderr, "Usage: infprofiler profiler [flags]\n\nContinuous profiling until Ctrl+C.\n\nFlags:\n")
//		fs.PrintDefaults()
//	}
//	addCommonFlags(fs)
//	addOutputFlags(fs)
//	fs.Parse(args)
//	applyConfig()
//
//	if err := Cfg.Validate(); err != nil {
//		return err
//	}
//
//	manager := collecting.NewManager(Cfg)
//	defer manager.Close()
//
//	static := &collecting.StaticMetrics{}
//	manager.CollectStatic(static)
//
//	outPath := Cfg.OutputName
//	if outPath == "" {
//		outPath = Cfg.GenerateOutputPath("profiler")
//	}
//
//	exp, err := exporting.NewExporter(outPath, Cfg.OutputFormat)
//	if err != nil {
//		return fmt.Errorf("exporter: %w", err)
//	}
//
//	sigChan := make(chan os.Signal, 1)
//	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
//
//	log.Printf("Profiler started: interval=%v output=%s", Cfg.Interval, outPath)
//
//	ticker := time.NewTicker(Cfg.Interval)
//	defer ticker.Stop()
//
//	count := 0
//	start := time.Now()
//
//	for {
//		select {
//		case <-sigChan:
//			log.Println("Interrupted")
//			return finalize(exp, count, start)
//		case <-ticker.C:
//			dynamic := &collecting.DynamicMetrics{}
//			if err := exp.Write(manager.CollectDynamic(dynamic)); err != nil {
//				log.Printf("Write error: %v", err)
//				continue
//			}
//			count++
//			if count%100 == 0 {
//				log.Printf("Samples: %d", count)
//			}
//		}
//	}
//}
//
//// ============================================================================
//// profile command
//// ============================================================================
//
//func runProfile(args []string) error {
//	fs := flag.NewFlagSet("profile", flag.ExitOnError)
//	fs.Usage = func() {
//		fmt.Fprintf(os.Stderr, "Usage: infprofiler profile [flags] -- <command> [args]\n\nProfile a command.\n\nFlags:\n")
//		fs.PrintDefaults()
//	}
//	addCommonFlags(fs)
//	addOutputFlags(fs)
//
//	// Split at --
//	dashIdx := -1
//	for i, arg := range args {
//		if arg == "--" {
//			dashIdx = i
//			break
//		}
//	}
//
//	var flagArgs, cmdArgs []string
//	if dashIdx >= 0 {
//		flagArgs = args[:dashIdx]
//		cmdArgs = args[dashIdx+1:]
//	} else {
//		flagArgs = args
//	}
//
//	fs.Parse(flagArgs)
//	applyConfig()
//
//	if len(cmdArgs) == 0 {
//		return fmt.Errorf("no command specified\nUsage: infprofiler profile [flags] -- <command>")
//	}
//
//	if err := Cfg.Validate(); err != nil {
//		return err
//	}
//
//	manager := collecting.NewManager(Cfg)
//	defer manager.Close()
//
//	static := &collecting.StaticMetrics{}
//	manager.CollectStatic(static)
//
//	outPath := Cfg.OutputName
//	if outPath == "" {
//		outPath = Cfg.GenerateOutputPath("profile")
//	}
//
//	exp, err := exporting.NewExporter(outPath, Cfg.OutputFormat)
//	if err != nil {
//		return fmt.Errorf("exporter: %w", err)
//	}
//
//	targetCmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
//	targetCmd.Stdout, targetCmd.Stderr, targetCmd.Stdin = os.Stdout, os.Stderr, os.Stdin
//
//	log.Printf("Profiling: %v", cmdArgs)
//
//	if err := targetCmd.Start(); err != nil {
//		exp.Close()
//		return fmt.Errorf("start: %w", err)
//	}
//
//	ctx, cancel := context.WithCancel(context.Background())
//	done := make(chan struct{})
//
//	go func() {
//		defer close(done)
//		ticker := time.NewTicker(Cfg.Interval)
//		defer ticker.Stop()
//		for {
//			select {
//			case <-ctx.Done():
//				return
//			case <-ticker.C:
//				dynamic := &collecting.DynamicMetrics{}
//				exp.Write(manager.CollectDynamic(dynamic))
//			}
//		}
//	}()
//
//	start := time.Now()
//	cmdErr := targetCmd.Wait()
//	cancel()
//	<-done
//
//	log.Printf("Command finished in %v", time.Since(start))
//	exp.Close()
//	log.Printf("Output: %s", exp.Path())
//
//	if Cfg.GenerateGraphs {
//		genGraphs(exp.Path())
//	}
//
//	return cmdErr
//}
//
//// ============================================================================
//// snapshot command
//// ============================================================================
//
//func runSnapshot(args []string) error {
//	fs := flag.NewFlagSet("snapshot", flag.ExitOnError)
//	fs.Usage = func() {
//		fmt.Fprintf(os.Stderr, "Usage: infprofiler snapshot [flags]\n\nCapture metrics snapshot as JSON.\n\nFlags:\n")
//		fs.PrintDefaults()
//	}
//	addCommonFlags(fs)
//	var output string
//	fs.StringVar(&output, "o", "", "Output file (default: stdout)")
//	fs.Parse(args)
//	applyConfig()
//
//	manager := collecting.NewManager(Cfg)
//	defer manager.Close()
//
//	result := make(exporting.Record)
//
//	doStatic := collectBoth() || staticOnly
//	doDynamic := collectBoth() || dynamicOnly
//
//	if doStatic {
//		static := &collecting.StaticMetrics{}
//		manager.CollectStatic(static)
//		for k, v := range manager.GetStaticRecord() {
//			result[k] = v
//		}
//	}
//
//	if doDynamic {
//		dynamic := &collecting.DynamicMetrics{}
//		for k, v := range exporting.FlattenRecord(manager.CollectDynamic(dynamic)) {
//			result[k] = v
//		}
//	}
//
//	data, _ := json.MarshalIndent(result, "", "  ")
//
//	if output == "" {
//		fmt.Println(string(data))
//	} else {
//		if err := os.WriteFile(output, data, 0644); err != nil {
//			return err
//		}
//		fmt.Fprintf(os.Stderr, "Written: %s\n", output)
//	}
//	return nil
//}
//
//// ============================================================================
//// serve command
//// ============================================================================
//
//func runServe(args []string) error {
//	fs := flag.NewFlagSet("serve", flag.ExitOnError)
//	fs.Usage = func() {
//		fmt.Fprintf(os.Stderr, "Usage: infprofiler serve [flags]\n\nHTTP server: / /static /dynamic /both\n\nFlags:\n")
//		fs.PrintDefaults()
//	}
//	addCommonFlags(fs)
//	fs.StringVar(&serveAddr, "addr", ":8080", "Listen address")
//	fs.Parse(args)
//	applyConfig()
//
//	manager := collecting.NewManager(Cfg)
//	defer manager.Close()
//
//	static := &collecting.StaticMetrics{}
//	manager.CollectStatic(static)
//
//	srv := &server{manager: manager, static: static}
//
//	mux := http.NewServeMux()
//	mux.HandleFunc("/", srv.index)
//	mux.HandleFunc("/static", srv.handleStatic)
//	mux.HandleFunc("/dynamic", srv.handleDynamic)
//	mux.HandleFunc("/both", srv.handleBoth)
//
//	log.Printf("Serving on %s", serveAddr)
//	return http.ListenAndServe(serveAddr, mux)
//}
//
//type server struct {
//	manager *collecting.Manager
//	static  *collecting.StaticMetrics
//	mu      sync.RWMutex
//}
//
//func (s *server) index(w http.ResponseWriter, r *http.Request) {
//	if r.URL.Path != "/" {
//		http.NotFound(w, r)
//		return
//	}
//	w.Header().Set("Content-Type", "text/html")
//	fmt.Fprint(w, `<!DOCTYPE html><html><body><h1>InferenceProfiler</h1><ul><li><a href="/static">Static</a></li><li><a href="/dynamic">Dynamic</a></li><li><a href="/both">Both</a></li></ul></body></html>`)
//}
//
//func (s *server) handleStatic(w http.ResponseWriter, r *http.Request) {
//	s.mu.RLock()
//	defer s.mu.RUnlock()
//	s.manager.CollectStatic(s.static)
//	writeJSON(w, s.manager.GetStaticRecord())
//}
//
//func (s *server) handleDynamic(w http.ResponseWriter, r *http.Request) {
//	dynamic := &collecting.DynamicMetrics{}
//	writeJSON(w, exporting.FlattenRecord(s.manager.CollectDynamic(dynamic)))
//}
//
//func (s *server) handleBoth(w http.ResponseWriter, r *http.Request) {
//	s.mu.RLock()
//	s.manager.CollectStatic(s.static)
//	staticRec := s.manager.GetStaticRecord()
//	s.mu.RUnlock()
//
//	dynamic := &collecting.DynamicMetrics{}
//	dynamicRec := exporting.FlattenRecord(s.manager.CollectDynamic(dynamic))
//
//	result := make(exporting.Record)
//	for k, v := range staticRec {
//		result[k] = v
//	}
//	for k, v := range dynamicRec {
//		result[k] = v
//	}
//	writeJSON(w, result)
//}
//
//func writeJSON(w http.ResponseWriter, data interface{}) {
//	w.Header().Set("Content-Type", "application/json")
//	enc := json.NewEncoder(w)
//	enc.SetIndent("", "  ")
//	enc.Encode(data)
//}
//
//// ============================================================================
//// graph command
//// ============================================================================
//
//func runGraph(args []string) error {
//	fs := flag.NewFlagSet("graph", flag.ExitOnError)
//	fs.Usage = func() {
//		fmt.Fprintf(os.Stderr, "Usage: infprofiler graph [flags] <input-file>\n\nGenerate PNG graphs.\n\nFlags:\n")
//		fs.PrintDefaults()
//	}
//	addCommonFlags(fs)
//	var output string
//	fs.StringVar(&output, "o", "", "Output directory")
//	fs.Parse(args)
//	applyConfig()
//
//	if fs.NArg() < 1 {
//		return fmt.Errorf("input file required")
//	}
//
//	input := fs.Arg(0)
//	if _, err := os.Stat(input); err != nil {
//		return fmt.Errorf("file not found: %s", input)
//	}
//
//	if output == "" {
//		output = Cfg.GenerateGraphPath(input)
//	}
//
//	gen, err := exporting.NewGenerator(input, output, "png")
//	if err != nil {
//		return err
//	}
//	if err := gen.Generate(); err != nil {
//		return err
//	}
//	fmt.Printf("Graphs: %s\n", output)
//	return nil
//}
//
//// ============================================================================
//// helpers
//// ============================================================================
//
//func finalize(exp *exporting.Exporter, count int, start time.Time) error {
//	log.Printf("Done: %d samples in %v", count, time.Since(start))
//	if err := exp.Close(); err != nil {
//		return err
//	}
//	log.Printf("Output: %s", exp.Path())
//	if Cfg.GenerateGraphs {
//		genGraphs(exp.Path())
//	}
//	return nil
//}
//
//func genGraphs(dataPath string) {
//	gPath := Cfg.GenerateGraphPath(dataPath)
//	log.Printf("Generating graphs: %s", gPath)
//	gen, err := exporting.NewGenerator(dataPath, gPath, "png")
//	if err != nil {
//		log.Printf("Graph error: %v", err)
//		return
//	}
//	if err := gen.Generate(); err != nil {
//		log.Printf("Graph error: %v", err)
//	}
//}
