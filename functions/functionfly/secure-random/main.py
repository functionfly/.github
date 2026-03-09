import base64
import secrets


def handler(event):
    if isinstance(event, dict):
        lo = event.get("min", 0)
        hi = event.get("max", 999999)
        count = event.get("count", 1)
        nbytes = event.get("bytes")
    else:
        lo, hi, count, nbytes = 0, 999999, 1, None

    if nbytes is not None:
        try:
            n = max(1, min(1024, int(nbytes)))
        except (TypeError, ValueError):
            n = 32
        return {"ok": True, "bytes_base64": base64.urlsafe_b64encode(secrets.token_bytes(n)).decode("ascii")}

    try:
        lo = int(lo)
        hi = int(hi)
        count = max(1, min(100, int(count)))
    except (TypeError, ValueError):
        return {"ok": False, "error": "min, max, count must be integers"}

    if lo >= hi:
        return {"ok": False, "error": "min must be less than max"}

    try:
        values = [secrets.randbelow(hi - lo + 1) + lo for _ in range(count)]
    except Exception as e:
        return {"ok": False, "error": str(e)}
    return {"ok": True, "values": values}
