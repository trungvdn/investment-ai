"""National GSO/NSO CPI article scraping; comparison bases stay separate."""
import calendar
import json
import math
import re
import unicodedata
from datetime import date, datetime, timezone
from html.parser import HTMLParser
from urllib.parse import urljoin, urlsplit

from .core import ZONE
from .http import CollectionError, NotPublished

INDEX = "https://www.gso.gov.vn/cpi-vi/"
ALTERNATE = "https://www.nso.gov.vn/cpi-vi/"
HOSTS = {"www.gso.gov.vn", "gso.gov.vn", "www.nso.gov.vn", "nso.gov.vn"}
VERSION = "gso_cpi_article_v1"
MONTHS = {"mot":1,"gieng":1,"hai":2,"ba":3,"tu":4,"nam":5,"sau":6,"bay":7,"tam":8,"chin":9,"muoi":10,"muoi mot":11,"muoi hai":12}
MONTH = r"(?:muoi\s+hai|muoi\s+mot|gieng|mot|hai|ba|tu|nam|sau|bay|tam|chin|muoi|\d{1,2})"
NUM = r"\d+(?:[,.]\d+)?"
CHANGE = r"(tang|giam|khong doi|khong thay doi)\s*("+NUM+r")?\s*%?"


def fold(text):
    text = unicodedata.normalize("NFD", text.lower().replace("đ", "d"))
    return " ".join("".join(c for c in text if not unicodedata.combining(c)).split())


def official(url):
    parsed = urlsplit(url)
    return parsed.scheme=="https" and parsed.netloc in HOSTS and not parsed.fragment


class ArticleHTML(HTMLParser):
    def __init__(self):
        super().__init__()
        self.links=[];self.h1=[];self.paragraphs=[];self.text=[];self.dates=[]
        self._a=None;self._h=None;self._p=None;self.skip=0

    def handle_starttag(self,tag,attrs):
        attrs=dict(attrs)
        if tag in ("script","style"):
            self.skip+=1
        if tag=="meta" and attrs.get("property")=="article:published_time":
            self.dates.append(attrs.get("content",""))
        if self.skip:return
        if tag=="a":self._a=[attrs.get("href",""),attrs.get("title","")]
        if tag=="h1":self._h=[]
        if tag=="p":self._p=[]

    def handle_data(self,text):
        if self.skip:return
        self.text.append(text)
        if self._a is not None:self._a[1]+=" "+text
        if self._h is not None:self._h.append(text)
        if self._p is not None:self._p.append(text)

    def handle_endtag(self,tag):
        if tag in ("script","style"):
            self.skip=max(0,self.skip-1);return
        if self.skip:return
        if tag=="a" and self._a is not None:self.links.append(tuple(self._a));self._a=None
        if tag=="h1" and self._h is not None:self.h1.append(" ".join(self._h));self._h=None
        if tag=="p" and self._p is not None:self.paragraphs.append(" ".join(self._p));self._p=None


def month_number(word):
    return int(word) if word.isdigit() else MONTHS[word]


def report_period(title):
    text=fold(title)
    if re.search(r"\btỉnh\b|\bthành phố\b",title,re.I) or re.search(r"\b(?:thanh pho|khu vuc|ha noi|ho chi minh)\b",text):
        raise ValueError("Regional CPI is not the national series")
    if not ("chi so gia tieu dung" in text or "thong cao bao chi ve tinh hinh gia" in text):
        raise ValueError("Not a national CPI/price report")
    match=re.search(r"thang\s+("+MONTH+r")\b",text)
    years=set(re.findall(r"\b20\d{2}\b",text))
    if not match or len(years)!=1:raise ValueError("Ambiguous CPI report month/year")
    month=month_number(match[1]);year=int(years.pop())
    if not 1<=month<=12:raise ValueError("Invalid CPI report month")
    return f"{year:04d}-{month:02d}"


def change_value(direction,value):
    if direction in ("khong doi","khong thay doi"):
        if value is not None and float(value.replace(",","."))!=0:raise ValueError("Contradictory unchanged CPI")
        return 0.0
    if value is None:raise ValueError("Missing published CPI percentage")
    number=float(value.replace(",","."))
    return -number if direction=="giam" else number


def signed(match):
    return change_value(match.group(1),match.group(2))


