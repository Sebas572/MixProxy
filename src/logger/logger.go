package logger

import (
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/diode"
)

var Log zerolog.Logger

func init() {
	// O_APPEND: agrega al final, O_CREATE: lo crea si no existe, O_WRONLY: solo escritura.
	file, err := os.OpenFile(
		"./logs/requests.log",
		os.O_APPEND|os.O_CREATE|os.O_WRONLY,
		0664,
	)
	if err != nil {
		panic("The log file could not be opened: " + err.Error())
	}

	// Usamos DIODE para que la escritura al DISCO sea asíncrona.
	// - 1000 es el tamaño del buffer de mensajes.
	// - 10ms es el intervalo de "flush".
	// - El cuarto parámetro es una función que avisa si se perdieron logs por saturación.
	wr := diode.NewWriter(file, 10000, 10*time.Millisecond, func(dropped int) {
		os.Stderr.WriteString("Alert: Logs discarded due to slow disk\n")
	})

	Log = zerolog.New(wr).With().Timestamp().Logger()
}

func AddRequestLog(method, url, ip, subdomain string, statusCode int, withCache bool) {
	Log.Info().
		Str("method", method).
		Str("url", url).
		Str("ip", ip).
		Str("sub", subdomain).
		Int("status", statusCode).
		Bool("cache", withCache).
		Send()
}
