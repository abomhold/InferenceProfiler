package cmd

import (
	"InferenceProfiler/pkg/collecting"
	"InferenceProfiler/pkg/serving"
	"InferenceProfiler/pkg/utils"
	"context"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func Run(args []string) {
	cfg := utils.ParseArgs(args)

	if cfg.Pprof != "" {
		go func() {
			log.Printf("pprof: http://%s/debug/pprof/", cfg.Pprof)
			if err := http.ListenAndServe(cfg.Pprof, nil); err != nil {
				log.Fatal(err)
			}
		}()
	}

	manager := collecting.NewManager(cfg)
	switch cfg.Mode {
	case "server":
		runServer(manager, cfg.ServerPort)
	case "snapshot":
		runSnapshot(manager, cfg)
	default:
		runContinuous(manager, cfg)
	}
}

func runServer(manager *collecting.Manager, port int) {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	server := serving.NewServer(manager)

	go func() {
		if err := server.ListenAndServe(port); err != nil {
			log.Fatal(err)
		}
	}()
	log.Printf("server: listening on 0.0.0.0:%d", port)

	<-ctx.Done()
	log.Println("shutting down...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("server: shutdown error: %v", err)
	}
}

func runContinuous(manager *collecting.Manager, cfg *utils.Config) {
	log.Printf("continuous: interval=%dms", cfg.Interval)

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		waitForSignal()
		cancel()
	}()

	w := utils.NewWriter(cfg)
	defer w.Close()

	manager.Continuous(ctx, w)
}

func runSnapshot(manager *collecting.Manager, cfg *utils.Config) {
	log.Println("snapshot: collecting")

	w := utils.NewWriter(cfg)
	defer w.Close()

	manager.Snapshot(context.Background(), w)
}

func waitForSignal() {
	log.Println("press Ctrl+C to stop")
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	log.Println("shutting down...")
}
