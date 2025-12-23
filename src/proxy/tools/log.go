package tools

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

var (
	currentFile *os.File
	currentDate string
)

func PrintLog(method, path, ip, host string) {
	const (
		white   = "\033[97m"
		cyan    = "\033[36m"
		yellow  = "\033[33m"
		green   = "\033[32m"
		red     = "\033[31m"
		gray    = "\033[90m"
		reset   = "\033[0m"
		magenta = "\033[35m"
	)

	now := time.Now()
	dateStr := now.Format("2006-01-02")
	timeStr := now.Format("02/Jan/2006:15:04:05")

	if dateStr != currentDate {
		if currentFile != nil {
			currentFile.Close()
		}
		logPath := filepath.Join("logs", "log-"+dateStr)
		var err error
		currentFile, err = os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error opening log file: %v\n", err)
			return
		}
		currentDate = dateStr
	}

	plainLog := fmt.Sprintf("[%s] %s -> %s (%s) %s\n", timeStr, ip, host, method, path)
	if currentFile != nil {
		currentFile.WriteString(plainLog)
	}

	fmt.Printf("%s[%s]%s %s%s%s -> %s%s%s (%s%s%s) %s%s%s\n",
		gray, timeStr, reset,
		cyan, ip, reset,
		magenta, host, reset,
		gray, method, reset,
		white, path, reset)
}
