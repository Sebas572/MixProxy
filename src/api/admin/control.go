package api

import (
	"time"

	"github.com/gofiber/fiber/v2"
)

func SetupControlRoutes(api fiber.Router) {
	api.Get("/start", func(c *fiber.Ctx) error {
		go func() {
			time.Sleep(5 * time.Second)
			controlFunc("start")
		}()

		return c.JSON(fiber.Map{"status": "Processing"})
	})

	api.Get("/stop", func(c *fiber.Ctx) error {
		go func() {
			time.Sleep(5 * time.Second)
			controlFunc("stop")
		}()

		return c.JSON(fiber.Map{"status": "Processing"})
	})

	api.Post("/reload", func(c *fiber.Ctx) error {
		go func() {
			time.Sleep(5 * time.Second)
			controlFunc("reload")
		}()

		return c.JSON(fiber.Map{"status": "Processing"})
	})
}
