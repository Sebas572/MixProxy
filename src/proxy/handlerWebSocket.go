package proxy

import (
	"crypto/tls"
	"log"
	"mixproxy/src/proxy/tools"
	"mixproxy/src/redis"
	"strings"

	fws "github.com/fasthttp/websocket"
	"github.com/gofiber/websocket/v2"
)

func getSubdomainFromWebSocket(ctx *websocket.Conn) string {
	origin := ctx.Headers("Origin")
	originSplit := strings.Split(origin, "://")[1]

	if originSplit == cfg.Hostname {
		return ""
	}

	subdomain := strings.Split(strings.Split(origin, "://")[1], "."+cfg.Hostname)[0]

	return subdomain
}

func getHandleFuncFromWebSocket(ctx *websocket.Conn) (string, error) {
	subdomain := getSubdomainFromWebSocket(ctx)

	target, err := tools.GetTargetIPForSubdomain(subdomain)
	if err != nil {
		return "", err
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

	subdomain := getSubdomainFromWebSocket(c)

	// Check global blacklist
	_, err = redis.GetIPForGlobalBlacklist(c.RemoteAddr().String())
	if err == nil {
		c.WriteMessage(websocket.CloseMessage, []byte("You are on the global blacklist"))
		return
	}

	// Check subdomain blacklist
	if isEnabled, _ := redis.IsEnabledBlacklistForSubdomain(subdomain); isEnabled {
		_, err = redis.GetIPForBlacklist(subdomain, c.RemoteAddr().String())
		if err == nil {
			c.WriteMessage(websocket.CloseMessage, []byte("You are on the blacklist"))
			return
		}
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
