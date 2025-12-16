package proxy

import (
	"crypto/tls"
	"fmt"
	"log"
	"mixploy/src/proxy/config"
	"mixploy/src/proxy/tools"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

type LoadBalancer struct {
	URL      url.URL
	Capacity float64
}

func Start() {
	// Configurar backends

	cfg, _ := config.ReadConfig("proxy.config.json")

	loadBalancer := map[string]*[]LoadBalancer{}

	for _, e := range cfg.LoadBalancer {
		vps := []LoadBalancer{}
		subdomain := e.Subdomain

		for _, v := range e.VPS {
			vps = append(vps, LoadBalancer{
				URL:      *mustParseURL(v.IP),
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

	// Cargar certificado wildcard (para *.developer.space y developer.space)
	cert, err := tls.LoadX509KeyPair("certs/wildcard.developer.space.crt", "certs/wildcard.developer.space.key")
	if err != nil {
		log.Fatal("Error cargando certificado: ", err)
	}

	// Configuración TLS simple - usamos el mismo certificado para todo
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
	}

	// Manejador HTTP principal
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extraer dominio sin puerto
		host := strings.Split(r.Host, ":")[0]

		log.Printf("📨 Solicitud para: %s", host)

		if host == "developer.space" {
			w.Header().Set("Content-Type", "text/html")
			fmt.Fprintf(w, `<h1>Balanceador Go Funcionando!</h1>
				<p>Dominio actual: %s</p>
				<ul>
					<li><a href="https://api.developer.space">API</a></li>
					<li><a href="https://app.developer.space">App</a></li>
				</ul>`, host)
		}

		subdomain := strings.Split(host, ".")[0]

		if _, ok := config.Proxies[subdomain]; !ok {
			http.Error(w, "Dominio no configurado: "+host, http.StatusNotFound)
			return
		}

		target, err := tools.GetTargetIPForSubdomain(subdomain)
		if err != nil {
			http.Error(w, "Error al obtener target: "+err.Error(), http.StatusInternalServerError)
			return
		}
		target.ServeHTTP(w, r)
	})

	// Redirigir HTTP a HTTPS
	go func() {
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			target := "https://" + r.Host + r.URL.Path
			http.Redirect(w, r, target, http.StatusMovedPermanently)
		})
		log.Println("🌐 HTTP redirigiendo en :80")
		log.Fatal(http.ListenAndServe(":80", nil))
	}()

	// Servidor HTTPS
	log.Println("🚀 Balanceador iniciando en HTTPS :443")
	server := &http.Server{
		Addr:      ":443",
		Handler:   handler,
		TLSConfig: tlsConfig,
	}

	log.Fatal(server.ListenAndServeTLS("", ""))
}

func mustParseURL(raw string) *url.URL {
	u, err := url.Parse(raw)
	if err != nil {
		panic(err)
	}
	return u
}
