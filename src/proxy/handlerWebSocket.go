package proxy

import (
	"crypto/tls"
	"log"
	"mixproxy/src/proxy/config"
	"mixproxy/src/proxy/tools"
	"strings"

	fws "github.com/fasthttp/websocket"
	"github.com/gofiber/websocket/v2"
)

func getSubdomainFromWebSocket(ctx *websocket.Conn) string {
	origin := ctx.Headers("Origin")
	subdomain := strings.Split(strings.Split(origin, "://")[1], ".")[0]

	return subdomain
}

func getHandleFuncFromWebSocket(ctx *websocket.Conn) (string, error) {
	subdomain := getSubdomainFromWebSocket(ctx)

	target, err := tools.GetTargetIPForSubdomain(subdomain)
	if err != nil {
		return config.URL_ADMIN_PANEL, err
	}

	target = strings.Replace(target, "http://", "ws://", 1)
	target = strings.Replace(target, "https://", "wss://", 1)

	return target, nil
}

// Nueva función para manejar WebSockets
func handleWebSocket(c *websocket.Conn) {
	defer c.Close()

	url, err := getHandleFuncFromWebSocket(c)
	if err != nil {
		log.Printf("Error obtaining URL for WebSocket: %v", err)
		return
	}

	// Verificar si la URL es wss:// o ws:// y configurar el Dialer
	dialer := fws.Dialer{}
	if strings.HasPrefix(url, "wss://") {
		dialer.TLSClientConfig = &tls.Config{
			InsecureSkipVerify: true, // Esto permite conexiones inseguras, pero no es recomendado para producción
		}
	}

	serverConn, _, err := dialer.Dial(url, nil)
	if err != nil {
		log.Printf("Error connecting to the WebSocket server: %v", err)
		return
	}
	defer serverConn.Close()

	// Túnel entre cliente y servidor
	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			messageType, message, err := c.ReadMessage()
			if err != nil {
				log.Printf("Error reading client message: %v", err)
				break
			}
			if err := serverConn.WriteMessage(messageType, message); err != nil {
				log.Printf("Error writing message to server: %v", err)
				break
			}
		}
	}()

	for {
		select {
		case <-done:
			return
		default:
			messageType, message, err := serverConn.ReadMessage()
			if err != nil {
				log.Printf("Error reading message from server: %v", err)
				return
			}
			if err := c.WriteMessage(messageType, message); err != nil {
				log.Printf("Error writing message to customer: %v", err)
				return
			}
		}
	}
}
