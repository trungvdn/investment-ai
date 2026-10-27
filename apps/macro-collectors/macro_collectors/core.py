import fcntl
import json
import logging
import os
import tempfile
from dataclasses import dataclass
from datetime import datetime, timedelta, timezone
from pathlib import Path
from zoneinfo import ZoneInfo

from .http import CollectionError, NotPublished

ZONE = ZoneInfo("Asia/Ho_Chi_Minh")
LOG = logging.getLogger("macro_collectors")


def stamp(value):
    return value.astimezone(timezone.utc).isoformat().replace("+00:00", "Z")


def target_period(now):
    local = now.astimezone(ZONE)
    year, month = local.year, local.month - 1
    if month == 0:
        year, month = year - 1, 12
    return f"{year:04d}-{month:02d}"


def next_month(now):
    local = now.astimezone(ZONE)
    year, month = local.year, local.month + 1
    if month == 13:
        year, month = year + 1, 1
    return datetime(year, month, 1, 7, 30, tzinfo=ZONE)


def atomic_write(path, data):
    path.parent.mkdir(parents=True, exist_ok=True)
    fd, name = tempfile.mkstemp(prefix=".pending-", dir=path.parent)
    try:
        with os.fdopen(fd, "wb") as output:
            output.write(data)
            output.flush()
            os.fsync(output.fileno())
        os.chmod(name, 0o644)
        os.replace(name, path)
        directory = os.open(path.parent, os.O_DIRECTORY)
        try:
            os.fsync(directory)
        finally:
            os.close(directory)
    finally:
        if os.path.exists(name):
            os.unlink(name)


@dataclass
class MonthlySchedule:
    hour: int = 7
    minute: int = 30
    release_window_days: int = 10

    def pending_overdue(self, now):
        return now.astimezone(ZONE).day > self.release_window_days

    def next_success(self, now):
        return next_month(now).replace(hour=self.hour, minute=self.minute)

    def next_pending(self, now):
        local = now.astimezone(ZONE)
        opening = local.replace(day=1, hour=self.hour, minute=self.minute, second=0, microsecond=0)
        if local < opening:
            return opening
        return now + timedelta(hours=6 if local.day <= self.release_window_days else 24)


@dataclass
class IntervalSchedule:
    interval: timedelta
    retry_interval: timedelta = timedelta(hours=1)

    def next_success(self, now):
        return now + self.interval

    def next_pending(self, now):
        return now + self.retry_interval

    def pending_overdue(self, now):
        return False


class Runner:
    """One independent state/lock per collector; schedule survives restarts."""
    def __init__(self, root, collectors):
        self.root, self.collectors = Path(root).resolve(), collectors

    def run(self, name, now=None, force=False):
        now = now or datetime.now(timezone.utc)
        collector = self.collectors[name]
        directory = self.root / name
        directory.mkdir(parents=True, exist_ok=True)
        with (directory / ".lock").open("a") as lock:
            try:
                fcntl.flock(lock, fcntl.LOCK_EX | fcntl.LOCK_NB)
            except BlockingIOError:
                return "BUSY"
            path = directory / "manifest.json"
            try:
                previous = json.loads(path.read_text()) if path.exists() else {}
                if not isinstance(previous, dict):
                    previous = {}
            except (ValueError, OSError):
                previous = {}
                LOG.warning("Collector state unreadable; rebuilding from source")
            try:
                due = datetime.fromisoformat(previous["nextDue"].replace("Z", "+00:00")) if previous.get("nextDue") else None
                if due is not None and due.tzinfo is None:
                    due = None
            except (ValueError, TypeError, AttributeError):
                due = None
            if not force and due is not None and due > now:
                return "NOT_DUE"
            state = dict(previous)
            state.update(version=1, collector=name, targetPeriod=collector.target(now), checkedAt=stamp(now))
            schedule = collector.schedule
            failures = previous.get("consecutiveFailures", 0)
            try:
                result = collector.collect(directory, state["targetPeriod"], now)
                state.update(result)
                state.update(status="READY", errorKind="", message="Official release collected", consecutiveFailures=0, lastSuccess=stamp(now))
                next_due = schedule.next_success(now)
            except NotPublished:
                overdue = schedule.pending_overdue(now)
                state.update(status="ERROR" if overdue else "PENDING", errorKind="RELEASE_OVERDUE" if overdue else "", message="Expected report month not found in published releases", consecutiveFailures=failures + 1 if overdue else 0)
                next_due = schedule.next_pending(now)
            except CollectionError as error:
                failures += 1
                delay = 6 * 3600 if error.kind in ("ACCESS_DENIED", "DATA_QUALITY") else min(6 * 3600, 900 * 2 ** min(failures - 1, 5))
                next_due = now + timedelta(seconds=max(delay, error.retry_after))
                state.update(status="ERROR", errorKind=error.kind, message=str(error), consecutiveFailures=failures)
            except Exception:
                # Do not print response bodies, credential-bearing URLs or local paths.
                next_due = now + timedelta(hours=6)
                state.update(status="ERROR", errorKind="COLLECTOR", message="Collector failed; inspect configuration and source format", consecutiveFailures=failures + 1)
            state["nextDue"] = stamp(next_due)
            atomic_write(path, json.dumps(state, ensure_ascii=False, indent=2).encode())
            LOG.log(logging.WARNING if state["status"] == "ERROR" else logging.INFO, json.dumps({key: state.get(key) for key in ("collector", "status", "errorKind", "message", "targetPeriod", "consecutiveFailures", "nextDue")}))
            return state["status"]
