import logging
import os

import requests


class TelegramSender:
    def __init__(self) -> None:
        self._bot_token = os.environ["TELEGRAM_BOT_TOKEN"]
        self._chat_id = os.environ["TELEGRAM_CHAT_ID"]
        self._api_url = f"https://api.telegram.org/bot{self._bot_token}/sendMessage"

    def send(self, text: str) -> None:
        resp = requests.post(
            self._api_url,
            json={"chat_id": self._chat_id, "text": text},
            timeout=10,
        )
        resp.raise_for_status()
        logging.info("telegram notification sent")
