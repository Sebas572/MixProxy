package core

import (
	"context"
	"fmt"
	"mixproxy/src/proxy"
	"mixproxy/src/proxy/config"
	"time"
)

func Control(action string) {
	fmt.Printf("Control -> %s\n", action)

	switch action {
	case "start":
		proxy.Start()
	case "stop":
		if err := config.SERVERS["HTTP"].Shutdown(context.TODO()); err != nil {
			panic(err)
		}
		if err := config.SERVERS["HTTPS"].Shutdown(context.TODO()); err != nil {
			panic(err)
		}
	case "reload":
		Control("stop")
		time.Sleep(10 * time.Second)
		Control("start")
	}
}
