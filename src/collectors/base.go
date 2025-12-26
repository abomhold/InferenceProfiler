package collectors

import (
	"bufio"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

const JiffiesPerSecond = 100

// GetTimestamp returns current time in milliseconds since epoch
func GetTimestamp() int64 {
	return time.Now().UnixNano()
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

// ProbeFunction executes a function and returns result with timestamp
func ProbeFunction[T any](fn func() (T, error), defaultVal T) (T, int64) {
	ts := GetTimestamp()
	result, err := fn()
	if err != nil {
		log.Printf("ProbeFunction error: %v", err)
		return defaultVal, ts
	}
	return result, ts
}

// ProbeFileInt reads a file and converts content to int64
func ProbeFileInt(path string) (int64, int64) {
	content, ts := ProbeFile(path)
	if content == "" {
		return 0, ts
	}
	val, err := strconv.ParseInt(content, 10, 64)
	if err != nil {
		return 0, ts
	}
	return val, ts
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

// ParseProcKV parses a key-value file like /proc/meminfo
func ParseProcKV(path string, separator string) (map[string]string, int64) {
	lines, ts := ProbeFileLines(path)
	data := make(map[string]string)
	for _, line := range lines {
		if idx := strings.Index(line, separator); idx != -1 {
			key := strings.TrimSpace(line[:idx])
			val := strings.TrimSpace(line[idx+1:])
			data[key] = val
		}
	}
	return data, ts
}
