package cmd

import (
	"InferenceProfiler/pkg/exporting"
	"InferenceProfiler/pkg/utils"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"
)

// profileSession holds state for an active profiling session
type profileSession struct {
	id            string
	initialRecord exporting.Record
	startTime     time.Time
	active        bool
}

// server holds the HTTP server state
type server struct {
	ctx      *CmdContext
	sessions map[string]*profileSession
	mu       sync.RWMutex
}

func Serve(args []string) {
	ctx, cleanup := InitCmd("serve", args)
	defer cleanup()

	srv := &server{
		ctx:      ctx,
		sessions: make(map[string]*profileSession),
	}

	mux := http.NewServeMux()

	// Existing endpoints
	mux.HandleFunc("/metrics", srv.handleMetrics)
	mux.HandleFunc("/static", srv.handleStatic)
	mux.HandleFunc("/health", srv.handleHealth)
	mux.HandleFunc("/info", srv.handleInfo)

	// New profiling endpoints
	mux.HandleFunc("/profile/start", srv.handleProfileStart)
	mux.HandleFunc("/profile/stop", srv.handleProfileStop)
	mux.HandleFunc("/profile/delta", srv.handleProfileDelta)
	mux.HandleFunc("/profile/status", srv.handleProfileStatus)

	// Convenience endpoint for one-shot delta profiling
	mux.HandleFunc("/profile/snapshot", srv.handleProfileSnapshot)

	addr := fmt.Sprintf(":%d", ctx.Config.Port)
	httpServer := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	log.Printf("HTTP server listening on %s", addr)
	log.Printf("Endpoints:")
	log.Printf("  GET  /metrics          - Current metrics snapshot")
	log.Printf("  GET  /static           - Static metrics")
	log.Printf("  GET  /health           - Health check")
	log.Printf("  GET  /info             - Server info")
	log.Printf("  POST /profile/start    - Start profiling session (returns session_id)")
	log.Printf("  POST /profile/stop     - Stop session and return delta (requires session_id)")
	log.Printf("  GET  /profile/delta    - Get current delta without stopping (requires session_id)")
	log.Printf("  GET  /profile/status   - Get session status (requires session_id)")
	log.Printf("  POST /profile/snapshot - One-shot: wait duration_ms, return delta")

	if err := httpServer.ListenAndServe(); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func (s *server) handleMetrics(w http.ResponseWriter, r *http.Request) {
	record := CollectSnapshot(s.ctx.Manager, ShouldExpandAll(s.ctx.Config))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(record)
}

func (s *server) handleStatic(w http.ResponseWriter, r *http.Request) {
	staticRecord := s.ctx.Manager.GetStaticRecord()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(staticRecord)
}

func (s *server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "OK")
}

