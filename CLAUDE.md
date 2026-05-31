# Weather monitoring

This project simulates a production monitoring pipeline where multiple metrics are collected from multiple systems and forwarded to an observability or notification tool.

## What it does

A scheduled **collector job** fetches metrics from multiple sources concurrently and publishes each reading as a message to **Azure Service Bus**. A **Function App** subscribes to those messages and forwards the data to a destination — in production this would be a monitoring platform (Datadog, Grafana, Dynatrace) or a notification channel (Slack, Teams). Here we use **Telegram** for simplicity.

## Components

| Component | Technology | Role |
|---|---|---|
| `collector/` | Go, Azure Container App Job | Fetches metrics concurrently, publishes to Service Bus |
| `notifier/` | Python, Azure Function App | Consumes Service Bus messages, sends to destination |
| `.devops/` | Azure DevOps YAML | CI/CD pipelines for both components |
| `infra/` | Terraform (future) | Infrastructure as code |

## Simulation

Instead of real monitored systems, this project fetches live weather and air quality data from [Open Meteo](https://open-meteo.com) for 10 cities — each city represents a monitored system. Two metrics are collected per city per run (weather + air quality), producing 20 messages to Service Bus every 5 minutes.

## Local development

```bash
# Collector — offline mode (no Azure required)
cd collector
MOCK_PUBLISH=true go run ./cmd/collector

# Notifier
cd notifier
pip install -r requirements.txt azure-functions
func start
```

See `README.md` for full setup instructions and Azure resource names.
