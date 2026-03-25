import os, hashlib, time


def handler(event):
    last4 = event.get("last4") if isinstance(event, dict) else None
    card_type = event.get("card_type", "unknown")
    expiry = event.get("expiry", "")
    prefix = event.get("prefix", "tok")
    if not last4:
        return {"ok": False, "error": "last4 is required"}
    try:
        last4_str = str(last4)
        if len(last4_str) != 4 or not last4_str.isdigit():
            return {"ok": False, "error": "last4 must be exactly 4 digits"}
        entropy = os.urandom(16).hex()
        ts = str(int(time.time() * 1000))
        token_hash = hashlib.sha256(f"{entropy}{last4_str}{ts}".encode()).hexdigest()[:24]
        token = f"{prefix}_{token_hash}"
        return {
            "ok": True,
            "result": token,
            "token": token,
            "last4": last4_str,
            "card_type": card_type,
            "created_at": int(time.time())
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
