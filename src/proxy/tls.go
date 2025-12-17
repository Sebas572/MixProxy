package proxy

import (
	"crypto/tls"
	"log"
)

func getTlsConfig() *tls.Config {
	// Cargar certificado wildcard (para *.developer.space y developer.space)
	cert, err := tls.LoadX509KeyPair("certs/wildcard.developer.space.crt", "certs/wildcard.developer.space.key")
	if err != nil {
		log.Fatal("Error cargando certificado: ", err)
	}

	// Configuración TLS simple - usamos el mismo certificado para todo
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
	}

	return tlsConfig
}
