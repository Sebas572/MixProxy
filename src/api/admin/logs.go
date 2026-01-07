package api

import (
	"os"
	"strings"

	"mixproxy/src/logger"

	"github.com/gofiber/fiber/v2"
)

func SetupLogsPublic(api fiber.Router) {
	api.Get("/logs", func(c *fiber.Ctx) error {
		date := c.Query("date")
		var logFile string
		if date == "" {
			logFile = logger.GetCurrentLogFile()
			if logFile == "" {
				return c.Status(404).SendString("No current log file")
			}
		} else {
			logFile = "./logs/log-" + date
		}

		// Check if file exists
		if _, err := os.Stat(logFile); os.IsNotExist(err) {
			return c.Status(404).SendString("Log file not found")
		}

		return c.SendFile(logFile)
	})
}

func SetupLogsProtected(api fiber.Router) {
	api.Get("/logs/list", func(c *fiber.Ctx) error {
		entries, err := os.ReadDir("./logs")
		if err != nil {
			return c.Status(500).SendString("Error reading logs directory")
		}

		var logFiles []string
		for _, entry := range entries {
			if !entry.IsDir() && strings.HasPrefix(entry.Name(), "log-") {
				logFiles = append(logFiles, entry.Name())
			}
		}

		return c.JSON(logFiles)
	})
}
