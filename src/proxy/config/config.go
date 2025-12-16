package config

import (
	"encoding/json"
	"net/http/httputil"
	"net/url"
	"os"
)

type Config struct {
	Hostname      string              `json:"hostname"`
	OnHTTPS       bool                `json:"on_https"`
	ModeDeveloper bool                `json:"mode_developer"`
	LoadBalancer  []LoadBalancerEntry `json:"load_balancer"`
}

type LoadBalancerEntry struct {
	VPS       []VPSEntry `json:"vps"`
	Type      string     `json:"type"`
	Subdomain string     `json:"subdomain"`
	Active    bool       `json:"active"`
}

type VPSEntry struct {
	IP       string  `json:"ip"`
	Capacity float64 `json:"capacity"`
	Active   bool    `json:"active"`
}

var Proxies map[string][]*httputil.ReverseProxy = make(map[string][]*httputil.ReverseProxy)
var URL_ADMIN_PANEL *url.URL = mustParseURL("http://localhost:5173")

func validate_percentage(cfg *Config) {
	sum := 0.00

	for _, e := range cfg.LoadBalancer {
		for _, v := range e.VPS {
			if !v.Active {
				continue
			}

			if v.Capacity > 1 || v.Capacity < 0 {
				panic("capacity must be between 0 and 1")
			}
			sum += v.Capacity
		}
		if sum != 1 {
			panic("sum of capacities must be 1")
		}
	}
}

func ReadConfig(filePath string) (*Config, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	validate_percentage(&cfg)

	return &cfg, nil
}

func mustParseURL(raw string) *url.URL {
	u, err := url.Parse(raw)
	if err != nil {
		panic(err)
	}
	return u
}
