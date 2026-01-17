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
	applyFlags := utils.GetFlags(fs, cfg)
	fs.Parse(args)
	applyFlags()

	manager := collecting.NewManager(cfg)
	defer manager.Close()

	static := &collecting.StaticMetrics{
		UUID:     cfg.UUID,
		Hostname: cfg.Hostname,
	}
	manager.CollectStatic(static)

	mux := http.NewServeMux()

	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		dynamic := &collecting.DynamicMetrics{}
		record := manager.CollectDynamic(dynamic)

		// Check DisableFlatten (inverted)
		if !cfg.DisableFlatten {
			record = exporting.FlattenRecord(record)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(record)
	})

	mux.HandleFunc("/static", func(w http.ResponseWriter, r *http.Request) {
		staticRecord := manager.GetStaticRecord()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(staticRecord)
	})

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "OK")
	})

	mux.HandleFunc("/info", func(w http.ResponseWriter, r *http.Request) {
		info := map[string]interface{}{
			"collectors": manager.CollectorNames(),
			"concurrent": cfg.Concurrent,
			"flatten":    !cfg.DisableFlatten, // Inverted for clarity
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(info)
	})

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
