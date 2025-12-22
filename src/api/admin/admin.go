package api

import (
	"bytes"
	"encoding/json"
	"mixproxy/src/proxy/config"
	"os"
	"strings"
	"time"

	"mixproxy/src/redis"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

var controlFunc func(string)
var cfg *config.Config

func init() {
	cfg, _ = config.ReadConfig()
}

func SetControlFunc(f func(string)) {
	controlFunc = f
}

func adminApiMiddleware(c *fiber.Ctx) error {
	hostAndPort := string(c.BaseURL())
	host := strings.Split(hostAndPort, "//")[1]
	subdomain := ""

	if host != cfg.Hostname {
		subdomain = strings.Split(host, ".")[0]
	}

	if subdomain != "admin-api" {
		return c.Status(fiber.StatusNotFound).SendString("Not found")
	}

	// TODO:
	// token := c.Get("Authorization")

	// Verificar si el token es válido (aquí un ejemplo simple)
	// if token != "Bearer mi_token_secreto" {
	//     return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
	//         "error": "No autorizado",
	//     })
	// }

	return c.Next()
}

func HandleAdminAPI() {
	api := config.SERVERS["HTTPS"].Group("/api")
	api.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,PUT,OPTIONS",
		AllowHeaders: "Content-Type",
	}))
	api.Use(adminApiMiddleware)

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

	api.Get("/stats", func(c *fiber.Ctx) error {
		stats, err := redis.GetStats()
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}

		return c.JSON(stats)
	})

	api.Get("/requests", func(c *fiber.Ctx) error {
		requests, err := redis.GetRequestLogs()
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}

		return c.JSON(requests)
	})

	api.Get("/ips", func(c *fiber.Ctx) error {
		ips, err := redis.GetIPStats()
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}

		return c.JSON(ips)
	})

	api.Get("/config", func(c *fiber.Ctx) error {
		cfg, _ = config.ReadConfig()
		return c.JSON(cfg)
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
}
