package models

import "time"

type WeatherMessage struct {
	RunID           string    `json:"run_id"`
	City            string    `json:"city"`
	MetricType      string    `json:"metric_type"`
	Latitude        float64   `json:"latitude"`
	Longitude       float64   `json:"longitude"`
	TemperatureC    float64   `json:"temperature_c"`
	WindSpeedKmh    float64   `json:"wind_speed_kmh"`
	HumidityPct     float64   `json:"humidity_pct"`
	PrecipitationMm float64   `json:"precipitation_mm"`
	CollectedAt     time.Time `json:"collected_at"`
}

type AirQualityMessage struct {
	RunID              string    `json:"run_id"`
	City               string    `json:"city"`
	MetricType         string    `json:"metric_type"`
	Latitude           float64   `json:"latitude"`
	Longitude          float64   `json:"longitude"`
	PM25               float64   `json:"pm2_5"`
	PM10               float64   `json:"pm10"`
	CarbonMonoxideUgM3 float64   `json:"carbon_monoxide_ug_m3"`
	NitrogenDioxideUgM3 float64  `json:"nitrogen_dioxide_ug_m3"`
	OzoneUgM3          float64   `json:"ozone_ug_m3"`
	CollectedAt        time.Time `json:"collected_at"`
}
