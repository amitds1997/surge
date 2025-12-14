package utils

import (
	"fmt"
	"os"
	"sync"
)

var (
	debugFile *os.File
	debugOnce sync.Once
)

// Debug writes a message to debug.log file
func Debug(format string, args ...any) {
	debugOnce.Do(func() {
		debugFile, _ = os.Create("debug.log")
	})
	if debugFile != nil {
		fmt.Fprintf(debugFile, format+"\n", args...)
		debugFile.Sync() // Flush immediately
	}
}
