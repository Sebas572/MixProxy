package main

import (
	"fmt"
	"mixploy/src/proxy"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	fmt.Println("Init Mixploy")

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		proxy.Control("start")
	}()

	<-signalChan
}
