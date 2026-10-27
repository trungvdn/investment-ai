"""Strict NHNN credit YTD archive and explicitly dated O/N tables."""
import json
import math
import os
import re
from datetime import date, datetime, timezone
from html.parser import HTMLParser
from urllib.parse import quote, urlencode, urljoin, urlsplit, urlunsplit

from .cpi_reports import ArticleHTML, fold
from .http import CollectionError
from .sdmx import month_end

ARTICLES="https://sbv.gov.vn/o/article/v1.0/articles"
CREDIT="https://sbv.gov.vn/vi/du-no-tin-dung-doi-voi-nen-kt-dttktt"
INTERBANK="https://sbv.gov.vn/vi/lãi-suất1"


def number(value):
    if isinstance(value,(int,float)) and not isinstance(value,bool):
        result=float(value)
    else:
        value=str(value).strip()
        if not re.fullmatch(r"[+-]?(?:\d{1,3}(?:\.\d{3})+|\d+)(?:,\d+)?",value):
            raise ValueError("Invalid Vietnamese number")
        result=float(value.replace(".","").replace(",","."))
    if not math.isfinite(result):raise ValueError("Nonfinite value")
    return result


def field(value):
    if isinstance(value,dict):
        for key in ("content","value","vi_VN"):
            if key in value:return field(value[key])
        return ""
    return value if value is not None else ""


def stamp(value):
    if not value:return None
    result=datetime.fromisoformat(value.replace("Z","+00:00"))
    if result.tzinfo is None:raise ValueError("Source timestamp needs timezone")
    return result


def parse_credit(payload,config,today):
    if len(config)!=1 or config[0]["extractor_key"]!="sbv_credit_history_v1" or config[0]["indicator_code"]!="VN_CREDIT_GROWTH" or config[0]["unit"]!="% YTD":
        raise CollectionError("DATA_QUALITY","Unsupported credit definition")
    selected={};warnings=[]
    for article in payload["articles"]:
        if article.get("status")!=0 or article.get("ddmStructureId")!=10034332:continue
        try:
            title=article["title"]
            if re.search(r"\btỉnh\b|\bthành phố\b",title,re.I) or any(name in fold(title) for name in ("ha noi","ho chi minh","tphcm")):
                raise ValueError("Regional credit is not national credit")
            match=re.search(r"thang\s+(\d{1,2})\s+nam\s+(20\d{2})",fold(title))
            if not match:raise ValueError("Missing monthly title")
            period=f"{int(match[2]):04d}-{int(match[1]):02d}"
            fields=article["fields"]
            end=date.fromisoformat(str(field(fields["Date62813077"])))
            if end!=month_end(period) or end>today:raise ValueError("Not a past explicit month end")
            level=number(field(fields["soDuTongCong"]))
            growth=number(field(fields["tocDoTangGiamTongCong"]))
            if level<=0 or growth < -100:raise ValueError("Invalid national total")
            published=stamp(article.get("datePublished"))
            updated=stamp(article.get("dateModified")) or published or datetime(1970,1,1,tzinfo=timezone.utc)
            version=float(article.get("version",0))
            if not math.isfinite(version):raise ValueError("Invalid revision")
            row={"period":period,"period_end":end,"value":growth,"source_url":CREDIT,"published_at":published,
                 "evidence":{"articleId":str(article.get("articleId","")),"nationalCreditLevel":level,"basis":"YTD","version":version}}
            old=selected.get(period)
            if old and (updated,version)==old[:2] and (growth,level)!=(old[2]["value"],old[2]["evidence"]["nationalCreditLevel"]):
                raise CollectionError("CONFLICT","Conflicting same-vintage credit totals")
            if not old or (updated,version)>old[:2]:selected[period]=(updated,version,row)
        except (KeyError,ValueError,TypeError,OverflowError):
            warnings.append({"article":str(article.get("articleId","")),"kind":"INVALID_MONTH_END_TOTAL"})
    return [item[2] for _,item in sorted(selected.items())],warnings


class Node:
    def __init__(self,tag="",attrs=None,parent=None):
        self.tag,self.attrs,self.parent,self.children=tag,dict(attrs or []),parent,[]


class Tree(HTMLParser):
    def __init__(self):
        super().__init__();self.root=Node();self.current=self.root
    def handle_starttag(self,tag,attrs):
        child=Node(tag,attrs,self.current);self.current.children.append(child)
        if tag not in ("br","img","input","meta","link","hr","source","wbr"):self.current=child
    def handle_endtag(self,tag):
        node=self.current
        while node.parent:
            if node.tag==tag:self.current=node.parent;return
            node=node.parent
    def handle_data(self,text):self.current.children.append(text)


def nodes(node,tag):
    result=[]
    for child in node.children:
        if isinstance(child,Node):
            if child.tag==tag:result.append(child)
            result+=nodes(child,tag)
    return result


