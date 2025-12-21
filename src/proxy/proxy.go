package proxy

import (
	"encoding/base64"
	"fmt"
	"log"
	api "mixproxy/src/api/admin"
	certificate "mixproxy/src/certs"
	"mixproxy/src/proxy/config"
	"mixproxy/src/proxy/tools"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/proxy"
)

type LoadBalancer struct {
	URL      url.URL
	Capacity float64
}

func Stop() {
	if err := config.SERVERS["HTTP"].Shutdown(); err != nil {
		log.Println("Error stopping HTTP:", err)
	}
	if err := config.SERVERS["HTTPS"].Shutdown(); err != nil {
		log.Println("Error stopping HTTPS:", err)
	}
}

func Control(action string) {
	switch action {
	case "start":
		Start()
	case "stop":
		Stop()
	case "reload":
		Stop()
		time.Sleep(10 * time.Second)
		Start()
	case "createCertificates":
		os.RemoveAll("./certs")
		createCertificates()
	default:
		fmt.Println("Invalid action")
	}
}

func createCertificates() {
	cfg, _ := config.ReadConfig()

	if _, err := os.Stat("./certs/wildcard.developer.space.crt"); err == nil {
		return
	}

	host := cfg.Hostname

	DNSnames := []string{host, cfg.SubdomainAdminPanel + "." + host, "admin-api." + host, "localhost", "127.0.0.1", "*.localhost"}

	for _, server := range cfg.LoadBalancer {
		if server.Subdomain != "" {
			DNSnames = append(DNSnames, server.Subdomain+"."+host)
		}
	}

	certificate.Create(DNSnames)
}

func startHttpAndHttpsServer(wg *sync.WaitGroup) {
	defer wg.Done()

	// Configurar proxy reverso para HTTPS
	config.SERVERS["HTTPS"].All("/*", func(c *fiber.Ctx) error {
		url, err := getHandleFunc(c)
		if err != nil {
			return err
		}

		if strings.Contains(url, "admin") {
			auth := c.Get("Authorization")
			if auth == "" || !strings.HasPrefix(auth, "Basic ") {
				c.Status(401).Set("WWW-Authenticate", `Basic realm="Admin"`)
				return c.SendString("Unauthorized")
			}
			encoded := strings.TrimPrefix(auth, "Basic ")
			decoded, err := base64.StdEncoding.DecodeString(encoded)
			if err != nil {
				c.Status(401)
				return c.SendString("Unauthorized")
			}
			creds := string(decoded)
			parts := strings.SplitN(creds, ":", 2)
			if len(parts) != 2 || parts[0] != "admin" || parts[1] != "password" {
				c.Status(401)
				return c.SendString("Unauthorized")
			}
		}

		c.Request().Header.Set("Host", c.Hostname())

		if err := proxy.Do(c, url+c.OriginalURL()); err != nil {
			return err
		}

		c.Response().Header.Del(fiber.HeaderServer)

		return nil
	})

	// Redirigir todas las peticiones HTTP a HTTPS
	config.SERVERS["HTTP"].All("/*", func(c *fiber.Ctx) error {
		host := string(c.Request().Header.Host())
		url := "https://" + host + c.OriginalURL()
		return c.Redirect(url, fiber.StatusMovedPermanently)
	})

	// Iniciar servidor HTTPS en puerto 443 con certificados wildcard
	go func() {
		log.Println("✅ Servidor HTTPS iniciado en puerto 443")
		crt, key := getTlsConfig()

		if err := config.SERVERS["HTTPS"].ListenTLS(":443", crt, key); err != nil {
			log.Fatalf("❌ Error HTTPS: %v", err)
		}
	}()

	// Iniciar servidor HTTP en puerto 80 (redirige a HTTPS)
	log.Println("🔄 Servidor HTTP iniciado en puerto 80 (redirige a HTTPS)")
	if err := config.SERVERS["HTTP"].Listen(":80"); err != nil {
		log.Fatalf("❌ Error HTTP: %v", err)
	}

}

func Start() {
	api.SetControlFunc(Control)

	config.Proxies = make(map[string][]string)
	tools.ServerSelected = make(map[string]*tools.ServerEntry)

	// Configurar backends
	cfg, _ := config.ReadConfig()

	if err := config.ValidateConfig(cfg); err != nil {
		fmt.Println("❌ Error de validación:", err)
		os.Exit(0)
	}

	loadBalancer := map[string]*[]LoadBalancer{}

	if cfg.ModeDeveloper {
		fmt.Println("Configuring certificates in development mode")
		createCertificates()
	}

	for _, e := range cfg.LoadBalancer {
		vps := []LoadBalancer{}
		subdomain := e.Subdomain

		for _, v := range e.VPS {
			vps = append(vps, LoadBalancer{
				URL:      *tools.MustParseURL(v.IP),
				Capacity: v.Capacity,
			})
		}

		loadBalancer[subdomain] = &vps
		probability := []tools.VpsProbability{}
		for _, v := range e.VPS {
			probability = append(probability, tools.VpsProbability{
				Probability: v.Capacity,
				IP:          v.IP,
			})
		}
		tools.SetupServerSelected(subdomain, probability)
	}

	if cfg.RootLoadBalancer != nil && config.AllValuesNonEmpty(cfg.RootLoadBalancer) {
		vps := []LoadBalancer{}
		subdomain := ""

		for _, v := range cfg.RootLoadBalancer.VPS {
			vps = append(vps, LoadBalancer{
				URL:      *tools.MustParseURL(v.IP),
				Capacity: v.Capacity,
			})
		}

		loadBalancer[subdomain] = &vps
		probability := []tools.VpsProbability{}
		for _, v := range cfg.RootLoadBalancer.VPS {
			probability = append(probability, tools.VpsProbability{
				Probability: v.Capacity,
				IP:          v.IP,
			})
		}
		tools.SetupServerSelected(subdomain, probability)
	}

	if len(loadBalancer) != 0 {
		for subdomain, targets := range loadBalancer {
			for _, target := range *targets {
				config.Proxies[subdomain] = append(config.Proxies[subdomain], target.URL.String())
			}
		}
	}
	// Redirigir HTTP a HTTPS
	httpServerExitDone := &sync.WaitGroup{}
	httpServerExitDone.Add(1)

	startHttpAndHttpsServer(httpServerExitDone)

	httpServerExitDone.Wait()
}
