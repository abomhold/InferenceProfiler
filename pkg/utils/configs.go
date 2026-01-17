package utils

type Config struct {
	// Collection toggles
	EnableVM        bool
	EnableContainer bool
	EnableProcess   bool
	EnableNvidia    bool
	EnableVLLM      bool

	// GPU-specific options
	CollectGPUProcesses bool

	// Performance options
	Concurrent bool

	// Output options
	OutputFormat string
	OutputFile   string
	Flatten      bool

	// Streaming options
	EnableStreaming bool
	StreamInterval  int // milliseconds

	// Metadata
	UUID     string
	Hostname string
}

func NewConfig() *Config {
	return &Config{
		// Defaults
		EnableVM:        true,
		EnableContainer: false,
		EnableProcess:   false,
		EnableNvidia:    false,
		EnableVLLM:      false,

		CollectGPUProcesses: false,
		Concurrent:          false,

		OutputFormat: "json",
		OutputFile:   "",
		Flatten:      false,

		EnableStreaming: false,
		StreamInterval:  1000,

		UUID:     "",
		Hostname: "",
	}
}
