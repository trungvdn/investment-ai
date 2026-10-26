import argparse
import logging
import os
import signal
import threading

from .core import Runner
from .pmi import INDEX, PMICollector


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--data-dir", default=os.getenv("COLLECTOR_DATA_DIR", "/collector-data"))
    factories = {"vn_pmi": lambda: PMICollector(index=os.getenv("PMI_RELEASE_INDEX_URL") or INDEX)}
    parser.add_argument("--collector", choices=factories, default="vn_pmi")
    parser.add_argument("--once", action="store_true")
    parser.add_argument("--force", action="store_true", help="ignore due date; never bypass source lock")
    args = parser.parse_args()
    logging.basicConfig(level=logging.INFO, format="%(message)s")
    registry = {args.collector: factories[args.collector]()}
    runner = Runner(args.data_dir, registry)
    if args.once:
        return 1 if runner.run(args.collector, force=args.force) == "ERROR" else 0
    if args.force:
        parser.error("--force requires --once")
    stop = threading.Event()
    for name in (signal.SIGTERM, signal.SIGINT):
        signal.signal(name, lambda *_: stop.set())
    while not stop.is_set():
        runner.run(args.collector)
        stop.wait(60)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
