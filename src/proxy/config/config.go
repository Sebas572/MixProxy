package config

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/gofiber/fiber/v2"
)

type Config struct {
	Hostname            string              `json:"hostname"`
	SubdomainAdminPanel string              `json:"subdomain_admin_panel"`
	OnHTTPS             bool                `json:"on_https"`
	ModeDeveloper       bool                `json:"mode_developer"`
	LoadBalancer        []LoadBalancerEntry `json:"load_balancer"`
	RootLoadBalancer    *LoadBalancerEntry  `json:"root_load_balancer,omitempty"`
}

type LoadBalancerEntry struct {
	VPS          []VPSEntry `json:"vps"`
	Type         string     `json:"type"`
	Subdomain    string     `json:"subdomain"`
	Active       bool       `json:"active"`
	CacheEnabled bool       `json:"cache_enabled"`
	CachePaths   []string   `json:"cache_paths"`
}

type VPSEntry struct {
	IP string `json:"ip"`
	// Capacity is a decimal value between 0.0 and 1.0 representing the proportion of requests to route to this backend.
	// All capacities in a load balancer entry must sum to exactly 1.0 for proper load balancing.
	Capacity float64 `json:"capacity"`
	Active   bool    `json:"active"`
}

var SERVERS map[string]*fiber.App = map[string]*fiber.App{
	"HTTP":  fiber.New(fiber.Config{DisableStartupMessage: true}),
	"HTTPS": fiber.New(fiber.Config{DisableStartupMessage: true}),
}
var Proxies map[string][]string = make(map[string][]string)
var URL_ADMIN_PANEL string = "http://admin:4173"
var CONFIG_PATH string = filepath.Join(".", ".config", "proxy.config.json")

func AllValuesNonEmpty(entry *LoadBalancerEntry) bool {
	return entry.Type != "" && len(entry.VPS) != 0
}

func ValidateConfig(cfg *Config) error {
	if cfg.LoadBalancer != nil && len(cfg.LoadBalancer) != 0 {
		for _, e := range cfg.LoadBalancer {
			sum := 0.0
			for _, v := range e.VPS {
				if !v.Active {
					fmt.Println("➖ Skipping inactive VPS:", v.IP)
					continue
				}

				if v.Capacity > 1 || v.Capacity < 0 {
					return fmt.Errorf("capacity for backend %s must be between 0.0 and 1.0 (inclusive), representing the proportion of requests to route to this backend", v.IP)
				}
				sum += v.Capacity
			}
			if sum != 1 {
				fmt.Printf("❌ Sum of capacities for subdomain '%s' is %.2f, but must be 1.0\n", e.Subdomain, sum)
				return fmt.Errorf("invalid load balancer configuration for subdomain '%s': sum of capacities must be 1.0", e.Subdomain)
			} else {
				fmt.Printf("✅ Load balancer for subdomain '%s' is correctly configured (sum = 1.0)\n", e.Subdomain)
			}

			// Validate cache paths
			if e.CacheEnabled {
				if len(e.CachePaths) == 0 {
					return fmt.Errorf("cache enabled for subdomain '%s' but no cache paths specified", e.Subdomain)
				}
				for _, path := range e.CachePaths {
					if !strings.HasPrefix(path, "/") {
						return fmt.Errorf("cache path '%s' for subdomain '%s' must start with '/'", path, e.Subdomain)
					}
				}
			}
		}
	} else {
		fmt.Println("The configuration file is empty")
	}

	if cfg.RootLoadBalancer != nil && AllValuesNonEmpty(cfg.RootLoadBalancer) {
		sum := 0.0
		for _, v := range cfg.RootLoadBalancer.VPS {
			if !v.Active {
				continue
			}

			if v.Capacity > 1 || v.Capacity < 0 {
				return fmt.Errorf("capacity for backend %s must be between 0.0 and 1.0 (inclusive), representing the proportion of requests to route to this backend", v.IP)
			}
			sum += v.Capacity
		}
		if sum != 1 {
			return fmt.Errorf("invalid root load balancer configuration: sum of capacities must be 1.0")
		}

		// Validate cache paths for root
		if cfg.RootLoadBalancer.CacheEnabled {
			if len(cfg.RootLoadBalancer.CachePaths) == 0 {
				return fmt.Errorf("cache enabled for root load balancer but no cache paths specified")
			}
			for _, path := range cfg.RootLoadBalancer.CachePaths {
				if !strings.HasPrefix(path, "/") {
					return fmt.Errorf("cache path '%s' for root load balancer must start with '/'", path)
				}
			}
		}
	}

	return nil
}

func ReadConfig() (*Config, error) {
	data, err := os.ReadFile(CONFIG_PATH)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func mustParseURL(raw string) *url.URL {
	u, err := url.Parse(raw)
	if err != nil {
		panic(err)
	}
	return u
}
