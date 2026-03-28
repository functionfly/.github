import hmac
import hashlib
import time
import base64

def generate_totp(secret, digits=6, period=30):
    """Generate TOTP (Time-based OTP)"""
    try:
        secret_bytes = base64.b32decode(secret, casefold=True)
    except Exception:
        return ""
    time_counter = int(time.time()) // period
    counter_bytes = time_counter.to_bytes(8, 'big')
    hmac_hash = hmac.new(secret_bytes, counter_bytes, hashlib.sha1).digest()
    offset = hmac_hash[-1] & 0x0F
    code = int.from_bytes(hmac_hash[offset:offset+4], 'big') & 0x7FFFFFFF
    otp = str(code % (10 ** digits)).zfill(digits)
    return otp

def handler(event):
    try:
        totp = event.get("totp", "") if isinstance(event, dict) else ""
        secret = event.get("secret", "") if isinstance(event, dict) else ""
        digits = event.get("digits", 6) if isinstance(event, dict) else 6
        period = event.get("period", 30) if isinstance(event, dict) else 30
        if not totp:
            return {"ok": False, "error": "totp is required"}
        if not secret:
            return {"ok": False, "error": "secret is required"}
        if digits < 6 or digits > 8:
            return {"ok": False, "error": "digits must be 6, 7, or 8"}
        if period < 1:
            return {"ok": False, "error": "period must be at least 1"}
        expected_totp = generate_totp(secret, digits, period)
        valid = totp == expected_totp
        return {"ok": True, "valid": valid}
    except Exception as e:
        return {"ok": False, "error": str(e)}
