// Package probing provides utilities for reading system files.
package probing

import (
	"os"
	"strconv"
	"strings"
	"time"
)

// GetTimestamp returns the current time in nanoseconds.
func GetTimestamp() int64 {
	return time.Now().UnixNano()
}

// File reads a file and returns its trimmed content with timestamp.
func File(path string) (string, int64, error) {
	ts := GetTimestamp()
	data, err := os.ReadFile(path)
	if err != nil {
		return "", ts, err
	}
	return strings.TrimSpace(string(data)), ts, nil
}

// FileOrEmpty reads a file and returns empty string on error.
func FileOrEmpty(path string) (string, int64) {
	v, ts, _ := File(path)
	return v, ts
}

// FileInt reads a file and parses it as int64.
func FileInt(path string) (int64, int64, error) {
	v, ts, err := File(path)
	if err != nil {
		return 0, ts, err
	}
	val, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return 0, ts, err
	}
	return val, ts, nil
}

// FileIntOrZero reads a file as int64, returns 0 on error.
func FileIntOrZero(path string) (int64, int64) {
	v, ts, _ := FileInt(path)
	return v, ts
}

// FileLines reads a file and splits by newline.
func FileLines(path string) ([]string, int64, error) {
	ts := GetTimestamp()
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, ts, err
	}
	return strings.Split(string(data), "\n"), ts, nil
}

// FileLinesOrEmpty reads a file into lines, returns empty slice on error.
func FileLinesOrEmpty(path string) ([]string, int64) {
	lines, ts, _ := FileLines(path)
	return lines, ts
}

// FileKV reads a key-value file like /proc/meminfo.
func FileKV(path, sep string) (map[string]string, int64, error) {
	lines, ts, err := FileLines(path)
	if err != nil {
		return nil, ts, err
	}
	kv := make(map[string]string)
	for _, line := range lines {
		idx := strings.Index(line, sep)
		if idx != -1 {
			key := strings.TrimSpace(line[:idx])
			val := strings.TrimSpace(line[idx+len(sep):])
			kv[key] = val
		}
	}
	return kv, ts, nil
}

// FileKVOrEmpty reads a key-value file, returns empty map on error.
func FileKVOrEmpty(path, sep string) (map[string]string, int64) {
	kv, ts, _ := FileKV(path, sep)
	if kv == nil {
		kv = make(map[string]string)
	}
	return kv, ts
}

// ParseInt64 safely parses int64, returns 0 on error.
func ParseInt64(s string) int64 {
	v, _ := strconv.ParseInt(strings.TrimSpace(s), 10, 64)
	return v
}

// ParseFloat64 safely parses float64, returns 0 on error.
func ParseFloat64(s string) float64 {
	v, _ := strconv.ParseFloat(strings.TrimSpace(s), 64)
	return v
}

// Exists checks if a path exists.
func Exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// IsDir checks if a path is a directory.
func IsDir(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}
