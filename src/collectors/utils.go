package collectors

import (
	"bufio"
	"os"
	"strconv"
	"strings"
	"time"
)

// Constants
const JiffiesPerSecond = 100

// GetTimestamp returns current time in nanoseconds
func GetTimestamp() int64 {
	return time.Now().UnixNano()
}

// parseInt64 is a quiet parser for DRYer metric collection
func parseInt64(s string) int64 {
	v, _ := strconv.ParseInt(strings.TrimSpace(s), 10, 64)
	return v
}

// parseFloat64 is a quiet parser for float values
func parseFloat64(s string) float64 {
	v, _ := strconv.ParseFloat(strings.TrimSpace(s), 64)
	return v
}

// ProbeFile reads a file and returns its content with timestamp
func ProbeFile(path string) (string, int64) {
	ts := GetTimestamp()
	data, err := os.ReadFile(path)
	if err != nil {
		return "", ts
	}
	return strings.TrimSpace(string(data)), ts
}

// ProbeFileInt reads a file and converts content to int64
func ProbeFileInt(path string) (int64, int64) {
	content, ts := ProbeFile(path)
	if content == "" {
		return 0, ts
	}
	return parseInt64(content), ts
}

// ProbeFileLines reads a file and returns lines
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

// ParseProcKV parses a key-value file like /proc/meminfo or /proc/self/status
func ParseProcKV(path string, separator string) (map[string]string, int64) {
	lines, ts := ProbeFileLines(path)
	data := make(map[string]string)
	for _, line := range lines {
		if before, after, found := strings.Cut(line, separator); found {
			data[strings.TrimSpace(before)] = strings.TrimSpace(after)
		}
	}
	return data, ts
}

// ByteSliceToString converts a null-terminated byte array to string
func ByteSliceToString(b []int8) string {
	n := 0
	for n < len(b) && b[n] != 0 {
		n++
	}
	s := make([]byte, n)
	for i := 0; i < n; i++ {
		s[i] = byte(b[i])
	}
	return string(s)
}
