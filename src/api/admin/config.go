package api

import (
	"bytes"
	"encoding/json"
	"os"

	"mixproxy/src/proxy/config"
	"mixproxy/src/redis"

	"github.com/gofiber/fiber/v2"
)

func SetupConfigRoutes(api fiber.Router) {
	api.Get("/config", func(c *fiber.Ctx) error {
		cfg, _ = config.ReadConfig()
		response := ConfigResponse{
			Hostname:            cfg.Hostname,
			SubdomainAdminPanel: cfg.SubdomainAdminPanel,
			OnHTTPS:             cfg.OnHTTPS,
			ModeDeveloper:       cfg.ModeDeveloper,
			LoadBalancer:        cfg.LoadBalancer,
			RootLoadBalancer:    cfg.RootLoadBalancer,
		}
		return c.JSON(response)
	})

	api.Put("/config", func(c *fiber.Ctx) error {
		var newCfg config.Config

		if err := json.NewDecoder(bytes.NewReader(c.Body())).Decode(&newCfg); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "JSON mal formado",
				"details": err.Error(),
			})
		}
		// Validate the config
		if err := config.ValidateConfig(&newCfg); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "Configuración inválida",
				"details": err.Error(),
			})
		}

		// Read current config for comparison
		oldCfg, err := config.ReadConfig()
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Error al leer configuración actual",
			})
		}

		// Find removed subdomains
		removedSubdomains := []string{}
		for _, oldLb := range oldCfg.LoadBalancer {
			found := false
			for _, newLb := range newCfg.LoadBalancer {
				if oldLb.Subdomain == newLb.Subdomain {
					found = true
					break
				}
			}
			if !found {
				removedSubdomains = append(removedSubdomains, oldLb.Subdomain)
			}
		}

		// Check if root load balancer was removed
		if oldCfg.RootLoadBalancer != nil && newCfg.RootLoadBalancer == nil {
			removedSubdomains = append(removedSubdomains, "")
		}

		// Clean up Redis for removed subdomains
		for _, sub := range removedSubdomains {
			redis.RemoveAllForWhitelistSubdomain(sub)
			redis.RemoveAllForBlacklistSubdomain(sub)
		}

		newCfg.AdminUsername = cfg.AdminUsername
		newCfg.AdminPassword = cfg.AdminPassword

		// Write to file
		data, err := json.MarshalIndent(newCfg, "", "  ")
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Error al serializar configuración",
			})
		}
		if err := os.WriteFile(config.CONFIG_PATH, data, 0644); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Error al escribir configuración",
			})
		}

		return c.JSON(fiber.Map{"status": "updated"})
	})

	api.Post("/config/change/subdominio", func(c *fiber.Ctx) error {
		var body struct {
			OldSubdomain string `json:"OldSubdomain"`
			NewSubdomain string `json:"NewSubdomain"`
		}

		if err := c.BodyParser(&body); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid JSON",
			})
		}

		cfg, err := config.ReadConfig()
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to read config",
			})
		}

		// Verify that the new domain is not being used by another server
		for _, lb := range cfg.LoadBalancer {
			if lb.Subdomain == body.NewSubdomain {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error": "The new subdomain is already in use",
				})
			}
		}

		// Find and update the load balancer entry
		found := false
		for i, lb := range cfg.LoadBalancer {
			if lb.Subdomain == body.OldSubdomain {
				cfg.LoadBalancer[i].Subdomain = body.NewSubdomain
				found = true
				break
			}
		}

		if !found {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Load balancer with old subdomain not found",
			})
		}

		// Validate the config
		if err := config.ValidateConfig(cfg); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "Configuración inválida",
				"details": err.Error(),
			})
		}

		// Write to file
		data, err := json.MarshalIndent(cfg, "", "  ")
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Error al serializar configuración",
			})
		}
		if err := os.WriteFile(config.CONFIG_PATH, data, 0644); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Error al escribir configuración",
			})
		}

		redis.ChangeSubdomainWhitelist(body.OldSubdomain, body.NewSubdomain)
		redis.ChangeSubdomainBlacklist(body.OldSubdomain, body.NewSubdomain)

		return c.JSON(fiber.Map{"status": "subdomain changed"})
	})
}
