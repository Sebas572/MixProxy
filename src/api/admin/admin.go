package api

import (
	"encoding/json"
	"mixproxy/src/proxy/config"
	"net/http"
	"os"
	"time"
)

var controlFunc func(string)

func SetControlFunc(f func(string)) {
	controlFunc = f
}

func HandleAdminAPI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		return
	}

	path := r.URL.Path

	switch path {
	case "/api/start":
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "Processing"})

		go func() {
			time.Sleep(5 * time.Second)
			controlFunc("start")
		}()
	case "/api/stop":
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "Processing"})

		go func() {
			time.Sleep(5 * time.Second)
			controlFunc("stop")
		}()
	case "/api/reload":
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "Processing"})

		go func() {
			time.Sleep(5 * time.Second)
			controlFunc("reload")
		}()
	case "/api/stats":
		stats := config.GetStats()
		json.NewEncoder(w).Encode(stats)
	case "/api/requests":
		requests := config.GetRequestLogs()
		json.NewEncoder(w).Encode(requests)
	case "/api/ips":
		ips := config.GetIPStats()
		json.NewEncoder(w).Encode(ips)
	case "/api/config":
		if r.Method == "GET" {
			cfg, _ := config.ReadConfig("proxy.config.json")
			json.NewEncoder(w).Encode(cfg)
		} else if r.Method == "PUT" {
			var newCfg config.Config
			if err := json.NewDecoder(r.Body).Decode(&newCfg); err != nil {
				http.Error(w, "Invalid JSON", http.StatusBadRequest)
				return
			}
			// Validate the config
			if err := config.ValidateConfig(&newCfg); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			// Write to file
			data, _ := json.MarshalIndent(newCfg, "", "  ")
			os.WriteFile("proxy.config.json", data, 0644)
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]string{"status": "updated"})
		}
	default:
		http.NotFound(w, r)
	}
}
