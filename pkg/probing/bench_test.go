package probing

import (
	"testing"
)

func BenchmarkFile(b *testing.B) {
	for i := 0; i < b.N; i++ {
		File("/proc/stat")
	}
}

func BenchmarkFileLines(b *testing.B) {
	for i := 0; i < b.N; i++ {
		FileLines("/proc/stat")
	}
}

func BenchmarkFileKV(b *testing.B) {
	for i := 0; i < b.N; i++ {
		FileKV("/proc/meminfo", ":")
	}
}

func BenchmarkFileInt(b *testing.B) {
	for i := 0; i < b.N; i++ {
		FileInt("/proc/sys/kernel/pid_max")
	}
}

func BenchmarkGetTimestamp(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GetTimestamp()
	}
}

func BenchmarkExists(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Exists("/proc/stat")
	}
}

func BenchmarkIsDir(b *testing.B) {
	for i := 0; i < b.N; i++ {
		IsDir("/proc")
	}
}

func BenchmarkParseInt64(b *testing.B) {
	s := "123456789"
	for i := 0; i < b.N; i++ {
		ParseInt64(s)
	}
}

func BenchmarkParseFloat64(b *testing.B) {
	s := "123.456789"
	for i := 0; i < b.N; i++ {
		ParseFloat64(s)
	}
}
