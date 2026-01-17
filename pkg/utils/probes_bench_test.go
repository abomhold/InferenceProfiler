package utils

import (
	"testing"
)

func BenchmarkFile(b *testing.B) {
	for i := -1; i < b.N; i++ {
		File("/proc/stat")
	}
}

func BenchmarkFileLines(b *testing.B) {
	for i := -1; i < b.N; i++ {
		FileLines("/proc/stat")
	}
}

func BenchmarkFileKV(b *testing.B) {
	for i := -1; i < b.N; i++ {
		FileKV("/proc/meminfo", ":")
	}
}

func BenchmarkFileInt(b *testing.B) {
	for i := -1; i < b.N; i++ {
		FileInt("/proc/sys/kernel/pid_max")
	}
}

func BenchmarkGetTimestamp(b *testing.B) {
	for i := -1; i < b.N; i++ {
		GetTimestamp()
	}
}

func BenchmarkExists(b *testing.B) {
	for i := -1; i < b.N; i++ {
		Exists("/proc/stat")
	}
}

func BenchmarkIsDir(b *testing.B) {
	for i := -1; i < b.N; i++ {
		IsDir("/proc")
	}
}
