package collectors

import (
	"bufio"
	"os"
	"strconv"
	"strings"
	"time"
)

// =============================================================================
// Timestamp Utilities
// =============================================================================

// GetTimestamp returns current time in nanoseconds since Unix epoch
func GetTimestamp() int64 {
	return time.Now().UnixNano()
}

// =============================================================================
// Generic File Probing
// =============================================================================

// Probe reads a file and transforms its contents using the provided function.
// Returns the transformed value and the timestamp when the file was read.
// On error, returns the zero value and the timestamp.
func Probe[T any](path string, transform func(string) T) (T, int64) {
	ts := GetTimestamp()
	data, err := os.ReadFile(path)
	if err != nil {
		var zero T
		return zero, ts
	}
	return transform(strings.TrimSpace(string(data))), ts
}

// =============================================================================
// Common Transformers for Probe
// =============================================================================

// Identity returns the string unchanged
func Identity(s string) string { return s }

// ParseInt64 converts string to int64, returns 0 on failure
func ParseInt64(s string) int64 {
	v, _ := strconv.ParseInt(strings.TrimSpace(s), 10, 64)
	return v
}

// ParseFloat64 converts string to float64, returns 0 on failure
func ParseFloat64(s string) float64 {
	v, _ := strconv.ParseFloat(strings.TrimSpace(s), 64)
	return v
}

// SplitLines splits content into lines
func SplitLines(s string) []string {
	if s == "" {
		return nil
	}
	return strings.Split(s, "\n")
}

// =============================================================================
// Convenience Functions (wrap Probe with common transformers)
// =============================================================================

// ProbeFile reads a file and returns its content as a trimmed string with timestamp
func ProbeFile(path string) (string, int64) {
	return Probe(path, Identity)
}

// ProbeFileInt reads a file and converts content to int64
func ProbeFileInt(path string) (int64, int64) {
	return Probe(path, ParseInt64)
}

// ProbeFileLines reads a file and returns lines with timestamp
// More efficient than Probe for large files as it doesn't load entire file into memory
func ProbeFileLines(path string) ([]string, int64) {
	ts := GetTimestamp()
	file, err := os.Open(path)
	if err != nil {
		return nil, ts
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, ts
}

// ProbeFileKV parses a key-value file (like /proc/meminfo or /proc/self/status)
// Returns a map of trimmed keys to trimmed values
func ProbeFileKV(path string, separator string) (map[string]string, int64) {
	lines, ts := ProbeFileLines(path)
	data := make(map[string]string, len(lines))
	for _, line := range lines {
		if before, after, found := strings.Cut(line, separator); found {
			data[strings.TrimSpace(before)] = strings.TrimSpace(after)
		}
	}
	return data, ts
}

// =============================================================================
// Parsing Helpers
// =============================================================================

// parseInt64 is a quiet int64 parser for cleaner metric collection code
// Returns 0 on parse failure
func parseInt64(s string) int64 {
	return ParseInt64(s)
}

// parseFloat64 is a quiet float64 parser
// Returns 0 on parse failure
func parseFloat64(s string) float64 {
	return ParseFloat64(s)
}
