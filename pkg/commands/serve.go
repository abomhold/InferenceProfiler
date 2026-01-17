package commands

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/spf13/cobra"

	"InferenceProfiler/pkg/collecting"
	"InferenceProfiler/pkg/config"
	"InferenceProfiler/pkg/metrics"
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
	manager    *collecting.Manager
	baseStatic *metrics.BaseStatic
	mu         sync.RWMutex
}

func runServe(cmd *cobra.Command, args []string) error {
	Cfg.ApplyDefaults()

	// Initialize collector manager
	manager := collecting.NewManager(Cfg)
	defer manager.Close()

	// Create base static
	baseStatic := &metrics.BaseStatic{
		UUID:     Cfg.UUID,
		VMID:     Cfg.VMID,
		Hostname: Cfg.Hostname,
		BootTime: config.GetBootTime(),
	}

	// Collect initial static metrics
	manager.CollectStatic(baseStatic)

	server := &metricsServer{
		manager:    manager,
		baseStatic: baseStatic,
	}

	// Setup routes
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
	baseStatic := s.baseStatic
	s.mu.RUnlock()

	// Re-collect static to get current values
	s.manager.CollectStatic(baseStatic)

	data := metrics.Record{
		"uuid":      baseStatic.UUID,
		"vId":       baseStatic.VMID,
		"vHostname": baseStatic.Hostname,
		"vBootTime": baseStatic.BootTime,
	}

	writeJSONResponse(w, data)
}

func (s *metricsServer) handleDynamic(w http.ResponseWriter, r *http.Request) {
	baseDynamic := &metrics.BaseDynamic{}
	data := s.manager.CollectDynamic(baseDynamic)
	writeJSONResponse(w, data)
}

func (s *metricsServer) handleBoth(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	baseStatic := s.baseStatic
	s.mu.RUnlock()

	baseDynamic := &metrics.BaseDynamic{}
	dynamicData := s.manager.CollectDynamic(baseDynamic)

	// Merge static into dynamic
	dynamicData["uuid"] = baseStatic.UUID
	dynamicData["vId"] = baseStatic.VMID
	dynamicData["vHostname"] = baseStatic.Hostname
	dynamicData["vBootTime"] = baseStatic.BootTime

	writeJSONResponse(w, dynamicData)
}

func writeJSONResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")

	if err := enc.Encode(data); err != nil {
		http.Error(w, fmt.Sprintf("JSON error: %v", err), http.StatusInternalServerError)
	}
}
