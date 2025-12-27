package proxy

import (
	"encoding/base64"
	"log"
	"mixproxy/src/logger"
	"mixproxy/src/redis"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/proxy"
	"github.com/gofiber/websocket/v2"
)

func handleHTTPS(c *fiber.Ctx) error {
	if websocket.IsWebSocketUpgrade(c) {
		return c.Next()
	}

	subdomain, host := getSubdomainAndHost(c)

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
			logger.AddRequestLog(c.Method(), host+c.OriginalURL(), c.IP(), subdomain, c.Response().StatusCode(), true)
			return c.SendString(cached.Body)
		}
	}

	url, err := getHandleFunc(c)
	if err != nil {
		return err
	}

	logger.AddRequestLog(c.Method(), host+c.OriginalURL(), c.IP(), subdomain, c.Response().StatusCode(), false)

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
}
