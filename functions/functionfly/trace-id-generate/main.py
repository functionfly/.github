import secrets
import uuid


def handler(event):
    """Generate a distributed tracing trace ID."""
    try:
        fmt = event.get("format", "w3c")

        if fmt == "w3c":
            # W3C TraceContext: 32 hex chars (128-bit)
            trace_id = secrets.token_hex(16)
        elif fmt == "b3":
            # B3 format: 32 hex chars (128-bit) or 16 hex chars (64-bit)
            trace_id = secrets.token_hex(16)
        elif fmt == "uuid":
            trace_id = str(uuid.uuid4())
        elif fmt == "hex":
            trace_id = secrets.token_hex(16)
        else:
            trace_id = secrets.token_hex(16)

        return {"ok": True, "trace_id": trace_id, "format": fmt}
    except Exception as e:
        return {"ok": False, "error": str(e)}