def extract_article(html,url,today):
    if not official(url):raise CollectionError("DATA_QUALITY","CPI report must be from national GSO/NSO")
    parser=ArticleHTML();parser.feed(html)
    try:
        if len(parser.h1)!=1:raise ValueError("Missing/ambiguous article heading")
        period=report_period(parser.h1[0]);year,month=map(int,period.split("-"))
        end=date(year,month,calendar.monthrange(year,month)[1])
        if end>today:raise ValueError("Future CPI report period")
        # Retain source publication semantics; missing dates remain null.
        published=None
        if parser.dates:
            dates={datetime.fromisoformat(d.replace("Z","+00:00")) for d in parser.dates}
            if len(dates)!=1:raise ValueError("Ambiguous article publication time")
            published=dates.pop()
            if published.tzinfo is None:raise ValueError("Publication time requires timezone")
        else:
            dates=set(re.findall(r"ngay dang:\s*(\d{2}/\d{2}/\d{4})",fold(" ".join(parser.text))))
            if len(dates)>1:raise ValueError("Ambiguous publication date")
            if dates:published=datetime.strptime(dates.pop(),"%d/%m/%Y").replace(tzinfo=ZONE)
        if published and (published.date()<end or published.date()>today):raise ValueError("Inconsistent CPI publication date")
        values={"MoM":[],"YoY":[],"AVG_YOY":[]}
        evidence={"MoM":[],"YoY":[],"AVG_YOY":[]}
        for paragraph_index,paragraph in enumerate(parser.paragraphs):
            text=fold(paragraph)
            if re.match(r"(?:chi so gia tieu dung\s*(?:\(cpi\))?|cpi)\s+thang\b",text):
                monthly=re.split(r"cpi binh quan|binh quan|lam phat co ban|chi so gia vang|chi so gia do la",text,maxsplit=1)[0]
                for basis,pattern in (("MoM",CHANGE+r"\s+so voi thang truoc"),("YoY",CHANGE+r"\s+so voi cung ky (?:nam truoc|nam\s+"+str(year-1)+r")")):
                    for match in re.finditer(pattern,monthly):
                        values[basis].append(signed(match));evidence[basis].append({"text":match.group(0),"paragraphIndex":paragraph_index})
            subject=r"(?:chi so gia tieu dung\s*(?:\(cpi\))?|cpi)\s+thang\s+"+MONTH+r"(?:/\d{4}|\s+nam\s+\d{4})?\s+"
            for basis,comparator in (("MoM","thang truoc"),("YoY",r"cung ky (?:nam truoc|nam\s+"+str(year-1)+r")")):
                inverse=re.search(r"^so voi "+comparator+r"\s*[,;:]?\s*"+subject+CHANGE,text)
                if inverse:
                    values[basis].append(signed(inverse));evidence[basis].append({"text":inverse.group(0),"paragraphIndex":paragraph_index})
            # Full YTD average only; quarter averages and component-price changes cannot match.
            patterns=[r"binh quan\s+("+MONTH+r")\s+thang\s+(?:dau\s+)?nam\s+"+str(year)+r"\s*[,;:]?\s*cpi\s+"+CHANGE,
                      r"cpi\s+binh quan\s+("+MONTH+r")\s+thang\s+(?:dau\s+)?nam\s+"+str(year)+r"\s*[,;:]?\s*"+CHANGE]
            for pattern in patterns:
                for match in re.finditer(pattern,text):
                    if month_number(match[1])!=month:continue
                    direction,value=match[2],match[3]
                    percent=change_value(direction,value)
                    values["AVG_YOY"].append(percent);evidence["AVG_YOY"].append({"text":match.group(0),"paragraphIndex":paragraph_index})
        if month==12:
            for paragraph_index,paragraph in enumerate(parser.paragraphs):
                text=fold(paragraph)
                for pattern in (r"binh quan nam "+str(year)+r"\s*[,;:]?\s*cpi\s+"+CHANGE,r"cpi binh quan nam "+str(year)+r"\s*[,;:]?\s*"+CHANGE):
                    for match in re.finditer(pattern,text):
                        values["AVG_YOY"].append(signed(match));evidence["AVG_YOY"].append({"text":match.group(0),"paragraphIndex":paragraph_index})
        result=[]
        for basis,found in values.items():
            unique=set(found)
            if not unique and basis=="AVG_YOY":continue
            if len(unique)!=1:raise ValueError("Missing/conflicting CPI comparison basis")
            value=unique.pop()
            if not math.isfinite(value) or value < -100:raise ValueError("CPI percentage below definition bounds")
            result.append({"period":period,"period_end":end,"value":value,"basis":basis,"published_at":published,"source_url":url,"evidence":{**evidence[basis][0],"comparisonBasis":basis,"windowMonths":month if basis=="AVG_YOY" else 1,"publicationPrecision":"timestamp" if parser.dates else "date" if published else "unknown"}})
        return result
    except (ValueError,KeyError,TypeError,OverflowError):
        raise CollectionError("DATA_QUALITY","CPI article lacks unambiguous national monthly/YoY/YTD-average figures") from None


def collect(client,index,today):
    if not official(index) or urlsplit(index).path!="/cpi-vi/":raise CollectionError("DATA_QUALITY","CPI collector index must be national GSO/NSO CPI topic")
    try:page=client.get(index)
    except CollectionError:
        if urlsplit(index).netloc not in {"www.gso.gov.vn","gso.gov.vn"}:raise
        index=ALTERNATE;page=client.get(index)
    parser=ArticleHTML();parser.feed(page.decode("utf-8"))
    urls=[]
    periods={}
    for href,title in parser.links:
        try:period=report_period(title)
        except ValueError:continue
        url=urljoin(index,href)
        if official(url) and url not in urls:
            urls.append(url);periods[url]=period
    if not urls:
        error=CollectionError("DATA_QUALITY","No dated CPI article links on GSO/NSO topic page")
        error.payload=page
        raise error
    newest=max(periods.values())
    urls=[url for url in urls if periods[url]==newest]
    articles=[];rows=[];seen={};warnings=[]
    def archive():
        return json.dumps({"indexUrl":index,"indexHtml":page.decode("utf-8"),"articles":articles,"warnings":warnings},ensure_ascii=False).encode()
    def fail(message):
        error=CollectionError("DATA_QUALITY",message)
        error.payload=archive()
        raise error
    # Category cards may contain both summaries and detailed reports for a month.
    for url in urls[:12]:
        try:html=client.get(url).decode("utf-8")
        except CollectionError as error:
            warnings.append({"url":url,"kind":error.kind});continue
        articles.append({"url":url,"html":html})
        try:parsed=extract_article(html,url,today)
        except CollectionError:
            warnings.append({"url":url,"kind":"DATA_QUALITY"});continue
        for row in parsed:
            key=(row["period"],row["basis"])
            if key in seen and seen[key]["value"]!=row["value"]:fail("Conflicting GSO/NSO CPI reports for the same month/basis")
            if key not in seen:seen[key]=row;rows.append(row)
    if not rows:fail("No CPI articles passed all comparison-basis checks")
    bundle=archive()
    return bundle,rows
