package main

import (
	"context"
	cryptorand "crypto/rand"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/bitlabz/weather-monitoring/collector/internal/collector"
)

const jobTimeout = 30 * time.Second

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	runID := newRunID()
	slog.Info("run started", slog.String("run_id", runID))

	mockPublish := os.Getenv("MOCK_PUBLISH") == "true"
	collectorType := os.Getenv("COLLECTOR_TYPE")

	var col collector.Collector
	var err error
	switch collectorType {
	case "weather":
		col, err = collector.NewWeatherCollector(runID, mockPublish)
		if err != nil {
			slog.Error("init collector", slog.String("error", err.Error()))
			os.Exit(1)
		}
	default:
		slog.Error("unknown COLLECTOR_TYPE", slog.String("type", collectorType))
		os.Exit(1)
	}

	parentCtx, cancel := context.WithTimeout(context.Background(), jobTimeout)
	defer cancel()

	if err := col.Run(parentCtx); err != nil {
		slog.Error("run failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
}

func newRunID() string {
	var b [16]byte
	_, _ = cryptorand.Read(b[:])
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}
