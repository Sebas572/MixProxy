package tools

import (
	"fmt"
	"time"
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

	now := time.Now().Format("02/Jan/2006:15:04:05")

	fmt.Printf("%s[%s]%s %s%s%s -> %s%s%s (%s%s%s) %s%s%s\n",
		gray, now, reset,
		cyan, ip, reset,
		magenta, host, reset,
		gray, method, reset,
		white, path, reset)
}
