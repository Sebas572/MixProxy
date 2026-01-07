package api

import (
	"github.com/gofiber/fiber/v2"
)

func SetupStatsRoutes(api fiber.Router) {
	api.Get("/stats", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"totalRequests":     0,
			"activeConnections": 0,
			"uniqueIPs":         0,
		})
	})

	api.Get("/requests", func(c *fiber.Ctx) error {
		return c.JSON([]fiber.Map{})
	})

	api.Get("/ips", func(c *fiber.Ctx) error {
		return c.JSON([]fiber.Map{})
	})
}
