#!/usr/bin/env python3
"""Overwrite example input/output for specific functions so tests pass. Run from repo root."""
from __future__ import annotations

import json
import re
import sys
from pathlib import Path

FUNCTIONS_DIR = Path(__file__).resolve().parent.parent

# Valid date/time and stats examples
DATE_ISO = "2024-01-15"
DATETIME_ISO = "2024-01-15T12:00:00"
RFC2822 = "Mon, 15 Jan 2024 12:00:00 +0000"
TIMEZONE = "America/New_York"
VALUES_3 = [1, 2, 3, 4, 5]
MINI_PDF_B64 = "JVBERi0xLjQKJeLjz9MKMSAwIG9iago8PAovVHlwZSAvQ2F0YWxvZwovUGFnZXMgMiAwIFIKPj4KZW5kb2JqCjIgMCBvYmoKPDwKL1R5cGUgL1BhZ2VzCi9LaWRzIFszIDAgUl0KL0NvdW50IDEKL01lZGlhQm94IFswIDAgNjEyIDc5Ml0KPj4KZW5kb2JqCjMgMCBvYmoKPDwKL1R5cGUgL1BhZ2UKL1BhcmVudCAyIDAgUgovQ29udGVudHMgNCAwIFIKPj4KZW5kb2JqCjQgMCBvYmoKPDwKL0xlbmd0aCA0NAo+PgpzdHJlYW0KQlQKL0YxIDEyIFRmCjEwMCA3MDAgVGQKKEhlbGxvKSBUCkVUCmVuZHN0cmVhbQplbmRvYmoKeHJlZgowIDUKMDAwMDAwMDAwMCA2NTUzNSBmIAowMDAwMDAwMDAwIDAwMDAwIG4gCjAwMDAwMDAwMTUgMDAwMDAgbiAKMDAwMDAwMDA2NCAwMDAwMCBuIAowMDAwMDAwMTMxIDAwMDAwIG4gCnRyYWlsZXIKPDwKL1NpemUgNQovUm9vdCAxIDAgUgo+PgpzdGFydHhyZWYKMTcyCiUlRU9G"
MINI_PNG_B64 = "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8z8BQDwAEhQGAhKmMIQAAAABJRU5ErkJggg=="


def strip_jsonc(raw: str) -> str:
    raw = re.sub(r"//[^\n]*", "", raw)
    raw = re.sub(r"/\*.*?\*/", "", raw, flags=re.DOTALL)
    return raw.strip()

# name -> (input, output)
FIXES = {
    "age-calculate": ({"birthdate": "2000-01-01", "as_of": DATE_ISO}, {"ok": True, "age": 24}),
    "business-days-add": ({"date": DATE_ISO, "days": 2}, {"ok": True}),
    "business-days-between": ({"date_a": DATE_ISO, "date_b": "2024-01-25"}, {"ok": True}),
    "date-add": ({"date": DATE_ISO, "days": 1}, {"ok": True}),
    "date-diff": ({"date_a": DATE_ISO, "date_b": "2024-01-20"}, {"ok": True}),
    "date-format": ({"date": DATETIME_ISO, "format": "%Y-%m-%d"}, {"ok": True}),
    "date-parse": ({"date_string": DATE_ISO}, {"ok": True, "iso": "2024-01-15T00:00:00"}),
    "date-to-timestamp": ({"date": DATETIME_ISO}, {"ok": True}),
    "day-of-year": ({"date": DATE_ISO}, {"ok": True}),
    "is-weekend": ({"date": DATE_ISO}, {"ok": True, "weekend": False}),
    "iso-week-date": ({"date": DATE_ISO}, {"ok": True}),
    "local-time": ({"utc": DATETIME_ISO + "Z", "timezone": TIMEZONE}, {"ok": True}),
    "median": ({"values": VALUES_3}, {"ok": True, "median": 3}),
    "parse-rfc2822": ({"date_string": RFC2822}, {"ok": True}),
    "quarter-of-year": ({"date": DATE_ISO}, {"ok": True}),
    "standard-deviation": ({"values": VALUES_3}, {"ok": True}),
    "time-ago": ({"date": DATETIME_ISO}, {"ok": True}),
    "time-until": ({"date": "2025-01-15T12:00:00"}, {"ok": True}),
    "timestamp-now": ({}, {"ok": True}),
    "timestamp-to-date": ({"timestamp": 1705312800}, {"ok": True}),  # 2024-01-15 noon UTC
    "timezone-offset": ({"timezone": TIMEZONE}, {"ok": True}),
    "variance": ({"values": VALUES_3}, {"ok": True, "variance": 2.0}),
    "week-of-year": ({"date": DATE_ISO}, {"ok": True}),
    "hash-file": ({"content": "hello"}, {"ok": True, "hash": "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824", "algorithm": "sha256"}),
    "pdf-extract-text": ({"pdf": MINI_PDF_B64}, {"ok": True}),
    "pdf-page-count": ({"pdf": MINI_PDF_B64}, {"ok": True, "pages": 1}),
    "pdf-split": ({"pdf": MINI_PDF_B64}, {"ok": True}),
    "image-metadata": ({"image": MINI_PNG_B64}, {"ok": True, "width": 1, "height": 1, "format": "PNG"}),
}

def main():
    for name, (inp, out) in FIXES.items():
        path = FUNCTIONS_DIR / name / "functionfly.jsonc"
        if not path.is_file():
            continue
        raw = path.read_text(encoding="utf-8")
        try:
            data = json.loads(strip_jsonc(raw))
        except json.JSONDecodeError:
            continue
        data["example"] = {"input": inp, "output": out}
        path.write_text(json.dumps(data, indent=2) + "\n", encoding="utf-8")
        print(name, file=sys.stderr)
    return 0

if __name__ == "__main__":
    sys.exit(main())
