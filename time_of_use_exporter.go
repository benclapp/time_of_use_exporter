package main

import (
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

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

	// t := "00:00"
	// pt, err := time.ParseInLocation("05:04", t, time.UTC)
	// if err != nil {
	// 	slog.Error("error parsing time", "t", t, "err", err)
	// 	os.Exit(1)
	// }
	// slog.Info("Current time", "time", time.Now().Format(time.TimeOnly))
	// slog.Info("parsed time", "t", t, "pt", pt)

	loc, err := time.LoadLocation("Pacific/Auckland")
	if err != nil {
		slog.Error("error loading timezone", "err", err)
	}
	now := time.Now().In(loc)
	baseDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	slog.Info("now", "now", now, "baseDay", baseDay)

	start, err := time.ParseDuration("7h12m")
	if err != nil {
		slog.Error("error parsing duration", "err", err)
	}
	end, err := time.ParseDuration("22h27m5s")
	if err != nil {
		slog.Error("error parsing duration", "err", err)
	}
	startTime := baseDay.Add(start)
	endTime := baseDay.Add(end)
	slog.Info("start/end", "start", startTime, "end", endTime)

	if now.After(startTime) && now.Before(endTime) {
		fmt.Println("SUCCESS - WITHIN TIME WINDOW!!!")
		// All works well
		// Just need to handle resetitng the start/end time each day.
		// Maybe trigger if now.After(endTime), then add 24 * time.Hour to both start/end?

	}

	// os.Exit(0)

	configInit()
	// Loop to check config updates
	go func() {
		for {
			slog.Info("read config", "cfg", liveConfig)
			time.Sleep(1 * time.Second)
		}
	}()

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
