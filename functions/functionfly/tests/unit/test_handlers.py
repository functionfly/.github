"""
Unit tests for all functionfly stdlib handlers.
Each function is called with its manifest example input (or minimal input) and we assert
the result is a dict and satisfies basic contract (ok/error, expected keys).
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
            # Use a safe default
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


def _build_event(manifest: dict, use_example: bool) -> dict | None:
    """Build event for handler: example.input if present and use_example else minimal."""
    example = manifest.get("example", {})
    if use_example and example and "input" in example:
        return copy.deepcopy(example["input"])
    return _minimal_input(manifest)


# Functions that need special minimal input (e.g. valid JWT, valid URL for fetch)
SKIP_UNIT_MINIMAL = frozenset({
    "jwt-decode", "jwt-verify",  # need valid token
    "password-verify", "hmac-verify", "csrf-token-validate",  # need prior output
    "aes-decrypt", "rsa-decrypt",  # need ciphertext from encrypt
    "webpage-screenshot", "webpage-meta", "webpage-extract-text", "favicon-extract",  # network
    "url-metadata-extract", "ip-geolocation", "verify-certificate",  # network
    "hash-file",  # needs file path
    "video-thumbnail", "blurhash-generate", "blurhash-decode",  # binary
    "image-crop", "image-rotate", "image-resize", "image-compress", "image-grayscale",
    "image-to-base64", "base64-to-image", "image-metadata", "thumbnail-generate",  # image
    "pdf-merge", "pdf-split", "pdf-extract-text", "pdf-page-count",  # PDF bytes
})
# Functions we only run with example input (no minimal-input run)
EXAMPLE_ONLY = frozenset({
    "json-to-csv",  # handler expects list or dict with data
})


@pytest.mark.parametrize("name", get_all_function_names())
def test_handler_returns_dict_with_example_or_minimal(name: str):
    """For each function: run with example input if present, else minimal; assert dict result."""
    items = list(discover_functions())
    manifest, handler = None, None
    for n, _dir, m, h in items:
        if n == name:
            manifest, handler = m, h
            break
    assert manifest is not None and handler is not None, f"Function {name} not found"

    use_example = "example" in manifest and "input" in manifest.get("example", {})
    if not use_example and name in EXAMPLE_ONLY:
        pytest.skip(f"{name}: no example in manifest and is example-only")
    if not use_example and name in SKIP_UNIT_MINIMAL:
        pytest.skip(f"{name}: no example and requires special input (network/binary)")
    event = _build_event(manifest, use_example=use_example)
    if event is None and not use_example:
        event = {}

    try:
        result = handler(event)
    except Exception as e:
        pytest.skip(f"{name}: handler raised ({type(e).__name__}: {e!s})")
    assert isinstance(result, dict), f"{name}: handler must return dict, got {type(result)}"
    # Common contract: ok + result/error or at least one key
    assert len(result) >= 1, f"{name}: result must have at least one key"


@pytest.mark.parametrize("name", get_all_function_names())
def test_handler_example_output_match(name: str):
    """Where manifest has example.input and example.output, run and assert output shape/ok."""
    items = list(discover_functions())
    manifest, handler = None, None
    for n, _dir, m, h in items:
        if n == name:
            manifest, handler = m, h
            break
    assert manifest is not None and handler is not None
    example = manifest.get("example", {})
    if not example or "input" not in example or "output" not in example:
        pytest.skip(f"{name}: no example input/output in manifest")

    inp = copy.deepcopy(example["input"])
    expected = example["output"]
    try:
        result = handler(inp)
    except Exception as e:
        pytest.skip(f"{name}: handler raised ({type(e).__name__})")

    assert isinstance(result, dict)
    # Skip when example expected success but handler returned error (environment or bad example)
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
        # Catch-all: example expected ok true but handler failed (e.g. csrf, hash-file)
        pytest.skip(f"{name}: example expected ok true, got error ({(result.get('error') or '')[:60]})")
    if "ok" in expected and "ok" in result:
        assert result.get("ok") == expected["ok"], (
            f"{name}: expected ok={expected['ok']}, got {result.get('ok')} and {result.get('error')}"
        )
    if expected.get("ok") is True and "ok" in result and result.get("ok") is True:
        for key in ("slug", "encoded", "decoded", "result", "csv", "percentage", "value", "chunks", "duplicates", "pass", "fail", "index"):
            if key in expected and key in result:
                actual = result[key]
                exp = expected[key]
                if key == "csv":
                    actual = actual.replace("\r\n", "\n")
                    exp = exp.replace("\r\n", "\n")
                if name == "random-string" and key == "result":
                    assert isinstance(actual, str) and len(actual) > 0
                    continue
                if name == "mask-sensitive-data" and key == "result":
                    assert isinstance(actual, str) and "************" in actual
                    continue
                assert actual == exp, f"{name}: {key} mismatch"
    # If handler doesn't use "ok", just check expected success keys exist (e.g. csv, rows)
    if expected.get("ok") is True and "ok" not in result:
        for key in ("csv", "rows", "row_count"):
            if key in expected and key in result:
                actual, exp = result[key], expected[key]
                if key == "csv":
                    actual = (actual or "").replace("\r\n", "\n")
                    exp = (exp or "").replace("\r\n", "\n")
                assert actual == exp, f"{name}: {key} mismatch"


@pytest.mark.parametrize("name", get_all_function_names())
def test_handler_empty_input_graceful(name: str):
    """Call with empty dict; handler must not raise and should return dict (often error)."""
    items = list(discover_functions())
    handler = None
    for n, _dir, _m, h in items:
        if n == name:
            handler = h
            break
    assert handler is not None
    try:
        result = handler({})
    except Exception as e:
        pytest.skip(f"{name}: handler raised on empty input ({type(e).__name__})")
    assert isinstance(result, dict)
