package api

import (
	"bytes"
	"encoding/json"
	"mixproxy/src/logger"
	"mixproxy/src/proxy/config"
	"mixproxy/src/redis"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

type ConfigResponse struct {
	Hostname            string                     `json:"hostname"`
	SubdomainAdminPanel string                     `json:"subdomain_admin_panel"`
	OnHTTPS             bool                       `json:"on_https"`
	ModeDeveloper       bool                       `json:"mode_developer"`
	LoadBalancer        []config.LoadBalancerEntry `json:"load_balancer"`
	RootLoadBalancer    *config.LoadBalancerEntry  `json:"root_load_balancer,omitempty"`
}

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
		AllowMethods: "GET,POST,PUT,OPTIONS,DELETE",
		AllowHeaders: "Content-Type",
	}))

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

	api.Use(adminApiMiddleware)

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

	// Whitelist endpoints
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

	// Blacklist endpoints
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
