package commands

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/spf13/cobra"

	"InferenceProfiler/pkg/collectors"
	"InferenceProfiler/pkg/config"
	"InferenceProfiler/pkg/formatting"
)

var serveAddr string

// NewServeCmd creates the serve subcommand.
func NewServeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "serve",
		Aliases: []string{"s"},
		Short:   "Run HTTP server exposing metrics",
		Long: `Run an HTTP server that exposes metrics via endpoints.

Endpoints:
  /          Status page with links
  /static    Static system metrics (JSON)
  /dynamic   Dynamic system metrics (JSON)
  /both      Combined metrics (JSON)

Example:
  infprofiler serve --addr :8080
  infprofiler serve --addr 0.0.0.0:9090`,
		RunE: runServe,
	}

	Cfg.AddCollectionFlags(cmd)
	Cfg.AddSystemFlags(cmd)

	cmd.Flags().StringVar(&serveAddr, "addr", ":8080", "Listen address")

	return cmd
}

type metricsServer struct {
	manager *collectors.Manager
	static  *collectors.StaticMetrics
	mu      sync.RWMutex
}

func runServe(cmd *cobra.Command, args []string) error {
	Cfg.ApplyDefaults()

	manager := collectors.NewManager(Cfg)
	defer manager.Close()

	static := &collectors.StaticMetrics{
		UUID:     Cfg.UUID,
		VMID:     Cfg.VMID,
		Hostname: Cfg.Hostname,
		BootTime: config.GetBootTime(),
	}
	manager.CollectStatic(static)

	server := &metricsServer{
		manager: manager,
		static:  static,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", server.handleIndex)
	mux.HandleFunc("/static", server.handleStatic)
	mux.HandleFunc("/dynamic", server.handleDynamic)
	mux.HandleFunc("/both", server.handleBoth)

	log.Printf("Starting metrics server on %s", serveAddr)
	log.Printf("Endpoints: /, /static, /dynamic, /both")

	return http.ListenAndServe(serveAddr, mux)
}

func (s *metricsServer) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, `<!DOCTYPE html>
<html>
<head><title>InferenceProfiler</title></head>
<body>
<h1>InferenceProfiler Metrics Server</h1>
<ul>
<li><a href="/static">Static Metrics</a> - System information</li>
<li><a href="/dynamic">Dynamic Metrics</a> - Current utilization</li>
<li><a href="/both">All Metrics</a> - Combined view</li>
</ul>
</body>
</html>`)
}

func (s *metricsServer) handleStatic(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	static := s.static
	s.mu.RUnlock()

	s.manager.CollectStatic(static)

	writeJSONResponse(w, s.manager.GetStaticRecord())
}

func (s *metricsServer) handleDynamic(w http.ResponseWriter, r *http.Request) {
	dynamic := &collectors.DynamicMetrics{}
	data := s.manager.CollectDynamic(dynamic)
	data = formatting.FlattenRecord(data)
	writeJSONResponse(w, data)
}

func (s *metricsServer) handleBoth(w http.ResponseWriter, r *http.Request) {
	dynamic := &collectors.DynamicMetrics{}
	data := s.manager.CollectDynamic(dynamic)
	data = formatting.FlattenRecord(data)
	writeJSONResponse(w, data)
}

func writeJSONResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")

	if err := enc.Encode(data); err != nil {
		http.Error(w, fmt.Sprintf("JSON error: %v", err), http.StatusInternalServerError)
	}
}
