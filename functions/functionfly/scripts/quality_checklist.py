#!/usr/bin/env python3
"""
Phase 1 Quality Standards Checklist — automated checks across all 150 functions.
Run from repo root: python functions/functionfly/scripts/quality_checklist.py
"""
from __future__ import annotations

import json
import re
from pathlib import Path

FUNCTIONS_DIR = Path(__file__).resolve().parent.parent


def strip_jsonc(raw: str) -> str:
    raw = re.sub(r"//[^\n]*", "", raw)
    raw = re.sub(r"/\*.*?\*/", "", raw, flags=re.DOTALL)
    return raw.strip()


def load_manifest(function_dir: Path) -> dict | None:
    path = function_dir / "functionfly.jsonc"
    if not path.is_file():
        return None
    try:
        raw = path.read_text(encoding="utf-8")
        return json.loads(strip_jsonc(raw))
    except (json.JSONDecodeError, OSError):
        return None


def main() -> None:
    total = 0
    has_title = has_desc = has_category = has_tags = 0
    has_example_in = has_example_out = has_deterministic = has_timeout = 0
    title_ok = desc_ok = 0

    for child in sorted(FUNCTIONS_DIR.iterdir()):
        if not child.is_dir() or not (child / "functionfly.jsonc").is_file():
            continue
        if child.name.startswith(".") or child.name == "scripts" or child.name == "tests":
            continue
        total += 1
        m = load_manifest(child)
        if not m:
            continue
        if m.get("title"):
            has_title += 1
            if len(m["title"]) <= 50:
                title_ok += 1
        if m.get("description"):
            has_desc += 1
            if len(m["description"]) <= 200:
                desc_ok += 1
        if m.get("category"):
            has_category += 1
        if m.get("tags") and len(m["tags"]) > 0:
            has_tags += 1
        ex = m.get("example") or {}
        if ex.get("input") is not None:
            has_example_in += 1
        if ex.get("output") is not None:
            has_example_out += 1
        if "deterministic" in m:
            has_deterministic += 1
        if m.get("timeout_ms") is not None:
            has_timeout += 1

    print("Phase 1 Quality Standards Checklist (automated)")
    print("=" * 50)
    print(f"Total functions: {total}")
    print(f"  title present:        {has_title}/150 (≤50 chars: {title_ok})")
    print(f"  description present:  {has_desc}/150 (≤200 chars: {desc_ok})")
    print(f"  category assigned:     {has_category}/150")
    print(f"  tags present:         {has_tags}/150")
    print(f"  example.input:        {has_example_in}/150")
    print(f"  example.output:       {has_example_out}/150")
    print(f"  deterministic set:   {has_deterministic}/150")
    print(f"  timeout_ms set:      {has_timeout}/150")
    if total == 150 and has_category == 150 and has_title == 150 and has_desc == 150:
        print("\n✓ All 150 functions meet automated checklist (title, description, category).")
    else:
        print("\n⚠ Some functions need manifest updates.")


if __name__ == "__main__":
    main()
