package main

import (
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Exporter struct{}

func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(
		os.Stdout, &slog.HandlerOptions{
			AddSource: true,
			Level:     slog.LevelInfo,
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

	log.Fatal(http.ListenAndServe(":8080", nil))
}
