package api

import (
	"encoding/json"
	"mixproxy/src/docker"
	"os"

	"github.com/gofiber/fiber/v2"
)

type Template struct {
	Image       string   `json:"image"`
	Ports       int      `json:"ports"`
	Environment []string `json:"environment"`
	Url         string   `json:"url"`
}

type Templates struct {
	Databases   []Template `json:"databases"`
	Repositorys []Template `json:"repositorys"`
}

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

	api.Get("/docker/templates", func(c *fiber.Ctx) error {
		file, err := os.Open("containers.template.json")
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to read templates"})
		}
		defer file.Close()
		var templates Templates
		if err := json.NewDecoder(file).Decode(&templates); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to parse templates"})
		}
		return c.JSON(templates)
	})

	api.Post("/docker/containers/create-from-template", func(c *fiber.Ctx) error {
		var body struct {
			Type         string            `json:"type"`
			Index        int               `json:"index"`
			Name         string            `json:"name"`
			Environments map[string]string `json:"environments"`
		}
		if err := c.BodyParser(&body); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid JSON"})
		}
		file, err := os.Open("containers.template.json")
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to read templates"})
		}
		defer file.Close()
		var templates Templates
		if err := json.NewDecoder(file).Decode(&templates); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to parse templates"})
		}
		var selectedTemplate Template
		if body.Type == "databases" {
			if body.Index < 0 || body.Index >= len(templates.Databases) {
				return c.Status(400).JSON(fiber.Map{"error": "Invalid index"})
			}
			selectedTemplate = templates.Databases[body.Index]
		} else if body.Type == "repositorys" {
			if body.Index < 0 || body.Index >= len(templates.Repositorys) {
				return c.Status(400).JSON(fiber.Map{"error": "Invalid index"})
			}
			selectedTemplate = templates.Repositorys[body.Index]
		} else {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid type"})
		}
		envs := []string{}
		for _, envKey := range selectedTemplate.Environment {
			if val, ok := body.Environments[envKey]; ok && val != "" {
				envs = append(envs, envKey+"="+val)
			}
		}
		docker.CreateContainerFromTemplate(selectedTemplate.Image, body.Name, selectedTemplate.Ports, envs)
		return c.JSON(fiber.Map{"status": "created"})
	})
}
