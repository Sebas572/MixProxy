package logger

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/diode"
)

var Log zerolog.Logger
var rotatingWriter *DailyRotatingWriter

type DailyRotatingWriter struct {
	dir         string
	currentFile *os.File
	currentDate string
	mu          sync.Mutex
}

func NewDailyRotatingWriter(dir string) *DailyRotatingWriter {
	return &DailyRotatingWriter{dir: dir}
}

func (w *DailyRotatingWriter) Write(p []byte) (n int, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	now := time.Now()
	dateStr := now.Format("2006-01-02")

	if w.currentDate != dateStr {
		if w.currentFile != nil {
			w.currentFile.Close()
		}
		filename := fmt.Sprintf("%s/log-%s", w.dir, dateStr)
		file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0664)
		if err != nil {
			return 0, err
		}
		w.currentFile = file
		w.currentDate = dateStr
	}

	return w.currentFile.Write(p)
}

func (w *DailyRotatingWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.currentFile != nil {
		return w.currentFile.Close()
	}
	return nil
}

func (w *DailyRotatingWriter) CurrentFileName() string {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.currentDate == "" {
		return ""
	}
	return fmt.Sprintf("%s/log-%s", w.dir, w.currentDate)
}

func init() {
	// Ensure logs directory exists
	if err := os.MkdirAll("./logs", 0755); err != nil {
		panic("Could not create logs directory: " + err.Error())
	}

	rotatingWriter = NewDailyRotatingWriter("./logs")

	// Usamos DIODE para que la escritura al DISCO sea asíncrona.
	// - 10000 es el tamaño del buffer de mensajes.
	// - 10ms es el intervalo de "flush".
	// - El cuarto parámetro es una función que avisa si se perdieron logs por saturación.
	wr := diode.NewWriter(rotatingWriter, 10000, 10*time.Millisecond, func(dropped int) {
		os.Stderr.WriteString("Alert: Logs discarded due to slow disk\n")
	})

	Log = zerolog.New(wr).With().Timestamp().Logger()
}

func GetCurrentLogFile() string {
	if rotatingWriter == nil {
		return ""
	}
	return rotatingWriter.CurrentFileName()
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
