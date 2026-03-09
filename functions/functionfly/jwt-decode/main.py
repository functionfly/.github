import base64
import json


def _b64_decode(s):
    pad = 4 - len(s) % 4
    if pad != 4:
        s += "=" * pad
    return base64.urlsafe_b64decode(s)


def handler(event):
    if isinstance(event, dict):
        token = event.get("token", event.get("jwt", ""))
    else:
        token = ""

    if not token:
        return {"ok": False, "error": "Input 'token' is required"}

    parts = str(token).strip().split(".")
    if len(parts) != 3:
        return {"ok": False, "error": "Invalid JWT format (expected 3 parts)"}

    try:
        header = json.loads(_b64_decode(parts[0]).decode("utf-8"))
        payload = json.loads(_b64_decode(parts[1]).decode("utf-8"))
        return {"ok": True, "header": header, "payload": payload}
    except Exception as e:
        return {"ok": False, "error": str(e)}

