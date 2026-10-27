"""Published NHNN monthly VND deposit-rate range, 6–12 months."""
import base64
import json
import os
import re
from urllib.parse import unquote, urljoin

from .cpi_reports import ArticleHTML, fold
from .http import CollectionError
from .mbs_history import pdf_text
from .sbv_history import allowed, transport
from .sdmx import month_end

INDEX = 'https://sbv.gov.vn/vi/trang-chu'
TITLE = r'dien bien lai suat cua to chuc tin dung doi voi khach hang\s+thang\s+(\d{1,2})\s*[/.-]\s*(20\d{2})'
VERSION = 'sbv_deposit_v1'


def parse_report(content, source_url, binding, today):
    content = fold(content)
    matches = set(re.findall(TITLE, content[:1500]))
    if len(matches) != 1:
        raise CollectionError('DATA_QUALITY', 'Missing or ambiguous deposit report month')
    month, year = next(iter(matches))
    period = f'{int(year):04d}-{int(month):02d}'
    end = month_end(period)
    if end > today:
        raise CollectionError('DATA_QUALITY', 'Future deposit report period')
    start = re.search(r'1\s*[.)]?\s*lai suat tien gui\s*:', content)
    if not start:
        raise CollectionError('DATA_QUALITY', 'Missing deposit section')
    section = content[start.end():]
    section = re.split(r'lai suat usd|2\s*[.)]\s*lai suat cho vay', section, maxsplit=1)[0]
    required = ['lai suat vnd', 'tien gui bang vnd', 'nhtm nha nuoc', 'nhtm co phan', 'moi va cu']
    if not all(word in section for word in required):
        raise CollectionError('DATA_QUALITY', 'Deposit population/currency changed')
    ranges = set(re.findall(r'(\d+(?:,\d+)?)\s*[-–−]\s*(\d+(?:,\d+)?)\s*%\s*/\s*nam\s+doi voi tien gui co ky han tu\s+6\s+thang den\s+12\s+thang', section))
    if len(ranges) != 1:
        raise CollectionError('DATA_QUALITY', 'Missing or conflicting 6–12 month deposit range')
    lower, upper = (float(v.replace(',', '.')) for v in next(iter(ranges)))
    if not 0 <= lower <= upper <= 100:
        raise CollectionError('DATA_QUALITY', 'Invalid deposit range bounds')
    return dict(binding_id=binding['id'], indicator_code=binding['indicator_code'],
                series_id=binding['series_id'], unit=binding['unit'], priority=binding['priority'],
                period=period, period_end=end, value=lower, upper_value=upper, source_url=source_url,
                published_at=None, method='PUBLISHED_RANGE; VND; TENOR=6_12_MONTHS; NEW_AND_EXISTING_DEPOSITS; NO_MIDPOINT')


def collect(client, dataset, config, today, page=1, first=None, last=None, text_extractor=pdf_text):
    if dataset['url'] != INDEX or len(config) != 1 or config[0]['indicator_code'] != 'VN_DEPOSIT_RATE' or config[0]['extractor_key'] != VERSION or config[0]['frequency'] != 'MONTHLY' or config[0]['unit'] != '%' or config[0]['comparison_basis'] != 'RANGE':
        raise CollectionError('CONFIGURATION', 'Unexpected deposit dataset definition')
    documents = []; targets = []; warnings = []
    def capture(url):
        raw = client.get(transport(url))
        documents.append({'url': url, 'bodyBase64': base64.b64encode(raw).decode('ascii')})
        return raw
    explicit = os.getenv('SBV_DEPOSIT_REPORT_URLS', '').split()
    if any(not allowed(url) for url in explicit):
        raise CollectionError('CONFIGURATION', 'Deposit report overrides must be public NHNN URLs')
    try:
        raw = capture(INDEX)
        parser = ArticleHTML(); parser.feed(raw.decode('utf-8'))
        for href, label in parser.links:
            target = urljoin(INDEX, href)
            if re.search(TITLE, fold(label)) and allowed(target) and target not in targets:
                targets.append(target)
    except CollectionError as error:
        if not explicit:
            raise
        warnings.append({'kind': error.kind, 'message': 'Discovery unavailable; explicit reports only'})
    targets = list(dict.fromkeys(explicit + targets))
    if len(explicit) > 20:
        raise CollectionError('CONFIGURATION', 'At most 20 explicit deposit report URLs')
    targets = targets[(page-1)*4:page*4]
    if not targets:
        raise CollectionError('NO_DATA', 'No NHNN monthly deposit reports discovered')
    rows = {}
    for target in targets:
        try:
            raw = capture(target)
            source = target
            if raw.startswith(b'%PDF-'):
                content = text_extractor(raw)
            else:
                parser = ArticleHTML(); parser.feed(raw.decode('utf-8'))
                content = ' '.join(parser.h1 + parser.text)
                pdfs = list(dict.fromkeys(urljoin(target, href) for href, label in parser.links
                            if '.pdf' in unquote(href).lower() and allowed(urljoin(target, href))
                            and 'dien bien lai suat' in fold(label)))
                if pdfs:
                    if len(pdfs) != 1:
                        raise CollectionError('DATA_QUALITY', 'Ambiguous deposit PDF attachment')
                    source = pdfs[0]
                    content = text_extractor(capture(source))
            row = parse_report(content, source, config[0], today)
            if first and not first <= row['period'] <= last:
                continue
            old = rows.get(row['period'])
            if old and (old['value'], old['upper_value']) != (row['value'], row['upper_value']):
                raise CollectionError('CONFLICT', 'Conflicting same-month deposit reports')
            rows[row['period']] = row
        except (CollectionError, ValueError, UnicodeError) as error:
            if getattr(error, 'kind', '') == 'CONFLICT':
                raise
            warnings.append({'url': target, 'kind': getattr(error, 'kind', 'DATA_QUALITY')})
    return json.dumps({'documents': documents}, ensure_ascii=False).encode(), list(rows.values()), {'warnings': warnings, 'publishedRange': True}
