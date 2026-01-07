package api

import (
	"time"

	"mixproxy/src/redis"

	"github.com/gofiber/fiber/v2"
)

func SetupWhitelistRoutes(api fiber.Router) {
	api.Get("/whitelist/enabled/:subdomain", func(c *fiber.Ctx) error {
		subdomain := c.Params("subdomain")
		enabled, err := redis.IsEnabledWhitelistForSubdomain(subdomain)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(fiber.Map{"enabled": enabled})
	})

	api.Put("/whitelist/enabled/:subdomain", func(c *fiber.Ctx) error {
		subdomain := c.Params("subdomain")
		var body struct {
			Enabled bool `json:"enabled"`
		}
		if err := c.BodyParser(&body); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid JSON"})
		}
		var err error
		if body.Enabled {
			err = redis.EnabledWhitelistForSubdomain(subdomain)
		} else {
			err = redis.DisabledWhitelistForSubdomain(subdomain)
		}
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(fiber.Map{"status": "ok"})
	})

	api.Put("/whitelist/enabled/", func(c *fiber.Ctx) error {
		var body struct {
			Enabled bool `json:"enabled"`
		}
		if err := c.BodyParser(&body); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid JSON"})
		}
		var err error
		if body.Enabled {
			err = redis.EnabledWhitelistForSubdomain("")
		} else {
			err = redis.DisabledWhitelistForSubdomain("")
		}
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(fiber.Map{"status": "ok"})
	})

	api.Get("/whitelist/enabled", func(c *fiber.Ctx) error {
		subdomains, err := redis.GetAllEnabledWhitelistSubdomains()
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(subdomains)
	})

	api.Get("/whitelist/ips/:subdomain", func(c *fiber.Ctx) error {
		subdomain := c.Params("subdomain")
		ips, err := redis.GetAllIPsForWhitelist(subdomain)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(ips)
	})

	api.Get("/whitelist/ips/", func(c *fiber.Ctx) error {
		ips, err := redis.GetAllIPsForWhitelist("")
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(ips)
	})

	api.Post("/whitelist/ip", func(c *fiber.Ctx) error {
		var body struct {
			Subdomain string       `json:"subdomain"`
			IP        string       `json:"ip"`
			Reason    redis.Reason `json:"reason"`
			Duration  string       `json:"duration"`
		}
		if err := c.BodyParser(&body); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid JSON"})
		}
		dur, err := time.ParseDuration(body.Duration)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid duration"})
		}
		err = redis.SetIPForWhitelist(body.Subdomain, body.IP, body.Reason, dur)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(fiber.Map{"status": "ok"})
	})

	api.Delete("/whitelist/ip/:subdomain/:ip", func(c *fiber.Ctx) error {
		subdomain := c.Params("subdomain")
		ip := c.Params("ip")
		err := redis.RemoveIPFromWhitelist(subdomain, ip)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(fiber.Map{"status": "ok"})
	})

	api.Delete("/whitelist/root/ip/remove/:ip", func(c *fiber.Ctx) error {
		ip := c.Params("ip")
		err := redis.RemoveIPFromWhitelist("", ip)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(fiber.Map{"status": "ok"})
	})
}
