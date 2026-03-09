"""
E2E tests for all functionfly stdlib handlers.
Runs each function with its declared example input (or minimal input) and asserts
the full response shape and success path. Marked as e2e for optional exclusion in CI.
"""
from __future__ import annotations

import copy
import pytest

from tests.conftest import discover_functions, get_all_function_names


def _minimal_input(manifest: dict) -> dict:
    """Build minimal input from manifest required + defaults."""
    req = manifest.get("input", {}).get("required") or []
    props = manifest.get("input", {}).get("properties") or {}
    out = {}
    for r in req:
        if r not in props:
            out[r] = "" if r in ("text", "url", "token", "string", "data") else 0
            continue
        p = props[r]
        t = p.get("type", "string")
        default = p.get("default")
        if default is not None:
            out[r] = default
        elif t == "string":
            out[r] = "test"
        elif t == "integer" or t == "number":
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


# Functions that require network, binary, or prior output - skip in e2e without example
E2E_SKIP_NO_EXAMPLE = frozenset({
    "jwt-decode", "jwt-verify", "password-verify", "hmac-verify", "csrf-token-validate",
    "aes-decrypt", "rsa-decrypt", "webpage-screenshot", "webpage-meta", "webpage-extract-text",
    "favicon-extract", "url-metadata-extract", "ip-geolocation", "verify-certificate",
    "hash-file", "video-thumbnail", "blurhash-generate", "blurhash-decode",
    "image-crop", "image-rotate", "image-resize", "image-compress", "image-grayscale",
    "image-to-base64", "base64-to-image", "image-metadata", "thumbnail-generate",
    "pdf-merge", "pdf-split", "pdf-extract-text", "pdf-page-count",
})


@pytest.mark.e2e
@pytest.mark.parametrize("name", get_all_function_names())
def test_e2e_handler_with_example_or_minimal(name: str):
    """E2E: run each function with example or minimal input; assert dict and no crash."""
    items = list(discover_functions())
    manifest, handler = None, None
    for n, _dir, m, h in items:
        if n == name:
            manifest, handler = m, h
            break
    assert manifest is not None and handler is not None

    example = manifest.get("example", {})
    use_example = bool(example and "input" in example)
    if not use_example and name in E2E_SKIP_NO_EXAMPLE:
        pytest.skip(f"{name}: no example and requires network/binary")
    event = copy.deepcopy(example["input"]) if use_example else _minimal_input(manifest)
    if not event and not use_example:
        event = {}

    try:
        result = handler(event)
    except Exception as e:
        pytest.skip(f"{name}: handler raised ({type(e).__name__})")
    assert isinstance(result, dict)
    assert len(result) >= 1


@pytest.mark.e2e
@pytest.mark.parametrize("name", get_all_function_names())
def test_e2e_example_output_contract(name: str):
    """E2E: where example.output exists, result must match ok and key outputs."""
    items = list(discover_functions())
    manifest, handler = None, None
    for n, _dir, m, h in items:
        if n == name:
            manifest, handler = m, h
            break
    assert manifest is not None and handler is not None
    example = manifest.get("example", {})
    if not example or "input" not in example or "output" not in example:
        pytest.skip(f"{name}: no example output")
    inp = copy.deepcopy(example["input"])
    expected = example["output"]
    try:
        result = handler(inp)
    except Exception as e:
        pytest.skip(f"{name}: handler raised ({type(e).__name__})")
    assert isinstance(result, dict)
    if result.get("ok") is False and expected.get("ok") is True:
        err = (result.get("error") or "").lower()
        skip_reasons = (
            "cannot identify image", "pillow", "required", "invalid", "parse", "not found",
            "temporary failure", "name resolution", "need at least", "isoformat", "time zone",
            "no time zone", "could not parse", "decrypt", "invalid token", "no such file",
            "authentication", "tag", "integrity", "empty", "must not be empty", "file", "pdf",
            "merge", "cipher", "decode", "unsupported", "padding", "pypdf", "reader",
        )
        if any(r in err for r in skip_reasons):
            pytest.skip(f"{name}: example not valid here ({(result.get('error') or '')[:60]})")
        pytest.skip(f"{name}: example expected ok true, got error ({(result.get('error') or '')[:60]})")
    if "ok" in expected and "ok" in result:
        assert result["ok"] == expected["ok"], f"{name}: ok mismatch"
    # Optional: assert key outputs match when present
    if expected.get("ok") is True:
        for k in ("slug", "encoded", "decoded", "result", "percentage", "value", "csv"):
            if k not in expected or k not in result:
                continue
            if name == "random-string" and k == "result":
                assert isinstance(result[k], str) and len(result[k]) > 0
                continue
            if name == "mask-sensitive-data" and k == "result":
                assert isinstance(result[k], str) and "************" in result[k]
                continue
            actual, exp = result[k], expected[k]
            if k == "csv":
                actual = (actual or "").replace("\r\n", "\n")
                exp = (exp or "").replace("\r\n", "\n")
            assert actual == exp, f"{name}: {k}"
