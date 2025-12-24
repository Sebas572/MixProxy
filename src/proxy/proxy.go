package proxy

import (
	"fmt"
	"log"
	api "mixproxy/src/api/admin"
	certificate "mixproxy/src/certs"
	"mixproxy/src/proxy/config"
	"mixproxy/src/redis"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"github.com/valyala/fasthttp"
)

func Stop() {
	log.Println("Stop server")

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
		reloadConfig()
	case "createCertificates":
		createCertificates()
	default:
		fmt.Println("Invalid action")
	}
}

func createCertificates() {
	os.RemoveAll("./certs")
	cfg, _ := config.ReadConfig()

	host := cfg.Hostname

	DNSnames := []string{host, cfg.SubdomainAdminPanel + "." + host, "admin-api." + host, "*." + host, "localhost", "127.0.0.1"}

	for _, server := range cfg.LoadBalancer {
		if server.Subdomain != "" {
			DNSnames = append(DNSnames, server.Subdomain+"."+host)
		}
	}

	certificate.Create(DNSnames)
}

var client *fasthttp.Client = &fasthttp.Client{
	MaxConnsPerHost:     1000,
	MaxIdleConnDuration: 90 * time.Second,
}

func startHttpAndHttpsServer(wg *sync.WaitGroup) {
	defer wg.Done()

	api.HandleAdminAPI()

	// Configurar proxy reverso para HTTPS y WSS
	config.SERVERS["HTTPS"].All("/*", func(c *fiber.Ctx) error {
		return handleHTTPS(c)
	})

	config.SERVERS["HTTPS"].Use("/*", websocket.New(func(c *websocket.Conn) {
		handleWebSocket(c)
	}))

	// Redirigir todas las peticiones HTTP a HTTPS
	config.SERVERS["HTTP"].All("/*", func(c *fiber.Ctx) error {
		host := string(c.Request().Header.Host())
		url := "https://" + host + c.OriginalURL()
		return c.Redirect(url, fiber.StatusMovedPermanently)
	})

	// Iniciar servidor HTTPS en puerto 443 con certificados wildcard
	go func() {
		log.Println("✅ Servidor HTTPS iniciado en puerto 443")
		crt, key := getCertificateConfig()

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

func generateCacheKey(c *fiber.Ctx) string {
	// Key: method:url/path:accept
	accept := c.Get("Accept")
	_, host := getSubdomainAndHost(c)

	return c.Method() + ":" + host + c.OriginalURL() + ":" + accept
}

func pathMatches(pattern, path string) bool {
	if pattern == "/*" {
		return true
	}

	if strings.HasSuffix(pattern, "/*") {
		prefix := strings.TrimSuffix(pattern, "/*")
		return strings.HasPrefix(path, prefix)
	}
	return pattern == path
}

func isCacheable(c *fiber.Ctx) bool {
	subdomain := getSubdomain(c)
	path := c.OriginalURL()

	// Check if subdomain allows cache
	if !redis.DoesTheSubdomainAllowCache(subdomain) {
		return false
	}

	// Get cache paths for subdomain
	paths, err := redis.GetCachePaths(subdomain)
	if err != nil {
		return false
	}

	pathMatchesConfig := false
	for _, p := range paths {
		if pathMatches(p, path) {
			pathMatchesConfig = true
			break
		}
	}

	if !pathMatchesConfig {
		return false
	}

	// Check request Cache-Control
	cc := c.Get("Cache-Control")
	if strings.Contains(cc, "no-cache") || strings.Contains(cc, "private") {
		return false
	}

	// Don't cache errors
	if c.Response().StatusCode() >= 400 {
		return false
	}

	return true
}

func Start() {
	log.Println("Start server")
	api.SetControlFunc(Control)

	reloadConfig()

	// Redirigir HTTP a HTTPS
	httpServerExitDone := &sync.WaitGroup{}
	httpServerExitDone.Add(1)

	startHttpAndHttpsServer(httpServerExitDone)

	httpServerExitDone.Wait()
}