func (s *server) handleInfo(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	activeSessions := 0
	for _, sess := range s.sessions {
		if sess.active {
			activeSessions++
		}
	}
	s.mu.RUnlock()

	info := map[string]interface{}{
		"collectors":      s.ctx.Manager.CollectorNames(),
		"concurrent":      s.ctx.Config.Concurrent,
		"expandAll":       ShouldExpandAll(s.ctx.Config),
		"active_sessions": activeSessions,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(info)
}

func (s *server) handleProfileStart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Generate session ID
	sessionID := utils.GenerateUUID()

	// Capture initial snapshot
	initialRecord := CollectSnapshot(s.ctx.Manager, ShouldExpandAll(s.ctx.Config))

	// Store session
	s.mu.Lock()
	s.sessions[sessionID] = &profileSession{
		id:            sessionID,
		initialRecord: initialRecord,
		startTime:     time.Now(),
		active:        true,
	}
	s.mu.Unlock()

	response := map[string]interface{}{
		"session_id": sessionID,
		"started_at": time.Now().UnixMilli(),
		"message":    "Profiling session started. Call /profile/stop with session_id to get delta.",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *server) handleProfileStop(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	sessionID := r.URL.Query().Get("session_id")
	if sessionID == "" {
		http.Error(w, "session_id required", http.StatusBadRequest)
		return
	}

	s.mu.Lock()
	session, exists := s.sessions[sessionID]
	if !exists {
		s.mu.Unlock()
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	if !session.active {
		s.mu.Unlock()
		http.Error(w, "Session already stopped", http.StatusBadRequest)
		return
	}

	session.active = false
	delete(s.sessions, sessionID) // Clean up session
	s.mu.Unlock()

	// Capture final snapshot
	finalRecord := CollectSnapshot(s.ctx.Manager, ShouldExpandAll(s.ctx.Config))
	elapsed := time.Since(session.startTime)

	// Calculate delta
	deltaRecord := ComputeDelta(session.initialRecord, finalRecord, elapsed.Milliseconds())

	response := map[string]interface{}{
		"session_id":  sessionID,
		"duration_ms": elapsed.Milliseconds(),
		"delta":       deltaRecord,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *server) handleProfileDelta(w http.ResponseWriter, r *http.Request) {
	sessionID := r.URL.Query().Get("session_id")
	if sessionID == "" {
		http.Error(w, "session_id required", http.StatusBadRequest)
		return
	}

	s.mu.RLock()
	session, exists := s.sessions[sessionID]
	if !exists {
		s.mu.RUnlock()
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	if !session.active {
		s.mu.RUnlock()
		http.Error(w, "Session already stopped", http.StatusBadRequest)
		return
	}

	initialRecord := session.initialRecord
	startTime := session.startTime
	s.mu.RUnlock()

	// Capture current snapshot (without stopping session)
	currentRecord := CollectSnapshot(s.ctx.Manager, ShouldExpandAll(s.ctx.Config))
	elapsed := time.Since(startTime)

	// Calculate delta
	deltaRecord := ComputeDelta(initialRecord, currentRecord, elapsed.Milliseconds())

	response := map[string]interface{}{
		"session_id":  sessionID,
		"duration_ms": elapsed.Milliseconds(),
		"active":      true,
		"delta":       deltaRecord,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *server) handleProfileStatus(w http.ResponseWriter, r *http.Request) {
	sessionID := r.URL.Query().Get("session_id")
	if sessionID == "" {
		// Return all sessions
		s.mu.RLock()
		sessions := make([]map[string]interface{}, 0, len(s.sessions))
		for id, sess := range s.sessions {
			sessions = append(sessions, map[string]interface{}{
				"session_id":  id,
				"active":      sess.active,
				"started_at":  sess.startTime.UnixMilli(),
				"duration_ms": time.Since(sess.startTime).Milliseconds(),
			})
		}
		s.mu.RUnlock()

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"sessions": sessions,
		})
		return
	}

	s.mu.RLock()
	session, exists := s.sessions[sessionID]
	if !exists {
		s.mu.RUnlock()
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	response := map[string]interface{}{
		"session_id":  session.id,
		"active":      session.active,
		"started_at":  session.startTime.UnixMilli(),
		"duration_ms": time.Since(session.startTime).Milliseconds(),
	}
	s.mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *server) handleProfileSnapshot(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get duration from query parameter (default to interval from config)
	durationStr := r.URL.Query().Get("duration_ms")
	duration := time.Duration(s.ctx.Config.Interval) * time.Millisecond

	if durationStr != "" {
		durationMs, err := strconv.ParseInt(durationStr, 10, 64)
		if err != nil {
			http.Error(w, "Invalid duration_ms", http.StatusBadRequest)
			return
		}
		duration = time.Duration(durationMs) * time.Millisecond
	}

	expandAll := ShouldExpandAll(s.ctx.Config)

	// Capture initial snapshot
	initialRecord := CollectSnapshot(s.ctx.Manager, expandAll)
	startTime := time.Now()

	// Wait for duration
	if duration > 0 {
		time.Sleep(duration)
	}

	// Capture final snapshot
	finalRecord := CollectSnapshot(s.ctx.Manager, expandAll)
	elapsed := time.Since(startTime)

	// Calculate delta
	deltaRecord := ComputeDelta(initialRecord, finalRecord, elapsed.Milliseconds())

	response := map[string]interface{}{
		"duration_ms": elapsed.Milliseconds(),
		"delta":       deltaRecord,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
