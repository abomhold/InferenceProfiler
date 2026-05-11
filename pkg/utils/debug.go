package utils

import (
	"fmt"
	"log"
	"os"
	"sync/atomic"
	"time"
)

var debugEnabled atomic.Bool

var debugLogger *log.Logger

func init() {
	debugLogger = log.New(os.Stderr, "[DEBUG] ", log.Ltime|log.Lmicroseconds)
}

func SetDebug(on bool) {
	debugEnabled.Store(on)
	if on {
		log.Println("debug: enabled")
	}
}

func Debugf(format string, args ...any) {
	if !debugEnabled.Load() {
		return
	}
	debugLogger.Output(2, fmt.Sprintf(format, args...))
}

func DebugTimer() time.Time {
	if !debugEnabled.Load() {
		return time.Time{}
	}
	return time.Now()
}

func DebugDuration(component string, operation string, start time.Time) {
	if start.IsZero() || !debugEnabled.Load() {
		return
	}
	debugLogger.Output(2, fmt.Sprintf("%s: %s took %v", component, operation, time.Since(start)))
}
