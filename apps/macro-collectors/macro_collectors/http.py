import http.client
import random
import time
import urllib.error
import urllib.request
from datetime import datetime, timezone
from email.utils import parsedate_to_datetime


class CollectionError(Exception):
    def __init__(self, kind, message, retry_after=0):
        super().__init__(message)
        self.kind, self.retry_after = kind, retry_after


class NotPublished(Exception):
    pass


def retry_seconds(value, now=None):
    try:
        return max(0, min(365 * 86400, int(value)))
    except (ValueError, TypeError):
        try:
            return max(0, min(365 * 86400, (parsedate_to_datetime(value) - (now or datetime.now(timezone.utc))).total_seconds()))
        except (TypeError, ValueError, OverflowError):
            return 0


class HTTPClient:
    def __init__(self, deadline=180, opener=urllib.request.urlopen, sleep=time.sleep):
        self.expires = time.monotonic() + deadline
        self.opener, self.sleep = opener, sleep

    def get(self, url):
        for attempt in range(3):
            remaining = self.expires - time.monotonic()
            if remaining <= 0:
                raise CollectionError("TIMEOUT", "Collector exceeded its job deadline")
            request = urllib.request.Request(url, headers={"User-Agent": "InvestmentAI-Collectors/1.0 (+https://github.com/trungvdn/investment-ai)"})
            delay = 2 ** attempt + random.random()
            try:
                with self.opener(request, timeout=min(30, remaining)) as response:
                    body = response.read(16 * 1024 * 1024 + 1)
                    if len(body) > 16 * 1024 * 1024:
                        raise CollectionError("DATA_QUALITY", "Source document exceeds 16 MiB")
                    return body
            except urllib.error.HTTPError as error:
                status = error.code
                retry_after = retry_seconds(error.headers.get("Retry-After"))
                error.close()
                kind = "ACCESS_DENIED" if status in (401, 403) else "RATE_LIMIT" if status == 429 else "HTTP"
                failure = CollectionError(kind, f"Public source returned HTTP {status}", retry_after)
                if status != 429 and status < 500:
                    raise failure from None
                delay = max(delay, retry_after)
            except (urllib.error.URLError, TimeoutError, OSError, http.client.HTTPException):
                failure = CollectionError("TRANSPORT", "Public source connection/read failed")
            if attempt == 2 or delay > 30 or time.monotonic() + delay >= self.expires:
                raise failure from None
            self.sleep(delay)
