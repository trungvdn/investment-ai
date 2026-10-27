"""Published NHNN monthly VND deposit-rate range, 6–12 months."""
import base64
import json
import os
import re
import time
from urllib.parse import unquote, urljoin, urlsplit, parse_qs, urlencode
from pathlib import Path
from .selection import periods_between

from .cpi_reports import ArticleHTML, fold
from .http import CollectionError, HTTPClient
from .mbs_history import pdf_text
from .sbv_history import allowed, transport, Tree, nodes, text
from .sdmx import month_end

INDEX = 'https://sbv.gov.vn/vi/trang-chu'
TITLE = r'dien bien lai suat cua to chuc tin dung doi voi khach hang\s+thang\s+(\d{1,2})\s*[/.-]\s*(20\d{2})'
VERSION = 'sbv_deposit_v1'
ARCHIVE = 'https://sbv.gov.vn/vi/thong-cao-bao-chi'


def load_registry():
    return json.loads(Path(__file__).with_name('sbv_deposit_reports.json').read_text())


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
    last_request = 0.0
    def capture(url):
        nonlocal last_request
        if isinstance(client, HTTPClient):
            time.sleep(max(0, last_request + 2 - time.monotonic()))
        last_request = time.monotonic()
        raw = client.get(transport(url))
        documents.append({'url': url, 'bodyBase64': base64.b64encode(raw).decode('ascii')})
        return raw
    explicit = os.getenv('SBV_DEPOSIT_REPORT_URLS', '').split()
    if any(not allowed(url) for url in explicit):
        raise CollectionError('CONFIGURATION', 'Deposit report overrides must be public NHNN URLs')
    expected = periods_between(first, last)[(page-1)*4:page*4] if first else None
    if first and not expected:
        raise CollectionError('CONFIGURATION', 'Page exceeds requested month interval')
    wanted = set(expected or [])
    registry = load_registry()
    registered = {entry['url']: entry['period'] for entry in registry if first and entry['period'] in wanted}
    discovered = dict(registered)
    queue = [] if wanted and wanted <= set(registered.values()) else [INDEX, ARCHIVE]
    visited = set()
    publishers = set()
    source_paths = set()
    limit = 10 if first else 2
    while queue and len(visited) < limit:
        index = queue.pop(0)
        if index in visited:
            continue
        visited.add(index)
        try:
            raw = capture(index)
            parser = ArticleHTML(); parser.feed(raw.decode('utf-8'))
            for href, label in parser.links:
                target = urljoin(index, href)
                if not allowed(target):
                    continue
                match = re.search(TITLE, fold(label))
                if match:
                    period = f'{int(match[2]):04d}-{int(match[1]):02d}'
                    if not first or period in wanted:
                        key = (urlsplit(target).hostname.removeprefix('www.'), urlsplit(target).path)
                        if key not in source_paths:
                            discovered.setdefault(target, period)
                            source_paths.add(key)
                # Follow actual publisher pagination, never a guessed article slug.
                parsed = urlsplit(target)
                if first and parsed.path == urlsplit(ARCHIVE).path:
                    query = parse_qs(parsed.query)
                    cursors = [v[0] for k,v in query.items() if k.endswith('_cur')]
                    if cursors and all(v.isdigit() and 1 <= int(v) <= 20 for v in cursors) and target not in visited and target not in queue:
                        queue.append(target)
            if first and urlsplit(index).path == urlsplit(ARCHIVE).path:
                tree = Tree(); tree.feed(raw.decode('utf-8'))
                for node in nodes(tree.root, 'a'):
                    if not re.search(TITLE, fold(text(node))):
                        continue
                    parent = node
                    while parent:
                        match = re.search(r'AssetPublisherPortlet_INSTANCE_[A-Za-z0-9]+', parent.attrs.get('id',''))
                        if match:
                            identity = 'com_liferay_asset_publisher_web_portlet_' + match[0]
                            if identity not in publishers:
                                publishers.add(identity)
                                # The normal page hides the older monthly releases
                                # in a separate publisher without visible page links.
                                query = {'p_p_id':identity,'p_p_lifecycle':0,'p_p_state':'normal','p_p_mode':'view',
                                         '_'+identity+'_cur':1,'_'+identity+'_delta':100}
                                queue.insert(0, ARCHIVE+'?'+urlencode(query))
                            break
                        parent = parent.parent
            if wanted and wanted <= set(discovered.values()):
                break
        except (CollectionError, UnicodeError) as error:
            warnings.append({'kind': getattr(error,'kind','DATA_QUALITY'), 'url': index, 'message': 'Deposit index unavailable'})
    if len(explicit) > 20:
        raise CollectionError('CONFIGURATION', 'At most 20 explicit deposit report URLs')
    targets = list(dict.fromkeys(explicit + list(discovered)))
    if not first:
        targets = targets[(page-1)*4:page*4]
    if len(targets) > 24:
        raise CollectionError('CONFIGURATION', 'Too many candidate reports in bounded month batch')
    if not targets and not first:
        raise CollectionError('NO_DATA', 'No NHNN deposit reports found for requested months')
    rows = {}
    for target in targets:
        if target in registered and registered[target] in rows:
            continue  # Alternate address for a report already downloaded successfully.
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
            if target in discovered and row['period'] != discovered[target]:
                raise CollectionError('DATA_QUALITY', 'PDF reference month differs from archive link')
            if wanted and row['period'] not in wanted:
                continue
            old = rows.get(row['period'])
            if old and (old['value'], old['upper_value']) != (row['value'], row['upper_value']):
                raise CollectionError('CONFLICT', 'Conflicting same-month deposit reports')
            rows[row['period']] = row
        except (CollectionError, ValueError, UnicodeError) as error:
            if getattr(error, 'kind', '') == 'CONFLICT':
                raise
            warnings.append({'url': target, 'kind': getattr(error, 'kind', 'DATA_QUALITY')})
    missing = sorted(wanted - set(rows))
    if missing:
        warnings.append({'kind': 'MISSING_PERIODS', 'months': missing})
    return json.dumps({'documents': documents}, ensure_ascii=False).encode(), list(rows.values()), {'warnings': warnings, 'publishedRange': True, 'requestedMonths': expected, 'missingMonths': missing}

