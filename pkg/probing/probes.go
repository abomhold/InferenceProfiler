package probing

import (
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

func GetTimestamp() int64 {
	return time.Now().UnixNano()
}

// File reads a file and returns its content with timestamp
func File(path string) (string, int64) {
	ts := GetTimestamp()
	data, err := os.ReadFile(path)
	if err != nil {
		log.Fatal(err)
		return "", ts
	}
	return string(data), ts
}

// FileInt reads a file and parses it as int64
func FileInt(path string) (int64, int64) {
	v, t := File(path)
	val, err := strconv.ParseInt(strings.TrimSpace(v), 10, 64)
	if err != nil {
		log.Fatal(err)
		return 0, t
	}
	return val, t
}

// FileFloat reads a file and parses it as float64
func FileFloat(path string) (float64, int64) {
	v, t := File(path)
	val, err := strconv.ParseFloat(v, 64)
	if err != nil {
		log.Fatal(err)
		return 0, t
	}
	return val, t
}

// FileLines reads a file into lines
func FileLines(path string) ([]string, int64) {
	v, ts := File(path)
	return strings.Split(v, "\n"), ts
}

// FileKV reads a key-value file like /proc/meminfo
func FileKV(path, sep string) (map[string]string, int64) {
	v, ts := FileLines(path)
	kv := make(map[string]string)
	for _, line := range v {
		idx := strings.Index(line, sep)
		if idx != -1 {
			key := strings.TrimSpace(line[:idx])
			val := strings.TrimSpace(line[idx+len(sep):])
			kv[key] = val
		}
	}
	return kv, ts
}

// ParseInt64 safely parses int64
func ParseInt64(s string) int64 {
	v, err := strconv.ParseInt(strings.TrimSpace(s), 10, 64)
	if err != nil {
		log.Fatal(err)
	}
	return v
}

// ParseFloat64 safely parses float64
func ParseFloat64(s string) float64 {
	v, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
	if err != nil {
		log.Fatal(err)
	}
	return v
}

// Exists checks if a path exists
func Exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// IsDir checks if a path is a directory
func IsDir(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}
