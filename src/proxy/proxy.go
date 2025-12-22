package proxy

import (
	"encoding/base64"
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
	"github.com/gofiber/fiber/v2/middleware/proxy"
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
		os.RemoveAll("./certs")
		createCertificates()
	default:
		fmt.Println("Invalid action")
	}
}

func createCertificates() {
	cfg, _ := config.ReadConfig()

	if _, err := os.Stat("./certs/wildcard.crt"); err == nil {
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

var client *fasthttp.Client = &fasthttp.Client{
	MaxConnsPerHost:     1000,
	MaxIdleConnDuration: 90 * time.Second,
}

func startHttpAndHttpsServer(wg *sync.WaitGroup) {
	defer wg.Done()

	api.HandleAdminAPI()

	// Configurar proxy reverso para HTTPS
	config.SERVERS["HTTPS"].All("/*", func(c *fiber.Ctx) error {
		// subdomain, host := getSubdomainAndHost(c)

		if c.Method() == "GET" {
			// Check cache for non-admin GET requests
			key := generateCacheKey(c)
			cached, found, err := redis.GetCachedResponse(key)
			if err != nil {
				log.Printf("Redis error: %v", err)
			} else if found {
				c.Status(cached.Status)
				for k, v := range cached.Headers {
					c.Set(k, v)
				}
				// Set Server header for cached response
				subdomain, _ := getSubdomainAndHost(c)
				if redis.DoesTheSubdomainAllowCache(subdomain) {
					c.Set(fiber.HeaderServer, "Mixproxy (with cache)")
				} else {
					c.Set(fiber.HeaderServer, "Mixproxy")
				}
				return c.SendString(cached.Body)
			}
		}

		url, err := getHandleFunc(c)
		if err != nil {
			return err
		}

		// go redis.AddRequestLog(c.Method(), host+c.OriginalURL(), c.IP(), subdomain, c.Response().StatusCode())
		// config.AddRequestLog(c.Method(), c.OriginalURL(), c.IP(), getSubdomain(c), c.Response().StatusCode())

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

		// c.Request().Header.Set("Host", c.Hostname())

		if err := proxy.Do(c, url+c.OriginalURL(), client); err != nil {
			return err
		}

		// Set Server header
		if redis.DoesTheSubdomainAllowCache(getSubdomain(c)) {
			c.Set(fiber.HeaderServer, "Mixproxy (with cache)")
		} else {
			c.Set(fiber.HeaderServer, "Mixproxy")
		}

		// Cache the response if GET, not admin and cacheable
		if c.Method() == "GET" && !strings.Contains(url, "admin") && isCacheable(c) {
			key := generateCacheKey(c)
			resp := redis.CachedResponse{
				Status:  c.Response().StatusCode(),
				Headers: make(map[string]string),
				Body:    string(c.Response().Body()),
			}
			c.Response().Header.VisitAll(func(key, value []byte) {
				resp.Headers[string(key)] = string(value)
			})
			ttl := 15 * time.Minute
			if err := redis.SetCachedResponse(key, resp, ttl); err != nil {
				log.Printf("Failed to cache response: %v", err)
			}
		}

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
