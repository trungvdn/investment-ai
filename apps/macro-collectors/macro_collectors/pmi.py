import hashlib
import re
import subprocess
from datetime import datetime, timezone
from html.parser import HTMLParser
from urllib.parse import urljoin, urlsplit, urlunsplit

from .core import MonthlySchedule, atomic_write, target_period
from .http import CollectionError, HTTPClient, NotPublished

INDEX = "https://www.pmi.spglobal.com/Public/Release/PressReleases?language=en"
HOSTS = {"www.pmi.spglobal.com", "pmi.spglobal.com"}
MONTHS = ("January February March April May June July August September October November December").split()


def official_url(url, index=False):
    parsed = urlsplit(url)
    return parsed.scheme == "https" and parsed.netloc in HOSTS and not parsed.fragment and (
        parsed.path == "/Public/Release/PressReleases" and parsed.query in ("", "language=en") if index else re.fullmatch(r"/Public/Home/PressRelease/[a-fA-F0-9]{32}", parsed.path) is not None and not parsed.query
    )


class ReleaseLinks(HTMLParser):
    def __init__(self, base):
        super().__init__()
        self.base, self.active, self.links = base, None, []

    def handle_starttag(self, tag, attrs):
        if tag == "a":
            attrs = dict(attrs)
            self.active = [attrs.get("href", ""), attrs.get("title", "")]

    def handle_data(self, text):
        if self.active is not None:
            self.active[1] += " " + text

    def handle_endtag(self, tag):
        if tag != "a" or self.active is None:
            return
        href, text = self.active
        self.active = None
        text = text.lower()
        if not ("vietnam" in text or "viet nam" in text) or "manufacturing" not in text:
            return
        resolved = urlsplit(urljoin(self.base, href))
        url = urlunsplit((resolved.scheme, resolved.netloc, resolved.path, "", ""))
        if official_url(url) and url not in self.links:
            self.links.append(url)


def pdf_text(path):
    try:
        result = subprocess.run(["pdftotext", "-f", "1", "-l", "1", "-raw", str(path), "-"], check=True, capture_output=True, timeout=15)
    except (OSError, subprocess.SubprocessError):
        raise CollectionError("DATA_QUALITY", "PMI PDF extraction failed; install poppler-utils") from None
    if len(result.stdout) > 1024 * 1024:
        raise CollectionError("DATA_QUALITY", "PMI extracted text too large")
    return result.stdout.decode("utf-8")


class PMICollector:
    schedule = MonthlySchedule()
    target = staticmethod(target_period)

    def __init__(self, index=INDEX, client_factory=HTTPClient, extractor=pdf_text):
        if not official_url(index, index=True):
            raise ValueError("PMI index must be the official S&P public release list")
        self.index, self.client_factory, self.extractor = index, client_factory, extractor

    def collect(self, directory, target, now):
        client = self.client_factory()
        html = client.get(self.index)
        atomic_write(directory / "index.html", html)
        parser = ReleaseLinks(self.index)
        parser.feed(html.decode("utf-8"))
        if not parser.links:
            raise CollectionError("DATA_QUALITY", "No explicit Vietnam manufacturing release links; source layout may have changed")
        year, month = map(int, target.split("-"))
        report_month = f"{MONTHS[month - 1]} {year}"
        for url in parser.links[:6]:
            body = client.get(url)
            if not body.startswith(b"%PDF-"):
                raise CollectionError("DATA_QUALITY", "PMI release is not a PDF document")
            path = directory / "raw" / (hashlib.sha256(body).hexdigest() + ".pdf")
            if not path.exists():
                atomic_write(path, body)
            text = " ".join(self.extractor(path).split())
            report_text = re.sub(r"Embargoed until [^(]+\([^)]*\)", "", text, flags=re.I)
            if report_month.lower() not in report_text.lower():
                continue
            if not re.search(r"S&P Global Viet\s*nam Manufacturing", text, re.I):
                raise CollectionError("DATA_QUALITY", "PMI document does not identify Vietnam manufacturing")
            embargo = re.search(r"Embargoed until (\d{1,2} [A-Za-z]+ \d{4}) \((\d{4}) ICT", text, re.I)
            if not embargo:
                raise CollectionError("DATA_QUALITY", "PMI publication embargo is missing")
            from .core import ZONE
            try:
                published = datetime.strptime(" ".join(embargo.groups()), "%d %B %Y %H%M").replace(tzinfo=ZONE)
            except ValueError:
                raise CollectionError("DATA_QUALITY", "PMI embargo timestamp is invalid") from None
            if published > now:
                raise NotPublished()
            # The Go importer remains the authoritative headline/value parser.
            return {"releases": [{"path": str(path), "sourceUrl": url}], "indexUrl": self.index}
        raise NotPublished()
