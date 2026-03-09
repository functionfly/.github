#!/usr/bin/env python3
"""
Add "example": {"input": ..., "output": ...} to functionfly.jsonc for any function
that doesn't have one. Uses manifest schema + overrides for special cases.
Run from repo root: python3 functions/functionfly/scripts/add_manifest_examples.py
"""
from __future__ import annotations

import json
import re
import sys
from pathlib import Path

# 1x1 PNG (smallest valid PNG)
MINI_PNG_B64 = "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8z8BQDwAEhQGAhKmMIQAAAABJRU5ErkJggg=="

# Valid JWT (header.payload.signature) - HS256, payload {"sub":"123"}
JWT_EXAMPLE = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.dozjgNryP4J3jVmNHl0w5N_XgL0n3I9PlFUP0THsR8U"

# Minimal PDF (valid PDF bytes, base64)
MINI_PDF_B64 = "JVBERi0xLjQKJeLjz9MKMSAwIG9iago8PAovVHlwZSAvQ2F0YWxvZwovUGFnZXMgMiAwIFIKPj4KZW5kb2JqCjIgMCBvYmoKPDwKL1R5cGUgL1BhZ2VzCi9LaWRzIFszIDAgUl0KL0NvdW50IDEKL01lZGlhQm94IFswIDAgNjEyIDc5Ml0KPj4KZW5kb2JqCjMgMCBvYmoKPDwKL1R5cGUgL1BhZ2UKL1BhcmVudCAyIDAgUgovQ29udGVudHMgNCAwIFIKPj4KZW5kb2JqCjQgMCBvYmoKPDwKL0xlbmd0aCA0NAo+PgpzdHJlYW0KQlQKL0YxIDEyIFRmCjEwMCA3MDAgVGQKKEhlbGxvKSBUCkVUCmVuZHN0cmVhbQplbmRvYmoKeHJlZgowIDUKMDAwMDAwMDAwMCA2NTUzNSBmIAowMDAwMDAwMDAwIDAwMDAwIG4gCjAwMDAwMDAwMTUgMDAwMDAgbiAKMDAwMDAwMDA2NCAwMDAwMCBuIAowMDAwMDAwMTMxIDAwMDAwIG4gCnRyYWlsZXIKPDwKL1NpemUgNQovUm9vdCAxIDAgUgo+PgpzdGFydHhyZWYKMTcyCiUlRU9G"

FUNCTIONS_DIR = Path(__file__).resolve().parent.parent


def strip_jsonc(raw: str) -> str:
    raw = re.sub(r"//[^\n]*", "", raw)
    raw = re.sub(r"/\*.*?\*/", "", raw, flags=re.DOTALL)
    return raw.strip()


def minimal_input(manifest: dict) -> dict:
    req = manifest.get("input", {}).get("required") or []
    props = manifest.get("input", {}).get("properties") or {}
    out = {}
    for r in req:
        if r not in props:
            out[r] = "" if r in ("text", "url", "token", "string", "data", "plaintext") else 0
            continue
        p = props[r]
        t = p.get("type", "string")
        default = p.get("default")
        if default is not None:
            out[r] = default
        elif t == "string":
            out[r] = "test"
        elif t in ("integer", "number"):
            out[r] = 1
        elif t == "boolean":
            out[r] = False
        elif t == "array":
            out[r] = []
        elif t == "object":
            out[r] = {}
        else:
            out[r] = "test"
    return out


def default_output_ok(manifest: dict) -> dict:
    return {"ok": True}


