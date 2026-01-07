package api

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"mixproxy/src/proxy/config"
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