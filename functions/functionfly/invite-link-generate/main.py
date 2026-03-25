import random
import string
import hashlib
import time


def handler(event):
    base_url = event.get("base_url") if isinstance(event, dict) else None
    user_id = event.get("user_id")
    expires_in_hours = event.get("expires_in_hours")
    campaign = event.get("campaign", "")
    token_length = int(event.get("token_length", 12))
    if not base_url:
        return {"ok": False, "error": "base_url is required"}
    try:
        from urllib.parse import urlparse, urlencode, urlunparse
        token_length = max(8, min(token_length, 32))
        seed = f"{user_id}{time.time()}{campaign}"
        h = hashlib.sha256(seed.encode()).hexdigest()
        token = h[:token_length]
        params = {"ref": token}
        if campaign:
            params["c"] = str(campaign)
        if user_id:
            uid_short = hashlib.md5(str(user_id).encode()).hexdigest()[:6]
            params["uid"] = uid_short
        parsed = urlparse(str(base_url))
        query = urlencode(params)
        invite_url = urlunparse((parsed.scheme, parsed.netloc, parsed.path, parsed.params, query, ""))
        expires_at = None
        if expires_in_hours:
            expires_at = int(time.time()) + int(float(expires_in_hours) * 3600)
        return {
            "ok": True,
            "result": invite_url,
            "invite_url": invite_url,
            "token": token,
            "expires_at_unix": expires_at,
            "user_id": str(user_id) if user_id else None
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
