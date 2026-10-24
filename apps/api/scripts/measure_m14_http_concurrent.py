#!/usr/bin/env python3
"""Measure authenticated HTTP latency under controlled concurrency."""

from __future__ import annotations

import argparse
import csv
import http.client
import json
import threading
import time
from concurrent.futures import ThreadPoolExecutor
from pathlib import Path


ROUTES = ("/api/cases", "/api/world/npcs?q=a")


def session_cookie(path: Path) -> str:
    for line in path.read_text().splitlines():
        if line.startswith("#HttpOnly_"):
            line = line[len("#HttpOnly_") :]
        elif line.startswith("#") or not line.strip():
            continue
        fields = line.split("\t")
        if len(fields) >= 7 and fields[5] == "noir_session":
            return fields[6]
    raise SystemExit("cookie file does not contain noir_session")


def percentile(values: list[float], quantile: float) -> float:
    ordered = sorted(values)
    index = max(0, min(len(ordered) - 1, int(quantile * len(ordered) + 0.999999) - 1))
    return ordered[index]


def main() -> None:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--host", default="127.0.0.1")
    parser.add_argument("--port", type=int, default=8080)
    parser.add_argument("--cookie-file", type=Path, required=True)
    parser.add_argument("--csv", type=Path, required=True)
    parser.add_argument("--samples", type=int, default=100)
    parser.add_argument("--warmups", type=int, default=5)
    parser.add_argument("--concurrency", type=int, default=8)
    parser.add_argument("--timeout", type=float, default=60)
    args = parser.parse_args()
    if args.samples < 1 or args.warmups < 0 or args.concurrency < 1:
        parser.error("samples and concurrency must be positive; warmups cannot be negative")

    cookie = "noir_session=" + session_cookie(args.cookie_file)
    rows: list[tuple[str, str, int, int, float, str]] = []
    for route in ROUTES:
        # Warm the same route and retain every result in the artifact.
        conn = http.client.HTTPConnection(args.host, args.port, timeout=args.timeout)
        try:
            for iteration in range(1, args.warmups + 1):
                started = time.perf_counter()
                try:
                    conn.request("GET", route, headers={"Cookie": cookie, "Accept": "application/json"})
                    response = conn.getresponse()
                    response.read()
                    status, error = response.status, ""
                except Exception as exc:
                    status, error = 0, f"{type(exc).__name__}: {exc}"
                    conn.close()
                rows.append(("warmup", route, iteration, status, time.perf_counter() - started, error))
                if status != 200:
                    raise SystemExit(f"warmup failed for {route}: status={status} {error}")
        finally:
            conn.close()

        workers = min(args.concurrency, args.samples)
        barrier = threading.Barrier(workers)
        base, extra = divmod(args.samples, workers)

        def worker(worker_id: int) -> list[tuple[str, str, int, int, float, str]]:
            count = base + (1 if worker_id < extra else 0)
            conn = http.client.HTTPConnection(args.host, args.port, timeout=args.timeout)
            output = []
            try:
                barrier.wait(timeout=30)
                for local_index in range(count):
                    iteration = worker_id * base + min(worker_id, extra) + local_index + 1
                    started = time.perf_counter()
                    try:
                        conn.request("GET", route, headers={"Cookie": cookie, "Accept": "application/json"})
                        response = conn.getresponse()
                        response.read()
                        status, error = response.status, ""
                    except Exception as exc:
                        status, error = 0, f"{type(exc).__name__}: {exc}"
                        conn.close()
                    output.append(("sample", route, iteration, status, time.perf_counter() - started, error))
            finally:
                conn.close()
            return output

        with ThreadPoolExecutor(max_workers=workers) as executor:
            for worker_rows in executor.map(worker, range(workers)):
                rows.extend(worker_rows)

    args.csv.parent.mkdir(parents=True, exist_ok=True)
    with args.csv.open("w", newline="") as output:
        writer = csv.writer(output)
        writer.writerow(["phase", "route", "iteration", "status", "seconds", "error"])
        writer.writerows((phase, route, iteration, status, f"{elapsed:.9f}", error)
                         for phase, route, iteration, status, elapsed, error in rows)

    summary = {}
    for route in ROUTES:
        measured = [row for row in rows if row[0] == "sample" and row[1] == route]
        values = [row[4] for row in measured if row[3] == 200]
        statuses: dict[str, int] = {}
        for row in measured:
            statuses[str(row[3])] = statuses.get(str(row[3]), 0) + 1
        summary[route] = {
            "samples": len(measured),
            "samples_200": len(values),
            "statuses": statuses,
            "concurrency": workers,
            "min_ms": round(min(values) * 1000, 2) if values else None,
            "p50_ms_nearest_rank": round(percentile(values, .50) * 1000, 2) if values else None,
            "p95_ms_nearest_rank": round(percentile(values, .95) * 1000, 2) if values else None,
            "mean_ms": round(sum(values) / len(values) * 1000, 2) if values else None,
            "max_ms": round(max(values) * 1000, 2) if values else None,
        }
        if len(values) != args.samples:
            raise SystemExit(f"{route}: expected {args.samples} HTTP 200 samples, got {len(values)}; inspect {args.csv}")
    print(json.dumps(summary, indent=2))


if __name__ == "__main__":
    main()
