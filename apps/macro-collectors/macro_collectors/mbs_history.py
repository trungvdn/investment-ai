"""Bounded public MBS report discovery and explicit monthly PMI extraction."""
import base64
import calendar
import hashlib
import json
import re
import subprocess
import tempfile
from datetime import date, datetime
from pathlib import Path
from urllib.parse import urljoin, urlsplit

from .core import ZONE
from .cpi_reports import ArticleHTML, fold
from .http import CollectionError

INDEX = "https://www.mbs.com.vn/trien-vong-vi-mo/"
VERSION = "mbs_pmi_history_v1"
MONTH = r"(?:0?[1-9]|1[0-2])(?!\d)"
NUMBER = r"\d{1,3}(?:[.,]\d+)?"
REFERENCE = r"thang\s+("+MONTH+r")(?:\s*[/.-]\s*(20\d{2})|\s+nam\s+(20\d{2}))?"
VALUE = r"(?:dat|o muc|ve muc|len muc|xuong muc|len|xuong)\s+("+NUMBER+r")"


def allowed(url):
    p = urlsplit(url)
    return p.scheme == "https" and p.netloc in ("mbs.com.vn", "www.mbs.com.vn") and not p.fragment and not p.username


def report_url(url):
    p = urlsplit(url)
    return allowed(url) and not p.query and bool(re.search(r"(?:trien-vong|vietnamoutlook|vi-mo)", p.path, re.I)) and not p.path.lower().endswith(".pdf")


def release_period(title):
    text = fold(title)
    match = re.search(r"(?:thang\s*|t)(\d{1,2})\s*[/.-]\s*(20\d{2})", text)
    if not match or not 1 <= int(match[1]) <= 12:
        raise CollectionError("DATA_QUALITY", "MBS report has no explicit issue month/year")
    return int(match[2]), int(match[1])


def extract_text(text, report_year, report_month, today, source_url):
    result = {}
    # PDF page separators preserve the exact evidence page. Do not digitize charts.
    for page_number, page in enumerate(text.split("\f"), 1):
        for segment in re.split(r"(?<=[.!?])\s+(?=[A-ZĐ])|\n\s*\n", page):
            segment = " ".join(segment.split())
            normalized = fold(segment)
            if "pmi" not in normalized or re.search(r"\b(?:du bao|ky vong|uoc tinh|forecast|expected|projected)\b", normalized):
                continue
            for pattern in (r"\bpmi\b[^;\n]{0,100}?"+REFERENCE+r"[^;\n]{0,70}?"+VALUE,
                            REFERENCE+r"[^;\n]{0,160}?\bpmi\b[^;\n]{0,70}?"+VALUE):
                for match in re.finditer(pattern, normalized):
                    month, year_a, year_b, value = match.groups()
                    month, value = int(month), float(value.replace(",", "."))
                    year = int(year_a or year_b) if year_a or year_b else report_year - (month > report_month)
                    end = date(year, month, calendar.monthrange(year, month)[1])
                    if not 0 <= value <= 100 or end > today or (year, month) >= (report_year, report_month):
                        continue
                    period = f"{year:04d}-{month:02d}"
                    if period in result and result[period]["value"] != value:
                        raise CollectionError("CONFLICT", "Conflicting actual PMI values in MBS report")
                    result[period] = {"period": period, "period_end": end, "value": value, "source_url": source_url,
                                      "published_at": None, "evidence": {"text": segment[:700], "page": page_number,
                                      "publisher": "MBS Research", "originalProducer": "S&P Global",
                                      "referenceYearBasis": "explicit" if year_a or year_b else "report_year_with_rollover"}}
    return list(result.values())


def pdf_text(body):
    if not body.startswith(b"%PDF-") or len(body) > 16*1024*1024:
        raise CollectionError("DATA_QUALITY", "MBS download is not a bounded PDF")
    try:
        with tempfile.TemporaryDirectory() as directory:
            path = Path(directory) / "report.pdf"
            path.write_bytes(body)
            proc = subprocess.run(["pdftotext", "-f", "1", "-l", "20", "-layout", str(path), "-"],
                                  capture_output=True, check=True, timeout=20)
            if len(proc.stdout) > 4*1024*1024:
                raise CollectionError("DATA_QUALITY", "PDF text exceeds extraction budget")
            return proc.stdout.decode("utf-8")
    except (OSError, subprocess.SubprocessError, UnicodeError):
        raise CollectionError("DATA_QUALITY", "MBS PDF text extraction failed") from None