def text(node):
    if isinstance(node,str):return node
    if node.tag in ("script","style"):return ""
    return " ".join(text(child) for child in node.children)


def _parse_interbank(body,today):
    tree=Tree();tree.feed(body.decode("utf-8"))
    result={}
    for table in nodes(tree.root,"table"):
        table_text=fold(text(table))
        if not re.search(r"lai suat (?:bq|binh quan) lien ngan hang",table_text) or "usd" in table_text:continue
        parent=table.parent
        while parent and "bi01-wrap" not in parent.attrs.get("class","").split():parent=parent.parent
        if not parent:continue
        context=fold(text(parent))
        effective=re.findall(r"ngay ap dung:\s*(\d{2}/\d{2}/\d{4})",context)
        if len(set(effective))!=1:raise CollectionError("DATA_QUALITY","Ambiguous O/N table date")
        end=datetime.strptime(effective[0],"%d/%m/%Y").date()
        candidates=[]
        for tr in nodes(table,"tr"):
            cells=[text(c).strip() for c in tr.children if isinstance(c,Node) and c.tag in ("td","th")]
            if len(cells)>=2 and fold(cells[0])=="qua dem":candidates.append(cells)
        if len(candidates)!=1:raise CollectionError("DATA_QUALITY","Ambiguous O/N row")
        cells=candidates[0]
        referenced=any("*" in cell for cell in cells[1:])
        if referenced:
            dates=re.findall(r"tham chieu[^;]{0,100}?ngay\s+(\d{2}/\d{2}/\d{4})",context)
            if len(set(dates))!=1:raise CollectionError("DATA_QUALITY","Missing starred O/N reference date")
            end=datetime.strptime(dates[0],"%d/%m/%Y").date()
        if end>today:raise CollectionError("DATA_QUALITY","Future O/N date")
        try:value=number(cells[1].replace("(*)","").replace("*","").strip())
        except ValueError:raise CollectionError("DATA_QUALITY","Invalid O/N rate") from None
        if value<0 or value>100:raise CollectionError("DATA_QUALITY","Invalid O/N rate bounds")
        period=end.isoformat()
        if period in result and result[period]["value"]!=value:raise CollectionError("CONFLICT","Conflicting O/N tables")
        result[period]={"period":period,"period_end":end,"value":value,"source_url":INTERBANK,"published_at":None,
                        "evidence":{"effectiveDate":effective[0],"referencedPreviousDate":referenced,"tenor":"O/N","currency":"VND"}}
    return list(result.values())



def parse_interbank(body,today):
    try:return _parse_interbank(body,today)
    except CollectionError as error:
        error.payload=body;raise
    except (ValueError,UnicodeError):
        error=CollectionError("DATA_QUALITY","Invalid O/N HTML date/encoding");error.payload=body;raise error from None


def parse_credit_html(body,config,today):
    tree=Tree();tree.feed(body.decode("utf-8"))
    output={}
    for table in nodes(tree.root,"table"):
        cells_by_row=[[text(c).strip() for c in tr.children if isinstance(c,Node) and c.tag in ("td","th")] for tr in nodes(table,"tr")]
        totals=[(cells,i) for cells in cells_by_row for i,cell in enumerate(cells) if fold(cell)=="tong cong"]
        if not totals:continue
        parent=table
        period=None;context=""
        for _ in range(8):
            context=fold(text(parent))
            matches=set(re.findall(r"thang\s+(\d{1,2})\s+nam\s+(20\d{2})",context))
            if len(matches)==1:
                month,year=next(iter(matches));period=f"{int(year):04d}-{int(month):02d}";break
            if not parent.parent:break
            parent=parent.parent
        if not period or "du no tin dung doi voi nen kinh te" not in context:continue
        if not ("so voi cuoi nam truoc" in context or "so voi cuoi nam "+str(int(period[:4])-1) in context):continue
        if any(name in context for name in ("dia ban ha noi","dia ban tphcm","dia ban thanh pho")):continue
        if len(totals)!=1:raise CollectionError("CONFLICT","Ambiguous national credit total")
        cells,i=totals[0]
        try:
            level,growth=number(cells[i+1]),number(cells[i+2])
            end=month_end(period)
            if end>today or level<=0 or growth < -100:raise ValueError()
        except (ValueError,IndexError):continue
        if period in output and output[period]["value"]!=growth:raise CollectionError("CONFLICT","Conflicting national credit tables")
        output[period]={"period":period,"period_end":end,"value":growth,"source_url":CREDIT,"published_at":None,
                        "evidence":{"basis":"YTD","nationalCreditLevel":level,"extraction":"explicit total table"}}
    return list(output.values())

def transport(url):
    p=urlsplit(url)
    return urlunsplit((p.scheme,p.netloc,quote(p.path,safe="/%"),p.query,p.fragment))


