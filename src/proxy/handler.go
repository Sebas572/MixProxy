package proxy

import (
	api "mixproxy/src/api/admin"
	"mixproxy/src/proxy/config"
	"mixproxy/src/proxy/tools"
	"net/http"
	"net/http/httputil"
	"strings"
)

func getHandleFunc(cfg *config.Config) http.HandlerFunc {
	// Manejador HTTP principal
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extraer dominio sin puerto
		host := strings.Split(r.Host, ":")[0]

		ip := r.RemoteAddr
		if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
			ip = forwarded
		}
		subdomain := ""
		if host != cfg.Hostname {
			subdomain = strings.Split(host, ".")[0]
		}

		tools.PrintLog(r.Method, r.RequestURI, ip, host)

		if subdomain == cfg.SubdomainAdminPanel {
			user, pass, ok := r.BasicAuth()
			if !ok || user != "admin" || pass != "password" {
				w.Header().Set("WWW-Authenticate", `Basic realm="Admin"`)
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			// Proxy to localhost:5173
			proxy := httputil.NewSingleHostReverseProxy(config.URL_ADMIN_PANEL)
			proxy.ServeHTTP(w, r)
			return
		}

		if subdomain == "admin-api" {
			api.HandleAdminAPI(w, r)
			return
		}

		if _, ok := config.Proxies[subdomain]; !ok {
			http.Error(w, "Dominio no configurado: "+host, http.StatusNotFound)
			return
		}

		target, err := tools.GetTargetIPForSubdomain(subdomain)
		if err != nil {
			http.Error(w, "Error al obtener target: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Log the request
		if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
			ip = forwarded
		}
		config.AddRequestLog(r.Method, r.URL.String(), ip, subdomain, 200) // Assuming success for now

		target.ServeHTTP(w, r)
	})

	return handler
}
