package api

import (
	"mixproxy/src/docker"

	"github.com/gofiber/fiber/v2"
)

func SetupDockerRoutes(api fiber.Router) {
	api.Get("/docker/containers", func(c *fiber.Ctx) error {
		all := c.Query("all") == "true"
		containers := docker.GetContainers(all)

		return c.JSON(fiber.Map{"containers": containers})
	})

	api.Post("/docker/containers/start/:id", func(c *fiber.Ctx) error {
		id := c.Params("id")
		docker.StartContainer(id)
		return c.JSON(fiber.Map{"status": "started"})
	})

	api.Post("/docker/containers/stop/:id", func(c *fiber.Ctx) error {
		id := c.Params("id")
		docker.StopContainer(id)
		return c.JSON(fiber.Map{"status": "stopped"})
	})

	api.Post("/docker/containers/create", func(c *fiber.Ctx) error {
		var body struct {
			Image string `json:"image"`
			Name  string `json:"name"`
		}
		if err := c.BodyParser(&body); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid JSON"})
		}
		docker.CreateContainer(body.Image, body.Name)
		return c.JSON(fiber.Map{"status": "created"})
	})
}
