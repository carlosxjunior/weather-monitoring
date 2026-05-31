# Weather monitoring

Simulates a production monitoring pipeline using Azure-native services.

A scheduled **Go Container App Job** fetches live weather and air quality data from [Open Meteo](https://open-meteo.com) for 10 cities concurrently, publishing each reading to an **Azure Service Bus** topic. A **Python Function App** consumes those messages and sends a **Telegram** notification per reading.

## Architecture

```
Container App Job (Go)          Azure Service Bus          Function App (Python)
┌─────────────────────┐        ┌──────────────┐          ┌────────────────────┐
│  10 cities × 2 APIs │        │ Topic:       │          │ Service Bus trigger│
│  = 20 goroutines    │──────▶   weather-      ───────▶ │                    │──▶ Telegram
│                     │        │ readings     │          │ integrations/      │
│  Weather Forecast   │        │              │          │   telegram/        │
│  Air Quality        │        │ Subscription:│          └────────────────────┘
└─────────────────────┘        │ notifier-sub │
                               └──────────────┘
```

Every 5 minutes, 20 messages are published (2 per city) and 20 Telegram notifications are sent.

## Components

| Path | Language | Azure resource |
|---|---|---|
| `collector/` | Go | Container App Job (`ca-weather-collector-job`) |
| `notifier/` | Python | Function App (`func-weather-notifier`) |
| `.devops/` | YAML | Azure DevOps pipelines |
| `infra/` | — | Terraform placeholder |

## Why Go and Python?

**Go — collector**

The collector is a fan-out workload: 10 cities × 2 APIs = 20 HTTP calls that must complete within a tight deadline before the job exits. Go's goroutines make this trivial — lightweight, cheap to spawn, and backed by a runtime scheduler that handles the fan-out without thread-per-request overhead. This is the same reason Go dominates production monitoring and observability tooling (Prometheus, Grafana Agent, OpenTelemetry Collector are all written in Go).

**Python — notifier**

The notifier has the opposite profile: it receives one message at a time and does one thing with it (format and send a Telegram notification). There is no concurrency requirement and no performance-sensitive path. Python is the right call here because virtually every engineering team already knows it, the Azure Functions Python SDK is mature, and the code stays readable without any language-specific expertise. In a real team setting, anyone can open `function_app.py` and understand or modify it without onboarding friction.

Python is also the de facto language for integrations and glue code — which is exactly what the notifier is.

## Local development

### Collector (offline mode)

```powershell
cd collector
go mod tidy
$env:MOCK_PUBLISH = "true"
go run ./cmd/collector
```

### Collector (with real Service Bus)

```powershell
$env:SERVICEBUS_NAMESPACE = "sb-weather-collector.servicebus.windows.net"
$env:SERVICEBUS_TOPIC_NAME = "weather-readings"
$env:AZURE_CLIENT_ID = "<user-assigned-mi-client-id>"
go run ./cmd/collector
```

### Notifier

```powershell
cd notifier
python -m venv .venv
.venv\Scripts\Activate.ps1
pip install -r requirements.txt
pip install azure-functions
# create local.settings.json from local.settings.json.example
func start
```

## Azure resources

| Resource | Name | SKU |
|---|---|---|
| Resource Group | `rg-weather-collector` | — |
| Service Bus Namespace | `sb-weather-collector` | Standard |
| Service Bus Topic | `weather-readings` | — |
| Container Registry | `acrweathercollector` | Basic |
| Container Apps Environment | `cae-weather-collector` | Consumption |
| Container App Job | `ca-weather-collector-job` | Scheduled (`*/5 * * * *`) |
| Function App | `func-weather-notifier` | Consumption Y1, Linux, Python 3.11 |
| Key Vault | `kv-weather-collector` | Standard |
| Managed Identity (collector) | `id-weather-collector` | User-assigned |

Region: West Europe
