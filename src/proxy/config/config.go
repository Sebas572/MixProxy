package config

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

type Config struct {
	Hostname            string              `json:"hostname"`
	SubdomainAdminPanel string              `json:"subdomain_admin_panel"`
	OnHTTPS             bool                `json:"on_https"`
	ModeDeveloper       bool                `json:"mode_developer"`
	LoadBalancer        []LoadBalancerEntry `json:"load_balancer"`
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

var SERVERS map[string]*http.Server = map[string]*http.Server{
	"HTTP":  &http.Server{Addr: ":80"},
	"HTTPS": &http.Server{Addr: ":443"},
}
var Proxies map[string][]*httputil.ReverseProxy = make(map[string][]*httputil.ReverseProxy)
var URL_ADMIN_PANEL *url.URL = mustParseURL("http://localhost:5173")

// Monitoring data structures
type RequestLog struct {
	ID        string    `json:"id"`
	Method    string    `json:"method"`
	URL       string    `json:"url"`
	IP        string    `json:"ip"`
	Subdomain string    `json:"subdomain"`
	Timestamp time.Time `json:"timestamp"`
	Status    int       `json:"status"`
}

type IPStat struct {
	IP       string    `json:"ip"`
	Count    int       `json:"count"`
	LastSeen time.Time `json:"lastSeen"`
}

type Stats struct {
	TotalRequests     int `json:"totalRequests"`
	ActiveConnections int `json:"activeConnections"`
	UniqueIPs         int `json:"uniqueIPs"`
}

var (
	requestLogs []RequestLog
	ipStats     map[string]*IPStat
	stats       Stats
	mu          sync.RWMutex
)

func init() {
	ipStats = make(map[string]*IPStat)
}

func AddRequestLog(method, url, ip, subdomain string, status int) {
	mu.Lock()
	defer mu.Unlock()

	log := RequestLog{
		ID:        fmt.Sprintf("%d", time.Now().UnixNano()),
		Method:    method,
		URL:       url,
		IP:        ip,
		Subdomain: subdomain,
		Timestamp: time.Now(),
		Status:    status,
	}

	requestLogs = append(requestLogs, log)
	if len(requestLogs) > 100 {
		requestLogs = requestLogs[1:]
	}

	stats.TotalRequests++

	if stat, exists := ipStats[ip]; exists {
		stat.Count++
		stat.LastSeen = time.Now()
	} else {
		ipStats[ip] = &IPStat{
			IP:       strings.Split(ip, ":")[0],
			Count:    1,
			LastSeen: time.Now(),
		}
		stats.UniqueIPs = len(ipStats)
	}
}

func GetRequestLogs() []RequestLog {
	mu.RLock()
	defer mu.RUnlock()
	return append([]RequestLog{}, requestLogs...)
}

func GetIPStats() []IPStat {
	mu.RLock()
	defer mu.RUnlock()
	var stats []IPStat = []IPStat{}
	for _, stat := range ipStats {
		stats = append(stats, *stat)
	}
	return stats
}

func GetStats() Stats {
	mu.RLock()
	defer mu.RUnlock()
	return stats
}

func ValidateConfig(cfg *Config) error {
	for _, e := range cfg.LoadBalancer {
		sum := 0.0
		for _, v := range e.VPS {
			if !v.Active {
				continue
			}

			if v.Capacity > 1 || v.Capacity < 0 {
				return fmt.Errorf("capacity must be between 0 and 1")
			}
			sum += v.Capacity
		}
		if sum != 1 {
			return fmt.Errorf("sum of capacities must be 1")
		}
	}
	return nil
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

	return &cfg, nil
}

func mustParseURL(raw string) *url.URL {
	u, err := url.Parse(raw)
	if err != nil {
		panic(err)
	}
	return u
}
