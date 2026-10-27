"""Trading Economics' public Vietnam PMI summary, isolated shadow collection."""
import calendar
import re
from datetime import date
from html.parser import HTMLParser

from .http import CollectionError

URL = "https://tradingeconomics.com/vietnam/manufacturing-pmi"
VERSION = "te_vietnam_pmi_v1"
MONTHS = {name.lower(): i for i, name in enumerate(calendar.month_name) if i}
MONTH = "(?:" + "|".join(MONTHS) + ")"


class SummaryHTML(HTMLParser):
    def __init__(self):
        super().__init__()
        self.headings = []
        self.active = None
        self.parts = []
        self.skip = 0

    def handle_starttag(self, tag, attrs):
        if tag in ("script", "style"):
            self.skip += 1
        if not self.skip and tag in ("h1", "h2"):
            self.active, self.parts = tag, []

    def handle_data(self, data):
        if self.active and not self.skip:
            self.parts.append(data)

    def handle_endtag(self, tag):
        if tag in ("script", "style"):
            self.skip = max(0, self.skip - 1)
        if tag == self.active:
            self.headings.append((tag, " ".join(" ".join(self.parts).split())))
            self.active = None


def extract(body, today):
    try:
        parser = SummaryHTML()
        parser.feed(body.decode("utf-8"))
        if [text.lower() for tag, text in parser.headings if tag == "h1"] != ["vietnam manufacturing pmi"]:
            raise ValueError("Unexpected page")
        # Only the published current/previous summary, never forecasts or news history.
        pattern = (r"^Manufacturing PMI in Vietnam (?:increased|decreased|rose|fell|was unchanged|remained unchanged|held steady) "
                   r"(?:to|at) (\d+(?:\.\d+)?) points in (" + MONTH + r")(?: of (20\d{2}))? "
                   r"from (\d+(?:\.\d+)?) points in (" + MONTH + r") of (20\d{2})\.")
        candidates = []
        for tag, text in parser.headings:
            if tag != "h2":
                continue
            match = re.match(pattern, text, re.I)
            if not match:
                continue
            value, month, year, previous, previous_month, previous_year = match.groups()
            month, previous_month, previous_year = MONTHS[month.lower()], MONTHS[previous_month.lower()], int(previous_year)
            year = int(year) if year else previous_year + (month < previous_month)
            end = date(year, month, calendar.monthrange(year, month)[1])
            prev_year, prev_month = (year - 1, 12) if month == 1 else (year, month - 1)
            if (previous_year, previous_month) != (prev_year, prev_month) or end > today:
                raise ValueError("Inconsistent reference dates")
            value, previous = float(value), float(previous)
            if not 0 <= value <= 100 or not 0 <= previous <= 100:
                raise ValueError("PMI outside definition")
            candidates.append({"period": f"{year:04d}-{month:02d}", "period_end": end, "value": value,
                               "source_url": URL, "published_at": None, "basis": "INDEX",
                               "evidence": {"text": match[0], "publisher": "Trading Economics", "originalProducer": "S&P Global",
                                            "previousValue": previous, "previousPeriod": f"{prev_year:04d}-{prev_month:02d}",
                                            "publicationPrecision": "unknown"}})
        if not candidates or len({(row["period"], row["value"]) for row in candidates}) != 1:
            raise ValueError("Missing or conflicting actual summary")
        return [candidates[0]]
    except (ValueError, UnicodeError, KeyError):
        error = CollectionError("DATA_QUALITY", "Trading Economics PMI lacks an unambiguous actual value/reference month")
        error.payload = body
        raise error from None


def normalize(client, dataset, config, today):
    if dataset["url"] != URL:
        raise CollectionError("DATA_QUALITY", "Unexpected Trading Economics PMI URL")
    if len(config) != 1 or config[0]["extractor_key"] != VERSION or config[0]["transform_key"] != "published_index" or config[0]["indicator_code"] != "VN_PMI" or config[0]["unit"] != "index":
        raise CollectionError("DATA_QUALITY", "Unsupported Trading Economics PMI binding")
    body = client.get(URL)
    binding = config[0]
    rows = extract(body, today)
    return body, [{**row, **{key: binding[key] for key in ("indicator_code", "series_id", "unit", "priority")},
                   "binding_id": binding["id"], "method": "PUBLISHED_PMI_INDEX; SECONDARY=Trading Economics; PARSER=" + VERSION} for row in rows]
