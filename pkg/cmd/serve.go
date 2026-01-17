package cmd

import (
	"InferenceProfiler/pkg/collecting"
	"InferenceProfiler/pkg/exporting"
	"InferenceProfiler/pkg/utils"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"
)

func Serve(args []string) {
	fs := flag.NewFlagSet("serve", flag.ExitOnError)
	cfg := utils.NewConfig()

	// Collection flags
	fs.BoolVar(&cfg.Concurrent, "concurrent", false, "Enable concurrent collection")
	fs.BoolVar(&cfg.EnableVM, "vm", true, "Collect VM metrics")
	fs.BoolVar(&cfg.EnableContainer, "container", false, "Collect container metrics")
	fs.BoolVar(&cfg.EnableProcess, "process", false, "Collect process metrics")
	fs.BoolVar(&cfg.EnableNvidia, "nvidia", false, "Collect NVIDIA GPU metrics")
	fs.BoolVar(&cfg.EnableVLLM, "vllm", false, "Collect vLLM metrics")
	fs.BoolVar(&cfg.CollectGPUProcesses, "gpu-procs", false, "Collect GPU process info")

	// Server flags
	fs.IntVar(&cfg.Port, "port", 8080, "HTTP server port")
	fs.BoolVar(&cfg.Flatten, "flatten", false, "Flatten nested structures in responses")

	fs.Parse(args)

	// Initialize collector manager
	manager := collecting.NewManager(cfg)
	defer manager.Close()

	// Collect static metrics once
	static := &collecting.StaticMetrics{
		UUID:     cfg.UUID,
		Hostname: cfg.Hostname,
	}
	manager.CollectStatic(static)

	// Setup HTTP handlers
	mux := http.NewServeMux()

	// Metrics endpoint
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		dynamic := &collecting.DynamicMetrics{}
		record := manager.CollectDynamic(dynamic)

		if cfg.Flatten {
			record = exporting.FlattenRecord(record)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(record)
	})

	// Static metrics endpoint
	mux.HandleFunc("/static", func(w http.ResponseWriter, r *http.Request) {
		staticRecord := manager.GetStaticRecord()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(staticRecord)
	})

	// Health check
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "OK")
	})

	// Info endpoint
	mux.HandleFunc("/info", func(w http.ResponseWriter, r *http.Request) {
		info := map[string]interface{}{
			"collectors": manager.CollectorNames(),
			"concurrent": cfg.Concurrent,
			"flatten":    cfg.Flatten,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(info)
	})

	// Start server
	addr := fmt.Sprintf(":%d", cfg.Port)
	server := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	log.Printf("HTTP server listening on %s", addr)
	log.Printf("Endpoints:")
	log.Printf("  GET /metrics - Current metrics")
	log.Printf("  GET /static  - Static metrics")
	log.Printf("  GET /health  - Health check")
	log.Printf("  GET /info    - Server info")

	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
