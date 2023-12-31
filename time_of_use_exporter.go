package main

import (
	"log"
	"log/slog"
	"net/http"
	"os"

	// Import timezone data as a backup if not provided by the OS
	// For example, in the Dockerfile built from scratch, if host OS path isn't mounted
	_ "time/tzdata"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Exporter struct{}

func main() {
	var logLevel slog.Level
	switch os.Getenv("LOG_LEVEL") {
	case "debug":
		logLevel = slog.LevelDebug
	case "info":
		logLevel = slog.LevelInfo
	case "warning":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}

	slog.SetDefault(slog.New(slog.NewTextHandler(
		os.Stdout, &slog.HandlerOptions{
			AddSource: true,
			Level:     logLevel,
		})))

	configInit()

	prometheus.MustRegister(&Exporter{})

	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
			<head><title>Time of Use Exporter</title></head>
			<body>
			<h1>Time of Use Exporter</h1>
			<p><a href=/metrics>Metrics</a></p>
			</body>
			</html>`))
	})
	addr := os.Getenv("LISTEN_ADDR")
	if addr == "" {
		addr = ":10007"
	}
	slog.Info("Starting Time of Use Exporter", "LISTEN_ADDR", addr, "LOG_LEVEL", logLevel)
	log.Fatal(http.ListenAndServe(addr, nil))
}
