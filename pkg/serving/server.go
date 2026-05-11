package serving

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"InferenceProfiler/pkg/collecting"
	"InferenceProfiler/pkg/utils"
)

type Server struct {
	manager    *collecting.Manager
	httpServer *http.Server

	mu        sync.Mutex
	uuid      string
	mode      string
	startTime time.Time
	count     int64
	cancel    context.CancelFunc
	done      chan struct{}
}

func NewServer(manager *collecting.Manager) *Server {
	return &Server{manager: manager}
}

func (s *Server) ListenAndServe(port int) error {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", s.handleHealth)
	mux.HandleFunc("GET /snapshot", s.handleSnapshot)
	mux.HandleFunc("GET /collect", s.handleCollectGet)
	mux.HandleFunc("PUT /collect", s.handleCollectPut)
	mux.HandleFunc("DELETE /collect", s.handleCollectDelete)
	mux.HandleFunc("GET /files", s.handleListFiles)
	mux.HandleFunc("GET /files/{uuid}", s.handleGetFile)

	s.httpServer = &http.Server{
		Addr:              fmt.Sprintf("%s:%d", "0.0.0.0", port),
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	if s.httpServer == nil {
		return nil
	}
	s.mu.Lock()
	if s.collecting() {
		slog.Info("server: stopping active collection before shutdown", "uuid", s.uuid)
		s.cancel()
		done := s.done
		s.mu.Unlock()
		<-done
	} else {
		s.mu.Unlock()
	}
	return s.httpServer.Shutdown(ctx)
}

func (s *Server) collecting() bool {
	return s.cancel != nil
}

func resolveUUID(requestUUID string) string {
	if requestUUID != "" {
		return requestUUID
	}
	return utils.GenerateUUID()
}

func (s *Server) startContinuous(requestUUID string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.collecting() {
		return "", fmt.Errorf("already collecting uuid=%s", s.uuid)
	}

	uuid := resolveUUID(requestUUID)
	ctx, cancel := context.WithCancel(context.Background())

	cfg := *s.manager.Config()
	cfg.UUID = uuid
	w := utils.NewWriter(&cfg)

	s.uuid = uuid
	s.mode = "continuous"
	s.startTime = time.Now()
	s.count = 0
	s.cancel = cancel
	s.done = make(chan struct{})

	go func() {
		defer close(s.done)
		defer w.Close()
		count, err := s.manager.Continuous(ctx, w)
		if err != nil {
			slog.Error("server: continuous stopped with error", "uuid", uuid, "error", err)
		}

		s.mu.Lock()
		s.count = int64(count)
		s.cancel = nil
		s.mu.Unlock()
	}()

	slog.Info("server: started continuous", "uuid", uuid)
	return uuid, nil
}

func (s *Server) stop() (string, error) {
	s.mu.Lock()
	if !s.collecting() {
		s.mu.Unlock()
		return "", fmt.Errorf("not collecting")
	}

	s.cancel()
	uuid := s.uuid
	done := s.done
	s.mu.Unlock()

	<-done
	slog.Info("server: stopped collection", "uuid", uuid)
	return uuid, nil
}

func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "ok")
}

func (s *Server) handleCollectGet(w http.ResponseWriter, _ *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

	resp := map[string]any{
		"collectors": s.manager.InitResults(),
	}

	if s.collecting() {
		resp["state"] = "collecting"
		resp["mode"] = s.mode
		resp["uuid"] = s.uuid
		resp["elapsed"] = time.Since(s.startTime).String()
	} else {
		resp["state"] = "idle"
		if s.uuid != "" {
			resp["last_uuid"] = s.uuid
			resp["last_mode"] = s.mode
			resp["last_count"] = s.count
		}
	}

	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) handleSnapshot(w http.ResponseWriter, _ *http.Request) {
	tick := s.manager.SnapshotTick()
	resp := map[string]any{
		"static": s.manager.StaticData(),
		"tick":   tick,
	}
	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) handleCollectPut(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UUID string `json:"uuid"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	uuid, err := s.startContinuous(req.UUID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"state": "started",
		"uuid":  uuid,
	})
}

func (s *Server) handleCollectDelete(w http.ResponseWriter, _ *http.Request) {
	uuid, err := s.stop()
	if err != nil {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"state": "stopped", "uuid": uuid})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func (s *Server) outputDir() string {
	return s.manager.Config().OutputDir
}

func (s *Server) requireOutputDir(w http.ResponseWriter) (string, bool) {
	dir := s.outputDir()
	if dir == "" {
		http.Error(w, "no output directory configured", http.StatusBadRequest)
		return "", false
	}
	return dir, true
}

func (s *Server) handleListFiles(w http.ResponseWriter, r *http.Request) {
	outDir, ok := s.requireOutputDir(w)
	if !ok {
		return
	}

	filter := r.URL.Query().Get("uuid")
	entries, err := os.ReadDir(outDir)
	if err != nil {
		http.Error(w, "failed to read output directory: "+err.Error(), http.StatusInternalServerError)
		return
	}

	type fileEntry struct {
		Name string `json:"name"`
		Size int64  `json:"size"`
	}

	files := make([]fileEntry, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if filter != "" && !strings.HasPrefix(e.Name(), filter) {
			continue
		}
		if info, err := e.Info(); err == nil {
			files = append(files, fileEntry{Name: e.Name(), Size: info.Size()})
		}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"files": files,
		"count": len(files),
	})
}

func (s *Server) handleGetFile(w http.ResponseWriter, r *http.Request) {
	outDir, ok := s.requireOutputDir(w)
	if !ok {
		return
	}

	uuid := r.PathValue("uuid")
	entries, err := os.ReadDir(outDir)
	if err != nil {
		http.Error(w, "failed to read output directory: "+err.Error(), http.StatusInternalServerError)
		return
	}

	for _, e := range entries {
		if !e.IsDir() && strings.HasPrefix(e.Name(), uuid) {
			http.ServeFile(w, r, filepath.Join(outDir, e.Name()))
			return
		}
	}

	http.Error(w, "not found", http.StatusNotFound)
}
