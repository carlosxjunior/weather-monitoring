package collector

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/bitlabz/weather-monitoring/collector/internal/cities"
	"github.com/bitlabz/weather-monitoring/collector/internal/publisher"
	"github.com/bitlabz/weather-monitoring/collector/internal/weather"
)

const sendTimeout = 15 * time.Second

type collectionResult struct {
	city       string
	metricType string
	records    int
	err        error
}

type callSummary struct {
	MetricType string `json:"metric_type"`
	Status     string `json:"status"`
	Records    int    `json:"records"`
	Error      string `json:"error,omitempty"`
}

type citySummary struct {
	City   string        `json:"city"`
	Status string        `json:"status"`
	Calls  []callSummary `json:"calls"`
}

// WeatherCollector collects weather and air-quality data for all cities and
// publishes each payload to Service Bus (or prints it when mockPublish is true).
type WeatherCollector struct {
	forecast    *weather.ForecastClient
	airQuality  *weather.AirQualityClient
	pub         *publisher.ServiceBusPublisher // nil when mockPublish is true
	runID       string
	mockPublish bool
}

// NewWeatherCollector builds the full collector, including HTTP client and
// Service Bus publisher. Returns an error if required env vars are missing or
// Azure SDK initialisation fails.
func NewWeatherCollector(runID string, mockPublish bool) (*WeatherCollector, error) {
	httpClient := &http.Client{Timeout: 15 * time.Second}

	var pub *publisher.ServiceBusPublisher
	if !mockPublish {
		sbNamespace, err := requireEnv("SERVICEBUS_NAMESPACE")
		if err != nil {
			return nil, err
		}
		sbQueue, err := requireEnv("SERVICEBUS_QUEUE_NAME")
		if err != nil {
			return nil, err
		}
		pub, err = publisher.New(sbNamespace, sbQueue)
		if err != nil {
			return nil, fmt.Errorf("create publisher: %w", err)
		}
	}

	return &WeatherCollector{
		forecast:    weather.NewForecastClient(httpClient),
		airQuality:  weather.NewAirQualityClient(httpClient),
		pub:         pub,
		runID:       runID,
		mockPublish: mockPublish,
	}, nil
}

// Run implements Collector. It returns a non-nil error only if every city
// failed for every metric type. ctx is the job-level context (30s overall).
func (w *WeatherCollector) Run(ctx context.Context) error {
	if w.pub != nil {
		defer w.pub.Close(context.Background())
	}

	type task struct {
		city       cities.City
		metricType string
	}

	tasks := make([]task, 0, len(cities.All)*2)
	for _, city := range cities.All {
		tasks = append(tasks, task{city, "weather"})
		tasks = append(tasks, task{city, "air_quality"})
	}

	results := make(chan collectionResult, len(tasks))
	var wg sync.WaitGroup

	for _, t := range tasks {
		wg.Add(1)
		go func(t task) {
			defer wg.Done()
			w.runTask(ctx, t.city, t.metricType, results)
		}(t)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	return w.aggregate(results)
}

// runTask executes one fetch + publish cycle for a single city/metric pair.
func (w *WeatherCollector) runTask(
	ctx context.Context,
	city cities.City,
	metricType string,
	results chan<- collectionResult,
) {
	fetchCtx, fetchCancel := context.WithTimeout(ctx, 15*time.Second)
	defer fetchCancel()

	var payload any
	var fetchErr error

	switch metricType {
	case "weather":
		payload, fetchErr = w.forecast.Fetch(fetchCtx, city, w.runID)
	case "air_quality":
		payload, fetchErr = w.airQuality.Fetch(fetchCtx, city, w.runID)
	}

	if fetchErr != nil {
		slog.Warn("fetch failed",
			slog.String("run_id", w.runID),
			slog.String("city", city.Name),
			slog.String("metric_type", metricType),
			slog.String("error", fetchErr.Error()),
		)
		results <- collectionResult{city: city.Name, metricType: metricType, err: fetchErr}
		return
	}

	var publishErr error
	if w.mockPublish {
		b, _ := json.Marshal(payload)
		fmt.Printf("%s\n", b)
	} else {
		// sendCtx is derived from ctx (the 30s job context), not fetchCtx.
		// Effective deadline = min(15s from now, remaining job time).
		sendCtx, sendCancel := context.WithTimeout(ctx, sendTimeout)
		defer sendCancel()
		publishErr = w.pub.Send(sendCtx, payload)
		if publishErr != nil {
			slog.Error("publish failed",
				slog.String("run_id", w.runID),
				slog.String("city", city.Name),
				slog.String("metric_type", metricType),
				slog.String("error", publishErr.Error()),
			)
		}
	}

	results <- collectionResult{
		city:       city.Name,
		metricType: metricType,
		records:    1,
		err:        publishErr,
	}
}

// aggregate drains the results channel, builds the structured log summary,
// and returns an error if every city failed.
func (w *WeatherCollector) aggregate(results <-chan collectionResult) error {
	cityMap := make(map[string][]callSummary)
	for r := range results {
		cs := callSummary{MetricType: r.metricType, Records: r.records}
		if r.err != nil {
			cs.Status = "failed"
			cs.Error = r.err.Error()
		} else {
			cs.Status = "success"
		}
		cityMap[r.city] = append(cityMap[r.city], cs)
	}

	summaries := make([]citySummary, 0, len(cities.All))
	successCities := 0

	for _, city := range cities.All {
		calls, ok := cityMap[city.Name]
		if !ok {
			continue
		}
		failed := 0
		for _, c := range calls {
			if c.Status == "failed" {
				failed++
			}
		}
		status := "success"
		switch {
		case failed == len(calls):
			status = "failed"
		case failed > 0:
			status = "partial"
		}
		if status != "failed" {
			successCities++
		}
		summaries = append(summaries, citySummary{City: city.Name, Status: status, Calls: calls})
	}

	slog.Info("run completed",
		slog.String("run_id", w.runID),
		slog.Any("cities", summaries),
	)

	if successCities == 0 {
		return fmt.Errorf("all cities failed")
	}
	return nil
}

func requireEnv(key string) (string, error) {
	v := os.Getenv(key)
	if v == "" {
		return "", fmt.Errorf("required environment variable %q is not set", key)
	}
	return v, nil
}
