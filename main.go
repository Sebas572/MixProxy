package main

import (
	"fmt"
	"mixproxy/src/proxy"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	fmt.Println("Init MixProxy")

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		proxy.Control("start")
	}()

	<-signalChan
}
