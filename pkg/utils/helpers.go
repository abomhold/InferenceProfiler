package utils

import (
	"os"
	"strconv"
	"strings"
	"time"
)

func GetTimestamp() int64 {
	return time.Now().UnixNano()
}

func ParseInt64(s string) int64 {
	v, err := strconv.ParseInt(strings.TrimSpace(s), 10, 64)
	if err != nil {
		Debugf("ParseInt64: failed to parse %q: %v", s, err)
	}
	return v
}

func ParseInt64Bytes(b []byte) int64 {
	v, err := strconv.ParseInt(string(b), 10, 64)
	if err != nil {
		Debugf("ParseInt64Bytes: failed to parse %q: %v", string(b), err)
	}
	return v
}

func ParseFloat64(s string) float64 {
	v, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
	if err != nil {
		Debugf("ParseFloat64: failed to parse %q: %v", s, err)
	}
	return v
}

func ByteSliceToString[T int8 | uint8](b []T) string {
	buf := make([]byte, 0, len(b))
	for _, c := range b {
		if c == 0 {
			break
		}
		buf = append(buf, byte(c))
	}
	return string(buf)
}

func File(path string) (string, int64, error) {
	ts := GetTimestamp()
	data, err := os.ReadFile(path)
	if err != nil {
		return "", ts, err
	}
	return strings.TrimSpace(string(data)), ts, nil
}

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

func FileLines(path string) ([]string, int64, error) {
	ts := GetTimestamp()
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, ts, err
	}
	return strings.Split(string(data), "\n"), ts, nil
}

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

func Exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func IsDir(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}
