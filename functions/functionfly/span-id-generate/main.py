import secrets


def handler(event):
    """Generate a distributed tracing span ID."""
    try:
        fmt = event.get("format", "w3c")

        # W3C TraceContext span ID: 16 hex chars (64-bit)
        span_id = secrets.token_hex(8)

        return {"ok": True, "span_id": span_id, "format": fmt}
    except Exception as e:
        return {"ok": False, "error": str(e)}
