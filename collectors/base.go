// Package collector provides system metric collection from Linux procfs, sysfs, cgroups, and NVIDIA GPUs.
package collectors

import (
	"bufio"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

const jiffiesPerSec = 100

func TimedAt[T any](v T, t int64) Timed[T] {
	return Timed[T]{Value: v, Time: t}
}
func NowMilli() int64 {
	return time.Now().UnixMilli()
}

type BaseCollector struct{}

func (b *BaseCollector) ReadFile(path string) (string, int64) {
	ts := NowMilli()
	data, err := os.ReadFile(path)
	if err != nil {
		return "", ts
	}
	return strings.TrimSpace(string(data)), ts
}

func (b *BaseCollector) ReadInt(path string) (int64, int64) {
	content, ts := b.ReadFile(path)
	if content == "" {
		return 0, ts
	}
	v, err := strconv.ParseInt(content, 10, 64)
	if err != nil {
		return 0, ts
	}
	return v, ts
}

func (b *BaseCollector) ReadLines(path string) ([]string, int64) {
	ts := NowMilli()
	f, err := os.Open(path)
	if err != nil {
		return nil, ts
	}
	defer f.Close()

	var lines []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, ts
}

func (b *BaseCollector) ParseKV(path string, sep byte) (map[string]string, int64) {
	lines, ts := b.ReadLines(path)
	m := make(map[string]string)
	for _, line := range lines {
		idx := strings.IndexByte(line, sep)
		if idx < 0 {
			continue
		}
		k := strings.TrimSpace(line[:idx])
		v := strings.TrimSpace(line[idx+1:])
		m[k] = v
	}
	return m, ts
}

func (b *BaseCollector) ParseInt(s string) int64 {
	v, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		log.Fatalln("Failed to parse int", s)
	}
	return v
}

func (b *BaseCollector) ParseMemValue(s string) int64 {
	s = strings.TrimSuffix(s, " kB")
	v, _ := strconv.ParseInt(s, 10, 64)
	return v * 1024
}

func (b *BaseCollector) ParseSize(s string) int64 {
	s = strings.TrimSpace(s)
	mult := int64(1)
	if strings.HasSuffix(s, "K") {
		mult = 1024
		s = s[:len(s)-1]
	} else if strings.HasSuffix(s, "M") {
		mult = 1024 * 1024
		s = s[:len(s)-1]
	} else if strings.HasSuffix(s, "G") {
		mult = 1024 * 1024 * 1024
		s = s[:len(s)-1]
	}
	v, _ := strconv.ParseInt(s, 10, 64)
	return v * mult
}
