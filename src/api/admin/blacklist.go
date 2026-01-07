package api

import (
	"time"

	"mixproxy/src/redis"

	"github.com/gofiber/fiber/v2"
)

func SetupBlacklistRoutes(api fiber.Router) {
	api.Get("/blacklist/enabled/:subdomain", func(c *fiber.Ctx) error {
		subdomain := c.Params("subdomain")
		enabled, err := redis.IsEnabledBlacklistForSubdomain(subdomain)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(fiber.Map{"enabled": enabled})
	})

	api.Put("/blacklist/enabled/:subdomain", func(c *fiber.Ctx) error {
		subdomain := c.Params("subdomain")
		var body struct {
			Enabled bool `json:"enabled"`
		}
		if err := c.BodyParser(&body); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid JSON"})
		}
		var err error
		if body.Enabled {
			err = redis.EnabledBlacklistForSubdomain(subdomain)
		} else {
			err = redis.DisabledBlacklistForSubdomain(subdomain)
		}
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(fiber.Map{"status": "ok"})
	})

	api.Put("/blacklist/enabled/", func(c *fiber.Ctx) error {
		var body struct {
			Enabled bool `json:"enabled"`
		}
		if err := c.BodyParser(&body); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid JSON"})
		}
		var err error
		if body.Enabled {
			err = redis.EnabledBlacklistForSubdomain("")
		} else {
			err = redis.DisabledBlacklistForSubdomain("")
		}
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(fiber.Map{"status": "ok"})
	})

	api.Get("/blacklist/enabled", func(c *fiber.Ctx) error {
		subdomains, err := redis.GetAllEnabledBlacklistSubdomains()
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(subdomains)
	})

	api.Get("/blacklist/ips/:subdomain", func(c *fiber.Ctx) error {
		subdomain := c.Params("subdomain")
		ips, err := redis.GetAllIPsForBlacklist(subdomain)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}

		return c.JSON(ips)
	})

	api.Get("/blacklist/global/ips", func(c *fiber.Ctx) error {
		ips, err := redis.GetAllIPsForGlobalBlacklist()
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(ips)
	})

	api.Get("/blacklist/ips/", func(c *fiber.Ctx) error {
		ips, err := redis.GetAllIPsForBlacklist("")
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(ips)
	})

	api.Post("/blacklist/ip", func(c *fiber.Ctx) error {
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
		err = redis.SetIPForBlacklist(body.Subdomain, body.IP, body.Reason, dur)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(fiber.Map{"status": "ok"})
	})

	api.Post("/blacklist/global/ip", func(c *fiber.Ctx) error {
		var body struct {
			IP       string       `json:"ip"`
			Reason   redis.Reason `json:"reason"`
			Duration string       `json:"duration"`
		}
		if err := c.BodyParser(&body); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid JSON"})
		}
		dur, err := time.ParseDuration(body.Duration)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid duration"})
		}
		err = redis.SetIPForGlobalBlacklist(body.IP, body.Reason, dur)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(fiber.Map{"status": "ok"})
	})

	api.Delete("/blacklist/ip/:subdomain/:ip", func(c *fiber.Ctx) error {
		subdomain := c.Params("subdomain")
		ip := c.Params("ip")
		err := redis.RemoveIPFromBlacklist(subdomain, ip)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(fiber.Map{"status": "ok"})
	})

	api.Delete("/blacklist/global/ip/:ip", func(c *fiber.Ctx) error {
		ip := c.Params("ip")
		err := redis.RemoveIPFromGlobalBlacklist(ip)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(fiber.Map{"status": "ok"})
	})

	api.Delete("/blacklist/root/ip/remove/:ip", func(c *fiber.Ctx) error {
		ip := c.Params("ip")
		err := redis.RemoveIPFromBlacklist("", ip)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(fiber.Map{"status": "ok"})
	})
}