def collect(client, dataset, config, today, page=1, first=None, last=None, text_extractor=pdf_text):
    if dataset["url"] != INDEX or not 1 <= page <= 100:
        raise CollectionError("DATA_QUALITY", "Unexpected MBS history index/page")
    if len(config) != 1 or config[0]["extractor_key"] != VERSION or config[0]["indicator_code"] != "VN_PMI" or config[0]["unit"] != "index" or config[0]["transform_key"] != "published_index":
        raise CollectionError("DATA_QUALITY", "Unsupported MBS PMI binding")
    url = INDEX if page == 1 else INDEX+"page/"+str(page)+"/"
    body = client.get(url)
    parser = ArticleHTML(); parser.feed(body.decode("utf-8"))
    links = []
    for href, title in parser.links:
        target = urljoin(url, href)
        if report_url(target) and target != INDEX and not re.fullmatch(r"/trien-vong-vi-mo/page/\d+/", urlsplit(target).path):
            try: release_period(title)
            except CollectionError: continue
            if target not in links: links.append(target)
    documents = [{"url": url, "html": body.decode("utf-8")}]
    warnings, rows, seen = [], [], {}
    used_bytes = len(body)
    for target in links[:8]:
        if used_bytes > 20*1024*1024:
            warnings.append({"kind": "BYTE_BUDGET"}); break
        try:
            html = client.get(target)
            used_bytes += len(html)
            if used_bytes > 20*1024*1024:
                raise CollectionError("DATA_QUALITY", "MBS job byte budget exceeded")
            article = ArticleHTML(); article.feed(html.decode("utf-8"))
            # The shared parser exposes h1 only; MBS uses h2 for article titles.
            title = next((title for href,title in parser.links if urljoin(url,href) == target), "")
            report_year, report_month = release_period(title)
            documents.append({"url": target, "html": html.decode("utf-8")})
            candidates = extract_text("\n\n".join(article.paragraphs), report_year, report_month, today, target)
            pdfs = []
            for href, label in article.links:
                pdf = urljoin(target, href)
                if allowed(pdf) and urlsplit(pdf).path.lower().endswith(".pdf") and not urlsplit(pdf).query and pdf not in pdfs:
                    pdfs.append(pdf)
            if len(pdfs) > 1:
                raise CollectionError("DATA_QUALITY", "Ambiguous MBS report PDF")
            for pdf in pdfs:
                raw = client.get(pdf)
                used_bytes += len(raw)
                if used_bytes > 20*1024*1024:
                    raise CollectionError("DATA_QUALITY", "MBS job byte budget exceeded")
                documents.append({"url": pdf, "sha256": hashlib.sha256(raw).hexdigest(),
                                  "pdfBase64": base64.b64encode(raw).decode("ascii")})
                candidates += extract_text(text_extractor(raw), report_year, report_month, today, pdf)
            if not candidates:
                warnings.append({"url": target, "kind": "NO_EXPLICIT_PMI"}); continue
            for row in candidates:
                if first and not first <= row["period"] <= last: continue
                key = row["period"]
                if key in seen and seen[key] != row["value"]:
                    raise CollectionError("CONFLICT", "Conflicting MBS reports for one PMI month")
                if key not in seen:
                    binding = config[0]
                    rows.append({**row, "binding_id": binding["id"], "indicator_code": "VN_PMI", "unit": "index",
                                 "series_id": binding["series_id"], "priority": binding["priority"],
                                 "method": "PUBLISHED_MANUFACTURING_PMI; SECONDARY=MBS"})
                    seen[key] = row["value"]
        except CollectionError as error:
            if error.kind == "CONFLICT":
                error.payload = json.dumps({"documents": documents, "warnings": warnings}, ensure_ascii=False).encode()
                raise
            warnings.append({"url": target, "kind": error.kind})
    if len(links)>8: warnings.append({"kind": "PAGE_TRUNCATED", "discovered": len(links), "processed": 8})
    raw = json.dumps({"documents": documents, "warnings": warnings}, ensure_ascii=False).encode()
    return raw, rows, {"warnings": warnings, "page": page, "discovered": len(links), "history": True}
