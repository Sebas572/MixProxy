package proxy

import (
	"context"
	"fmt"
	"log"
	api "mixproxy/src/api/admin"
	certificate "mixproxy/src/certs"
	"mixproxy/src/proxy/config"
	"mixproxy/src/proxy/tools"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"sync"
	"time"
)

type LoadBalancer struct {
	URL      url.URL
	Capacity float64
}

func Stop() {
	if err := config.SERVERS["HTTP"].Shutdown(context.TODO()); err != nil {
		log.Println("Error stopping HTTP:", err)
	}
	if err := config.SERVERS["HTTPS"].Shutdown(context.TODO()); err != nil {
		log.Println("Error stopping HTTPS:", err)
	}
}

func Control(action string) {
	switch action {
	case "start":
		Start()
	case "stop":
		Stop()
	case "reload":
		Stop()
		time.Sleep(10 * time.Second)
		Start()
	case "createCertificates":
		os.RemoveAll("./certs")
		createCertificates()
	default:
		fmt.Println("Invalid action")
	}
}

func createCertificates() {
	cfg, _ := config.ReadConfig()

	if _, err := os.Stat("./certs/wildcard.developer.space.crt"); err == nil {
		return
	}

	host := cfg.Hostname

	DNSnames := []string{host, cfg.SubdomainAdminPanel + "." + host, "admin-api." + host, "localhost", "127.0.0.1", "*.localhost"}

	for _, server := range cfg.LoadBalancer {
		if server.Subdomain != "" {
			DNSnames = append(DNSnames, server.Subdomain+"."+host)
		}
	}

	certificate.Create(DNSnames)
}

func startHttpAndHttpsServer(wg *sync.WaitGroup, cfg *config.Config) {
	handler := getHandleFunc(cfg)
	tlsConfig := getTlsConfig()

	config.SERVERS["HTTPS"].Handler = handler
	config.SERVERS["HTTPS"].TLSConfig = tlsConfig

	// Create mux for HTTP redirection
	httpMux := http.NewServeMux()
	httpMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		target := "https://" + r.Host + r.URL.Path
		http.Redirect(w, r, target, http.StatusMovedPermanently)
	})
	config.SERVERS["HTTP"].Handler = httpMux

	go func() {
		defer wg.Done()

		log.Println("🌐 HTTP redirigiendo en :80")

		// always returns error. ErrServerClosed on graceful close
		if err := config.SERVERS["HTTP"].ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("ListenAndServe(): %v", err)
			os.Exit(1)
		}

	}()

	go func() {
		defer wg.Done()

		log.Println("🚀 Balanceador iniciando en HTTPS :443")
		if err := config.SERVERS["HTTPS"].ListenAndServeTLS("", ""); err != http.ErrServerClosed {
			log.Fatalf("ListenAndServeTLS(): %v", err)
		}
	}()
}

func Start() {
	api.SetControlFunc(Control)

	// Reset global state for reload
	config.SERVERS = map[string]*http.Server{
		"HTTP":  &http.Server{Addr: ":80"},
		"HTTPS": &http.Server{Addr: ":443"},
	}
	config.Proxies = make(map[string][]*httputil.ReverseProxy)
	tools.ServerSelected = make(map[string]*tools.ServerEntry)

	// Configurar backends
	cfg, _ := config.ReadConfig()

	if err := config.ValidateConfig(cfg); err != nil {
		fmt.Println("❌ Error de validación:", err)
		os.Exit(0)
	}

	loadBalancer := map[string]*[]LoadBalancer{}

	if cfg.ModeDeveloper {
		fmt.Println("Configuring certificates in development mode")
		createCertificates()
	}

	for _, e := range cfg.LoadBalancer {
		vps := []LoadBalancer{}
		subdomain := e.Subdomain

		for _, v := range e.VPS {
			vps = append(vps, LoadBalancer{
				URL:      *tools.MustParseURL(v.IP),
				Capacity: v.Capacity,
			})
		}

		loadBalancer[subdomain] = &vps
		probability := []tools.VpsProbability{}
		for _, v := range e.VPS {
			probability = append(probability, tools.VpsProbability{
				Probability: v.Capacity,
				IP:          v.IP,
			})
		}
		tools.SetupServerSelected(subdomain, probability)
	}

	if cfg.RootLoadBalancer != nil && config.AllValuesNonEmpty(cfg.RootLoadBalancer) {
		vps := []LoadBalancer{}
		subdomain := ""

		for _, v := range cfg.RootLoadBalancer.VPS {
			vps = append(vps, LoadBalancer{
				URL:      *tools.MustParseURL(v.IP),
				Capacity: v.Capacity,
			})
		}

		loadBalancer[subdomain] = &vps
		probability := []tools.VpsProbability{}
		for _, v := range cfg.RootLoadBalancer.VPS {
			probability = append(probability, tools.VpsProbability{
				Probability: v.Capacity,
				IP:          v.IP,
			})
		}
		tools.SetupServerSelected(subdomain, probability)
	}

	if len(loadBalancer) != 0 {
		for subdomain, targets := range loadBalancer {
			for _, target := range *targets {
				proxy := httputil.NewSingleHostReverseProxy(&target.URL)

				// Guardar el director original
				originalDirector := proxy.Director

				// Sobrescribir el Director para manejar el encabezado Host
				proxy.Director = func(req *http.Request) {
					// 1. Primero, llamar al director original
					originalDirector(req)

					req.Host = target.URL.Host // target.Host -> ejemplo ("localhost:3001")

					if req.Header.Get("Upgrade") == "websocket" {
						log.Printf("🔄 Proxy: Iniciando conexión WebSocket a %s", target.URL.Host)
					}
				}

				config.Proxies[subdomain] = append(config.Proxies[subdomain], proxy)
			}
		}
	}
	// Redirigir HTTP a HTTPS
	httpServerExitDone := &sync.WaitGroup{}
	httpServerExitDone.Add(2)

	startHttpAndHttpsServer(httpServerExitDone, cfg)

	httpServerExitDone.Wait()
}
