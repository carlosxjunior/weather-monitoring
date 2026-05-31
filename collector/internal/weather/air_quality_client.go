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

type AirQualityClient struct {
	http *http.Client
}

func NewAirQualityClient(httpClient *http.Client) *AirQualityClient {
	return &AirQualityClient{http: httpClient}
}

func (c *AirQualityClient) Fetch(ctx context.Context, city cities.City, runID string) (*models.AirQualityMessage, error) {
	url := fmt.Sprintf(
		"https://air-quality-api.open-meteo.com/v1/air-quality?latitude=%f&longitude=%f&current=pm2_5,pm10,carbon_monoxide,nitrogen_dioxide,ozone&timezone=auto",
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
			PM25            float64 `json:"pm2_5"`
			PM10            float64 `json:"pm10"`
			CarbonMonoxide  float64 `json:"carbon_monoxide"`
			NitrogenDioxide float64 `json:"nitrogen_dioxide"`
			Ozone           float64 `json:"ozone"`
		} `json:"current"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &models.AirQualityMessage{
		RunID:               runID,
		City:                city.Name,
		MetricType:          "air_quality",
		Latitude:            city.Latitude,
		Longitude:           city.Longitude,
		PM25:                body.Current.PM25,
		PM10:                body.Current.PM10,
		CarbonMonoxideUgM3:  body.Current.CarbonMonoxide,
		NitrogenDioxideUgM3: body.Current.NitrogenDioxide,
		OzoneUgM3:           body.Current.Ozone,
		CollectedAt:         time.Now().UTC(),
	}, nil
}
