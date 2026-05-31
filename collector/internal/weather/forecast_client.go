package weather

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/bitlabz/weather-monitoring/collector/internal/cities"
	"github.com/bitlabz/weather-monitoring/collector/internal/models"
)

type ForecastClient struct {
	http *http.Client
}

func NewForecastClient(httpClient *http.Client) *ForecastClient {
	return &ForecastClient{http: httpClient}
}

func (c *ForecastClient) Fetch(ctx context.Context, city cities.City, runID string) (*models.WeatherMessage, error) {
	url := fmt.Sprintf(
		"https://api.open-meteo.com/v1/forecast?latitude=%f&longitude=%f&current=temperature_2m,wind_speed_10m,relative_humidity_2m,precipitation&timezone=auto",
		city.Latitude, city.Longitude,
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d", resp.StatusCode)
	}

	var body struct {
		Current struct {
			Temperature2m      float64 `json:"temperature_2m"`
			WindSpeed10m       float64 `json:"wind_speed_10m"`
			RelativeHumidity2m float64 `json:"relative_humidity_2m"`
			Precipitation      float64 `json:"precipitation"`
		} `json:"current"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &models.WeatherMessage{
		RunID:           runID,
		City:            city.Name,
		MetricType:      "weather",
		Latitude:        city.Latitude,
		Longitude:       city.Longitude,
		TemperatureC:    body.Current.Temperature2m,
		WindSpeedKmh:    body.Current.WindSpeed10m,
		HumidityPct:     body.Current.RelativeHumidity2m,
		PrecipitationMm: body.Current.Precipitation,
		CollectedAt:     time.Now().UTC(),
	}, nil
}
