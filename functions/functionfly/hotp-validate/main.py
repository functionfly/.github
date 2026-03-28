import hmac
import hashlib
import base64

def generate_hotp(secret, counter, digits=6):
    """Generate HOTP (HMAC-based OTP)"""
    try:
        secret_bytes = base64.b32decode(secret, casefold=True)
    except Exception:
        return ""
    counter_bytes = counter.to_bytes(8, 'big')
    hmac_hash = hmac.new(secret_bytes, counter_bytes, hashlib.sha1).digest()
    offset = hmac_hash[-1] & 0x0F
    code = int.from_bytes(hmac_hash[offset:offset+4], 'big') & 0x7FFFFFFF
    otp = str(code % (10 ** digits)).zfill(digits)
    return otp

def handler(event):
    try:
        hotp = event.get("hotp", "") if isinstance(event, dict) else ""
        secret = event.get("secret", "") if isinstance(event, dict) else ""
        counter = event.get("counter", 0) if isinstance(event, dict) else 0
        digits = event.get("digits", 6) if isinstance(event, dict) else 6
        if not hotp:
            return {"ok": False, "error": "hotp is required"}
        if not secret:
            return {"ok": False, "error": "secret is required"}
        if counter < 0:
            return {"ok": False, "error": "counter must be non-negative"}
        if digits < 6 or digits > 8:
            return {"ok": False, "error": "digits must be 6, 7, or 8"}
        expected_hotp = generate_hotp(secret, counter, digits)
        valid = hotp == expected_hotp
        return {"ok": True, "valid": valid}
    except Exception as e:
        return {"ok": False, "error": str(e)}
