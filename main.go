package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

func main() {
	// Configurar backends
	backends := map[string]*url.URL{
		"api": mustParseURL("http://localhost:3001"),
		"app": mustParseURL("http://localhost:3002"),
	}

	// Crear proxies
	proxies := make(map[string]*httputil.ReverseProxy)
	for name, target := range backends {
		proxies[name] = httputil.NewSingleHostReverseProxy(target)
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

		// Enrutamiento simple
		if strings.HasPrefix(host, "api.") {
			log.Printf("-> Enrutando a API backend")
			r.Header.Set("X-Forwarded-Host", r.Host)
			proxies["api"].ServeHTTP(w, r)
		} else if strings.HasPrefix(host, "app.") {
			log.Printf("-> Enrutando a App backend")
			r.Header.Set("X-Forwarded-Host", r.Host)
			proxies["app"].ServeHTTP(w, r)
		} else if host == "developer.space" {
			log.Printf("-> Página principal")
			w.Header().Set("Content-Type", "text/html")
			fmt.Fprintf(w, `<h1>Balanceador Go Funcionando!</h1>
				<p>Dominio actual: %s</p>
				<ul>
					<li><a href="https://api.developer.space">API</a></li>
					<li><a href="https://app.developer.space">App</a></li>
				</ul>`, host)
		} else {
			http.Error(w, "Dominio no configurado: "+host, http.StatusNotFound)
		}
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

// package main

// import "traefik-cli/src/cmd/cli"

// func main() {
// 	cli.Execute()
// }
