import json
import logging

import azure.functions as func

from integrations.telegram.formatter import format_reading
from integrations.telegram.sender import TelegramSender

app = func.FunctionApp()

@app.service_bus_queue_trigger(
    arg_name="message",
    queue_name="sbq-collector",
    connection="ServiceBusConnection",
)
def notify_weather(message: func.ServiceBusMessage) -> None:
    reading = json.loads(message.get_body().decode("utf-8"))
    logging.info("received message | run_id=%s city=%s metric=%s",
                 reading.get("run_id"), reading.get("city"), reading.get("metric_type"))
    TelegramSender().send(format_reading(reading))
