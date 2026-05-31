def format_reading(reading: dict) -> str:
    city = reading.get("city", "Unknown")
    metric_type = reading.get("metric_type", "")
    collected_at = reading.get("collected_at", "")

    if metric_type == "weather":
        return (
            f"Weather — {city}\n"
            f"Temp: {reading.get('temperature_c')}°C\n"
            f"Wind: {reading.get('wind_speed_kmh')} km/h\n"
            f"Humidity: {reading.get('humidity_pct')}%\n"
            f"Precipitation: {reading.get('precipitation_mm')} mm\n"
            f"At: {collected_at}"
        )

    if metric_type == "air_quality":
        return (
            f"Air Quality — {city}\n"
            f"PM2.5: {reading.get('pm2_5')} µg/m³\n"
            f"PM10: {reading.get('pm10')} µg/m³\n"
            f"CO: {reading.get('carbon_monoxide_ug_m3')} µg/m³\n"
            f"NO₂: {reading.get('nitrogen_dioxide_ug_m3')} µg/m³\n"
            f"O₃: {reading.get('ozone_ug_m3')} µg/m³\n"
            f"At: {collected_at}"
        )

    return f"Reading — {city} ({metric_type})\n{reading}"