def allowed(url):
    p=urlsplit(url)
    return p.scheme=="https" and p.netloc in ("sbv.gov.vn","www.sbv.gov.vn") and not p.fragment


def collect(client,dataset,config,today,page=1,first=None,last=None):
    documents=[];notes={"history":True,"warnings":[],"page":page}
    if dataset["id"]=="sbv_credit_history":
        if dataset["url"]!=ARTICLES or not 1<=page<=100:raise CollectionError("DATA_QUALITY","Unexpected credit archive URL/page")
        url=ARTICLES+"?"+urlencode({"scopeKey":"20117","contentStructureId":"10034332","pageSize":200,"page":page,"filter":"status eq 0"})
        try:
            raw=client.get(url)
        except CollectionError as denied:
            if denied.kind!="ACCESS_DENIED":raise
            raw=client.get(CREDIT);documents.append({"url":CREDIT,"html":raw.decode("utf-8")})
            try:rows=parse_credit_html(raw,config,today)
            except CollectionError as error:error.payload=raw;raise
            notes["warnings"].append({"kind":"API_ACCESS_DENIED","fallback":"public national credit HTML; archive may be incomplete"})
            binding=config[0]
            result=[{**r,"binding_id":binding["id"],"indicator_code":binding["indicator_code"],"unit":binding["unit"],"series_id":binding["series_id"],"priority":binding["priority"],"method":"SOURCE_TOTAL_CREDIT_YTD_NOT_YOY"} for r in rows if not first or first<=r["period"]<=last]
            return json.dumps({"documents":documents,"notes":notes},ensure_ascii=False).encode(),result,notes
        documents.append({"url":url,"json":raw.decode("utf-8")})
        try:
            data=json.loads(raw)
            if data["page"]!=page or not page<=data["lastPage"]<=100 or not 0<len(data["articles"])<=200:
                raise ValueError("Invalid archive pagination")
            rows,warnings=parse_credit(data,config,today)
            notes.update(lastPage=data["lastPage"],totalCount=data.get("totalCount"),warnings=warnings)
        except CollectionError as error:
            error.payload=raw;raise
        except (ValueError,KeyError,TypeError):
            error=CollectionError("DATA_QUALITY","Invalid SBV credit archive response");error.payload=raw;raise error from None
        method="SOURCE_TOTAL_CREDIT_YTD_NOT_YOY"
    else:
        if dataset["url"]!=INTERBANK or len(config)!=1 or config[0]["extractor_key"]!="sbv_interbank_history_v1" or config[0]["indicator_code"]!="VN_INTERBANK_ON" or config[0]["unit"]!="%":
            raise CollectionError("DATA_QUALITY","Unexpected O/N dataset definition")
        # No invented date-query parameter: only current table, actual dated links,
        # or explicitly configured public NHNN historical documents.
        raw=client.get(transport(INTERBANK));documents.append({"url":INTERBANK,"html":raw.decode("utf-8")})
        rows=parse_interbank(raw,today)
        parser=ArticleHTML();parser.feed(raw.decode("utf-8"))
        links=[]
        for href,label in parser.links:
            target=urljoin(INTERBANK,href)
            if allowed(target) and "lai suat" in fold(label) and re.search(r"20\d{2}",label) and target!=INTERBANK and target not in links:links.append(target)
        for target in os.getenv("SBV_INTERBANK_HISTORY_URLS","").split():
            if not allowed(target):raise CollectionError("CONFIGURATION","Historical O/N URLs must be public NHNN pages")
            if target not in links:links.append(target)
        if page!=1:links=links[(page-1)*8:page*8]
        else:links=links[:8]
        for target in links:
            try:
                raw=client.get(transport(target));documents.append({"url":target,"html":raw.decode("utf-8")})
                for row in parse_interbank(raw,today):
                    row["source_url"]=target;rows.append(row)
            except CollectionError as error:
                notes["warnings"].append({"url":target,"kind":error.kind})
        notes["archiveStatus"]="PARTIAL_VERIFIED_TABLES_ONLY"
        notes["dailyCoverage"]="Unobserved weekdays may include holidays; no zero fill or forward fill"
        method="SOURCE_VND_OVERNIGHT_WEIGHTED_AVERAGE_RATE; PERIOD=OBSERVATION_DATE"
    binding=config[0];unique={}
    for row in rows:
        if first and not first<=row["period_end"].strftime("%Y-%m")<=last:continue
        if row["period"] in unique and unique[row["period"]]["value"]!=row["value"]:
            raise CollectionError("CONFLICT","Conflicting SBV historical values")
        unique[row["period"]]={**row,"binding_id":binding["id"],"indicator_code":binding["indicator_code"],
                               "unit":binding["unit"],"series_id":binding["series_id"],"priority":binding["priority"],"method":method}
    return json.dumps({"documents":documents,"notes":notes},ensure_ascii=False).encode(),list(unique.values()),notes
