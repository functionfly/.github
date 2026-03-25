import json
import os
import time
import urllib.request
import urllib.error


def _env(name: str, default: str = "") -> str:
    v = os.environ.get(name)
    if v is None:
        return default
    return str(v)


def handler(event):
    """
    Production-ready webhook notifier example:
    - validates inputs
    - enforces HTTPS by default
    - sets timeouts
    - returns structured error payloads
    """
    webhook_url = _env("WEBHOOK_URL", "https://httpbin.org/post").strip()
    notification_type = _env("NOTIFICATION_TYPE", "notification").strip() or "notification"
    timeout_s = int(_env("WEBHOOK_TIMEOUT_SECONDS", "15"))

    if not webhook_url:
        return {"ok": False, "error": "WEBHOOK_URL is required"}
    if not (webhook_url.startswith("https://") or webhook_url.startswith("http://")):
        return {"ok": False, "error": "WEBHOOK_URL must start with http:// or https://"}

    payload = {
        "type": notification_type,
        "message": "Hello from FunctionFly webhook example!",
        "timestamp": int(time.time()),
        "source": "functionfly-webhook-notifier",
        "event": event,
    }

    body = json.dumps(payload).encode("utf-8")
    headers = {
        "Content-Type": "application/json",
        "User-Agent": "FunctionFly/webhook-notifier@1.0.0",
    }

    req = urllib.request.Request(webhook_url, data=body, headers=headers, method="POST")

    try:
        with urllib.request.urlopen(req, timeout=timeout_s) as resp:
            resp_body = resp.read()
            content_type = resp.headers.get("Content-Type", "")
            try:
                if "application/json" in content_type:
                    parsed = json.loads(resp_body.decode("utf-8"))
                else:
                    parsed = resp_body.decode("utf-8", errors="replace")
            except Exception:
                parsed = resp_body.decode("utf-8", errors="replace")

            return {
                "ok": True,
                "status_code": getattr(resp, "status", 200),
                "url": getattr(resp, "url", webhook_url),
                "result": parsed,
            }
    except urllib.error.HTTPError as e:
        return {"ok": False, "error": f"HTTP {e.code}: {e.reason}", "status_code": e.code}
    except urllib.error.URLError as e:
        return {"ok": False, "error": f"URL Error: {str(e)}"}
    except Exception as e:
        return {"ok": False, "error": f"Request failed: {str(e)}"}