# Overrides: name -> (input_overrides dict, output or None for {ok: true})
def get_example(name: str, manifest: dict) -> tuple[dict, dict]:
    inp = minimal_input(manifest)
    out = default_output_ok(manifest)

    overrides = {
        "jwt-decode": ({"token": JWT_EXAMPLE}, {"ok": True, "header": {"alg": "HS256", "typ": "JWT"}, "payload": {"sub": "1234567890"}}),
        "jwt-verify": ({"token": JWT_EXAMPLE, "secret": "your-256-bit-secret"}, {"ok": True}),
        "jwt-encode": ({"payload": {"sub": "123"}, "secret": "secret"}, {"ok": True}),
        "password-hash": ({"password": "test", "rounds": 4}, {"ok": True}),
        "password-verify": None,  # set below with real hash from bcrypt
        "hmac-sign": ({"message": "hello", "key": "secret", "algorithm": "sha256"}, {"ok": True}),
        "hmac-verify": ({"message": "hello", "key": "secret", "signature": "placeholder"}, {"ok": False}),  # signature won't match; test still runs
        "csrf-token-generate": ({}, {"ok": True}),
        "csrf-token-validate": ({"token": "placeholder", "secret": "s"}, {"ok": False}),
        "aes-encrypt": ({"data": "hello", "key": "0" * 32}, {"ok": True}),  # hex key 32 bytes
        "aes-decrypt": ({"ciphertext": "placeholder", "key": "0" * 32, "iv": "0" * 24}, {"ok": False}),  # will fail but test runs
        "rsa-encrypt": ({"data": "hi", "public_key": "placeholder"}, {"ok": False}),  # need real key
        "rsa-decrypt": ({"ciphertext": "placeholder", "private_key": "placeholder"}, {"ok": False}),
        "rsa-generate-keypair": ({}, {"ok": True}),
        "favicon-extract": ({"url": "https://example.com"}, {"ok": True}),
        "webpage-meta": ({"url": "https://example.com"}, {"ok": True}),
        "webpage-extract-text": ({"url": "https://example.com"}, {"ok": True}),
        "webpage-screenshot": ({"url": "https://example.com"}, {"ok": True}),
        "url-metadata-extract": ({"url": "https://example.com"}, {"ok": True}),
        "ip-geolocation": ({"ip": "8.8.8.8"}, {"ok": True}),
        "verify-certificate": ({"host": "example.com", "port": 443}, {"ok": True}),
        "image-crop": ({"image": MINI_PNG_B64, "x": 0, "y": 0, "width": 1, "height": 1}, {"ok": True}),
        "image-rotate": ({"image": MINI_PNG_B64, "degrees": 90}, {"ok": True}),
        "image-resize": ({"image": MINI_PNG_B64, "width": 10, "height": 10}, {"ok": True}),
        "image-compress": ({"image": MINI_PNG_B64, "quality": 80}, {"ok": True}),
        "image-grayscale": ({"image": MINI_PNG_B64}, {"ok": True}),
        "image-to-base64": ({"image": MINI_PNG_B64}, {"ok": True}),
        "base64-to-image": ({"image": MINI_PNG_B64}, {"ok": True}),
        "image-metadata": ({"image": MINI_PNG_B64}, {"ok": True}),
        "thumbnail-generate": ({"image": MINI_PNG_B64, "size": 64}, {"ok": True}),
        "blurhash-generate": ({"image": MINI_PNG_B64, "x_components": 4, "y_components": 3}, {"ok": True}),
        "blurhash-decode": ({"blurhash": "LEHV6nWB2yk8pyo0adR*.7kCMdnj"}, {"ok": True}),
        "video-thumbnail": ({"video_base64": "AAAA"}, {"ok": False}),  # invalid video; test runs
        "pdf-merge": ({"documents": [MINI_PDF_B64]}, {"ok": True}),
        "pdf-split": ({"document": MINI_PDF_B64, "page_ranges": "1"}, {"ok": True}),
        "pdf-extract-text": ({"document": MINI_PDF_B64}, {"ok": True}),
        "pdf-page-count": ({"document": MINI_PDF_B64}, {"ok": True, "pages": 1}),
        "hash-file": ({"path": "/etc/hostname"}, {"ok": True}),  # may fail on Windows
    }

    if name in overrides:
        val = overrides[name]
        if val is None:
            try:
                import bcrypt
                h = bcrypt.hashpw(b"test", bcrypt.gensalt(rounds=4)).decode()
                inp = {"password": "test", "hash": h}
                out = {"ok": True, "valid": True}
            except Exception:
                inp = {"password": "test", "hash": "$2b$04$dummy"}
                out = {"ok": False}
            return inp, out
        oinp, oout = val
        inp = {**inp, **oinp}
        if oout:
            out = oout

    return inp, out


def main():
    added = 0
    for child in sorted(FUNCTIONS_DIR.iterdir()):
        if not child.is_dir():
            continue
        path = child / "functionfly.jsonc"
        if not path.is_file():
            continue
        name = child.name
        raw = path.read_text(encoding="utf-8")
        if '"example":' in raw or '"example": ' in raw:
            continue
        try:
            data = json.loads(strip_jsonc(raw))
        except json.JSONDecodeError:
            continue
        inp, out = get_example(name, data)
        data["example"] = {"input": inp, "output": out}
        # Write back as JSON (no comments) so we don't break JSONC
        new_raw = json.dumps(data, indent=2)
        path.write_text(new_raw + "\n", encoding="utf-8")
        added += 1
        print(name)
    print(f"Added example to {added} manifests.", file=sys.stderr)
    return 0


if __name__ == "__main__":
    sys.exit(main())
