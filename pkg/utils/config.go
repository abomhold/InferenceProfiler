package utils

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/google/uuid"
)

const (
	DefaultVLLMEndpoint = "http://localhost:8000/metrics"
	DefaultMode         = "continuous"
	FieldSeparatorColon = ":"
	FieldSeparatorSpace = " "
	UnavailableValue    = "unavailable"
)

type Config struct {
	Mode                  string
	UUID                  string
	OutputDir             string
	Flatten               bool
	Interval              int
	Debug                 bool
	PollStats             bool
	DisableVM             bool
	DisableContainer      bool
	DisableProcess        bool
	DisableNvidia         bool
	DisableVLLM           bool
	DisableVLLMHistograms bool
	VLLMEndpoint          string
	Pprof                 string
	ServerPort            int
}

func ParseArgs(args []string) *Config {
	cfg := &Config{}
	cfg.Mode, args = parseMode(args)
	fs := flag.NewFlagSet("InferenceProfiler", flag.ExitOnError)

	fs.StringVar(&cfg.UUID, "uuid", GenerateUUID(), "Unique identifier (default: random)")
	fs.StringVar(&cfg.OutputDir, "output", "", "Output directory (default: stdout)")
	fs.BoolVar(&cfg.Flatten, "flatten", false, "Flatten nested structs to top-level keys")
	fs.IntVar(&cfg.Interval, "interval", 1000, "Collection interval in milliseconds")
	fs.BoolVar(&cfg.DisableVM, "no-vm", false, "Disable VM metrics")
	fs.BoolVar(&cfg.DisableContainer, "no-container", false, "Disable container metrics")
	fs.BoolVar(&cfg.DisableProcess, "no-procs", false, "Disable process metrics")
	fs.BoolVar(&cfg.DisableNvidia, "no-nvidia", false, "Disable NVIDIA GPU metrics")
	fs.BoolVar(&cfg.DisableVLLM, "no-vllm", false, "Disable vLLM metrics")
	fs.BoolVar(&cfg.DisableVLLMHistograms, "no-vllm-hist", false, "Disable vLLM histogram collection")
	fs.StringVar(&cfg.VLLMEndpoint, "vllm-endpoint", DefaultVLLMEndpoint, "vLLM metrics endpoint")
	fs.StringVar(&cfg.Pprof, "pprof", "", "Enable pprof profiling on the given address")
	fs.IntVar(&cfg.ServerPort, "port", 8888, "HTTP port (server mode)")
	fs.BoolVar(&cfg.Debug, "debug", false, "Enable verbose debug logging")
	fs.BoolVar(&cfg.PollStats, "poll-stats", false, "Show poller statistics on exit")

	var disabled string
	fs.StringVar(&disabled, "disabled", "", "Comma-separated list of collectors to disable (vm,container,process,nvidia,vllm,vllm-hist)")

	if err := fs.Parse(args); err != nil {
		log.Fatalf("Failed to parse args: %v", err)
	}

	applyEnv(fs)

	if cfg.Debug {
		SetDebug(true)
	}

	if cfg.Interval <= 0 {
		log.Fatalf("Invalid interval: %d", cfg.Interval)
	}

	applyDisabled(disabled, cfg)

	Debugf("config: mode=%s uuid=%s interval=%dms output=%q flatten=%v port=%d poll-stats=%v",
		cfg.Mode, cfg.UUID, cfg.Interval, cfg.OutputDir, cfg.Flatten, cfg.ServerPort, cfg.PollStats)
	Debugf("config: disabled vm=%v container=%v process=%v nvidia=%v vllm=%v vllm-hist=%v",
		cfg.DisableVM, cfg.DisableContainer, cfg.DisableProcess,
		cfg.DisableNvidia, cfg.DisableVLLM, cfg.DisableVLLMHistograms)
	Debugf("config: vllm-endpoint=%s pprof=%q", cfg.VLLMEndpoint, cfg.Pprof)

	return cfg
}

func applyEnv(fs *flag.FlagSet) {
	set := make(map[string]bool)
	fs.Visit(func(f *flag.Flag) { set[f.Name] = true })

	fs.VisitAll(func(f *flag.Flag) {
		if set[f.Name] {
			return
		}
		env := "INFPRO_" + strings.ToUpper(strings.ReplaceAll(f.Name, "-", "_"))
		if val, ok := os.LookupEnv(env); ok {
			if err := f.Value.Set(val); err != nil {
				log.Fatalf("env %s=%q: %v", env, val, err)
			}
		}
	})
}

func applyDisabled(disabled string, cfg *Config) {
	if disabled == "" {
		return
	}
	lookup := map[string]*bool{
		"vm":        &cfg.DisableVM,
		"container": &cfg.DisableContainer,
		"process":   &cfg.DisableProcess,
		"nvidia":    &cfg.DisableNvidia,
		"vllm":      &cfg.DisableVLLM,
		"vllm-hist": &cfg.DisableVLLMHistograms,
	}
	for _, name := range strings.Split(disabled, ",") {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		ptr, ok := lookup[name]
		if !ok {
			log.Fatalf("Invalid disabled collector: %s", name)
		}
		*ptr = true
	}
}

func parseMode(args []string) (string, []string) {
	if len(args) == 0 || strings.HasPrefix(args[0], "-") {
		return DefaultMode, args
	}
	switch strings.ToLower(args[0]) {
	case "c", "continuous":
		return "continuous", args[1:]
	case "s", "snapshot":
		return "snapshot", args[1:]
	case "ser", "server":
		return "server", args[1:]
	default:
		log.Fatalf("Unknown command: %q (must be continuous|c, snapshot|s, server|ser)", args[0])
		return "", nil
	}
}

func GenerateUUID() string {
	id, err := uuid.NewRandom()
	if err != nil {
		log.Fatal(fmt.Errorf("failed to generate UUID: %w", err))
	}
	return id.String()
}
