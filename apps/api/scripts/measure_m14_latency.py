#!/usr/bin/env python3
"""Collect serial, authenticated HTTP latency samples for representative M14 routes."""

from __future__ import annotations

import argparse
import csv
import http.client
import json
import random
import time
from collections import Counter
from pathlib import Path


ROUTES = [
    "/api/me",
    "/api/world",
    "/api/cases",
    "/api/whiteboards",
    "/api/world/npcs?q=a",
    "/api/world/newspaper",
]


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


def percentile(sorted_values: list[float], quantile: float) -> float:
    index = max(0, min(len(sorted_values) - 1, int(quantile * len(sorted_values) + 0.999999) - 1))
    return sorted_values[index]


def main() -> None:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--host", default="127.0.0.1")
    parser.add_argument("--port", type=int, default=8080)
    parser.add_argument("--cookie-file", type=Path, required=True)
    parser.add_argument("--csv", type=Path, required=True, help="output CSV for warmups and measured samples")
    parser.add_argument("--route", choices=ROUTES, help="measure one route instead of the full route set")
    parser.add_argument("--samples", type=int, default=40)
    parser.add_argument("--warmups", type=int, default=5)
    parser.add_argument("--cold-route", default="/api/world/npcs?q=a", help="route to call once before warmups, for a cold-cache sample")
    parser.add_argument("--seed", type=int, default=14014, help="fixed seed for route ordering")
    parser.add_argument("--timeout", type=float, default=45)
    args = parser.parse_args()
    if args.samples < 1 or args.warmups < 0:
        parser.error("samples must be positive and warmups cannot be negative")

    cookie = "noir_session=" + session_cookie(args.cookie_file)
    routes = [args.route] if args.route else ROUTES
    rng = random.Random(args.seed)
    conn = http.client.HTTPConnection(args.host, args.port, timeout=args.timeout)
    rows: list[tuple[str, int, str, int, str, str]] = []

    def request(route: str, phase: str, iteration: int) -> tuple[int, str]:
        started = time.perf_counter()
        try:
            conn.request(
                "GET",
                route,
                headers={"Cookie": cookie, "Connection": "keep-alive", "Accept": "application/json"},
            )
            response = conn.getresponse()
            response.read()
            elapsed = time.perf_counter() - started
            status = response.status
            error = ""
        except Exception as exc:  # retain failed requests in the measurement artifact
            elapsed = time.perf_counter() - started
            status = 0
            error = f"{type(exc).__name__}: {exc}"
            conn.close()
        rows.append((phase, iteration, route, status, f"{elapsed:.9f}", error))
        return status, error

    try:
        status, error = request(args.cold_route, "cold", 1)
        if status != 200:
            raise SystemExit(f"cold-cache request failed: {args.cold_route}; status={status}; {error}")
        for iteration in range(args.warmups):
            order = routes[:]
            rng.shuffle(order)
            for route in order:
                status, error = request(route, "warmup", iteration + 1)
                if status != 200:
                    raise SystemExit(f"warmup failed: {route}; status={status}; {error}")
        for iteration in range(args.samples):
            order = routes[:]
            rng.shuffle(order)
            for route in order:
                request(route, "sample", iteration + 1)
    finally:
        conn.close()

    args.csv.parent.mkdir(parents=True, exist_ok=True)
    with args.csv.open("w", newline="") as output:
        writer = csv.writer(output)
        writer.writerow(["phase", "iteration", "route", "status", "seconds", "error"])
        writer.writerows(rows)

    summary = {}
    cold_rows = [row for row in rows if row[0] == "cold" and row[2] == args.cold_route]
    if cold_rows:
        cold_row = cold_rows[0]
        summary["cold_cache"] = {
            "route": cold_row[2],
            "status": cold_row[3],
            "latency_ms": round(float(cold_row[4]) * 1000, 2),
            "error": cold_row[5],
        }
    for route in routes:
        route_rows = [row for row in rows if row[0] == "sample" and row[2] == route]
        values = sorted(float(row[4]) for row in route_rows if row[3] == 200)
        statuses = dict(Counter(str(row[3]) for row in route_rows))
        if values:
            summary[route] = {
                "samples_200": len(values),
                "statuses": statuses,
                "min_ms": round(values[0] * 1000, 2),
                "p50_ms_nearest_rank": round(percentile(values, 0.50) * 1000, 2),
                "p95_ms_nearest_rank": round(percentile(values, 0.95) * 1000, 2),
                "max_ms": round(values[-1] * 1000, 2),
                "mean_ms": round(sum(values) / len(values) * 1000, 2),
            }
        else:
            summary[route] = {"samples_200": 0, "statuses": statuses}
    print(json.dumps(summary, indent=2))


if __name__ == "__main__":
    main()
