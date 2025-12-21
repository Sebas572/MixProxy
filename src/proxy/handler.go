package proxy

import (
	"log"
	"mixproxy/src/proxy/config"
	"mixproxy/src/proxy/tools"
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
)

var cfg *config.Config
var redirect map[string][]config.VPSEntry = make(map[string][]config.VPSEntry)

func init() {
	config, err := config.ReadConfig()
	if err != nil {
		log.Println("Not found config")

		os.Exit(0)
	}

	cfg = config

	for _, load := range cfg.LoadBalancer {
		redirect[load.Subdomain] = load.VPS
	}
}

func getHandleFunc(ctx *fiber.Ctx) (string, error) {
	hostAndPort := string(ctx.BaseURL())
	host := strings.Split(hostAndPort, "//")[1]
	subdomain := ""

	if host != cfg.Hostname {
		subdomain = strings.Split(host, ".")[0]
	}

	ip := ctx.IP()

	tools.PrintLog("GET", ctx.OriginalURL(), ip, host)

	if subdomain == cfg.SubdomainAdminPanel {
		return "http://admin:4173", nil
	}

	target, err := tools.GetTargetIPForSubdomain(subdomain)
	if err != nil {
		return config.URL_ADMIN_PANEL, err
	}

	return target, nil
}
