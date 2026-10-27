"""NHNN policy rate effective-date events; never daily/monthly forward-filled."""
import json
import re
from datetime import date, datetime

from .cpi_reports import fold
from .http import CollectionError
from .sbv_history import Tree, Node, nodes, text, number, transport, INTERBANK

RELEASE = 'https://xaydungchinhsach.chinhphu.vn/nhnn-tiep-tuc-dieu-chinh-lai-suat-dieu-hanh-119230616150218505.htm'
CODES = {
    'lai suat tai cap von': 'VN_SBV_REFINANCING',
    'lai suat tai chiet khau': 'VN_SBV_DISCOUNT',
    'lai suat cho vay qua dem': 'VN_SBV_OVERNIGHT',
}


def _day(value):
    return datetime.strptime(value, '%d/%m/%Y').date()


def parse_table(body, today):
    tree = Tree(); tree.feed(body.decode('utf-8'))
    result = {}
    for table in nodes(tree.root, 'table'):
        for tr in nodes(table, 'tr'):
            cells = [fold(text(c)).strip() for c in tr.children if isinstance(c, Node) and c.tag in ('td', 'th')]
            if len(cells) < 2:
                continue
            code = CODES.get(cells[0])
            if not code and cells[0].startswith('lai suat cho vay qua dem trong thanh toan dien tu lien ngan hang'):
                code = 'VN_SBV_OVERNIGHT'
            if not code:
                continue
            context = ' '.join(cells)
            dates = re.findall(r'\b\d{1,2}/\d{1,2}/\d{4}\b', context)
            # A date outside the row must be explicitly labelled as effective.
            if not dates:
                parent = table.parent
                for _ in range(4):
                    if not parent:
                        break
                    scoped = fold(text(parent))
                    dates = re.findall(r'ngay ap dung\s*:?\s*(\d{1,2}/\d{1,2}/\d{4})', scoped)
                    if dates:
                        break
                    parent = parent.parent
            if len(set(dates)) != 1:
                raise CollectionError('DATA_QUALITY', 'Missing or ambiguous policy effective date')
            end = _day(dates[0])
            value = number(re.sub(r'\s*%\s*(?:/\s*nam)?$', '', cells[1]).strip())
            if not 0 <= value <= 100 or end > today:
                raise CollectionError('DATA_QUALITY', 'Invalid policy rate/date')
            key = (code, end.isoformat())
            if key in result and result[key]['value'] != value:
                raise CollectionError('CONFLICT', 'Conflicting policy rates')
            result[key] = dict(indicator_code=code, period=end.isoformat(), period_end=end, value=value,
                               source_url=INTERBANK, published_at=None,
                               method='POLICY_RATE; PERIOD=EFFECTIVE_DATE; PARSER=CURRENT_NHNN_TABLE')
    if set(r['indicator_code'] for r in result.values()) != set(CODES.values()):
        raise CollectionError('DATA_QUALITY', 'Current table must identify all three policy rates')
    return list(result.values())


def parse_release(body, today):
    """Bounded parser for the verified 1123 release; cannot imply current validity."""
    tree = Tree(); tree.feed(body.decode('utf-8'))
    content = fold(text(tree.root))
    if not re.search(r'1123/qd-nhnn', content) or not re.search(r'hieu luc tu ngay 19\s*(?:thang\s+6\s+nam\s+2023|/\s*6/2023)', content):
        raise CollectionError('DATA_QUALITY', 'Unrecognized policy release or effective date')
    result = []
    end = date(2023, 6, 19)
    for label, code in CODES.items():
        pattern = label + r'[^;]{0,350}?giam tu muc\s+\d+(?:,\d+)?\s*%/nam\s+xuong\s+(\d+(?:,\d+)?)\s*%/nam'
        values = set(re.findall(pattern, content))
        if len(values) != 1 or end > today:
            raise CollectionError('DATA_QUALITY', 'Ambiguous policy release rates')
        value = number(values.pop())
        if not 0 <= value <= 100:
            raise CollectionError('DATA_QUALITY', 'Invalid policy release rate')
        result.append(dict(indicator_code=code, period=end.isoformat(), period_end=end, value=value,
                           source_url=RELEASE, published_at=None,
                           method='POLICY_RATE; PERIOD=EFFECTIVE_DATE; PARSER=GOVERNMENT_1123_CURRENT_STATUS_UNVERIFIED'))
    return result


def collect(client, dataset, config, today, page=1, first=None, last=None):
    if dataset['url'] != INTERBANK or len(config) != 3 or {b['indicator_code'] for b in config} != set(CODES.values()) or any(b['frequency'] != 'EVENT' or b['unit'] != '%' or b['extractor_key'] != 'sbv_policy_v1' for b in config):
        raise CollectionError('CONFIGURATION', 'Unexpected policy dataset definition')
    documents = []; warnings = []
    try:
        raw = client.get(transport(INTERBANK))
        documents.append({'url': INTERBANK, 'html': raw.decode('utf-8')})
        rows = parse_table(raw, today)
    except (CollectionError, ValueError, UnicodeError) as error:
        warnings.append({'kind': getattr(error, 'kind', 'DATA_QUALITY'),
                         'message': 'NHNN current table not verified; historical decision only'})
        raw = client.get(RELEASE)
        documents.append({'url': RELEASE, 'html': raw.decode('utf-8')})
        try:
            rows = parse_release(raw, today)
        except (ValueError, UnicodeError) as error:
            raise CollectionError('DATA_QUALITY', 'Invalid government policy release') from error
    by_code = {b['indicator_code']: b for b in config}
    output = []
    for row in rows:
        if first and not first <= row['period'][:7] <= last:
            continue
        b = by_code[row['indicator_code']]
        output.append({**row, **{key: b[key] for key in ('series_id', 'unit')},
                       'priority': b['priority'] - (50 if warnings else 0), 'binding_id': b['id']})
    return json.dumps({'documents': documents}, ensure_ascii=False).encode(), output, {'warnings': warnings, 'effectiveDateEvents': True}
