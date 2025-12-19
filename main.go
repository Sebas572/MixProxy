package main

import (
	"fmt"
	"mixproxy/src/proxy"
	"mixproxy/src/proxy/config"
	"os"
	"os/signal"
	"syscall"

	"github.com/manifoldco/promptui"
)

func main() {
	fmt.Println("Init MixProxy")

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {

		if _, err := os.Stat(config.CONFIG_PATH); err != nil {
			config.CreateConfig()

			os.Exit(0)
		}

		if len(os.Args) >= 2 {
			if os.Args[1] == "--start-proxy" {
				proxy.Control("start")

				return
			}
		}

		prompt := promptui.Select{
			Label: "Select",
			Items: []string{"Start proxy", "Verificate proxy.config.json", "Create certificates SSL (developer)"},
		}

		_, result, err := prompt.Run()

		if err != nil {
			fmt.Printf("Prompt failed %v\n", err)
			return
		}

		switch result {
		case "Start proxy":
			proxy.Control("start")
		case "Verificate proxy.config.json":
			cfg, _ := config.ReadConfig()
			config.ValidateConfig(cfg)
			os.Exit(0)
		case "Create certificates SSL (developer)":
			proxy.Control("createCertificates")
			os.Exit(0)
		}
	}()

	<-signalChan
}
