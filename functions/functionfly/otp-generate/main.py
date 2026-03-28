import random

def generate_otp(length=6):
    """Generate a random OTP"""
    if length < 1:
        return ""
    otp = ''.join([str(random.randint(0, 9)) for _ in range(length)])
    return otp

def handler(event):
    try:
        length = event.get("length", 6) if isinstance(event, dict) else 6
        if length < 1:
            return {"ok": False, "error": "length must be at least 1"}
        otp = generate_otp(length)
        return {"ok": True, "otp": otp}
    except Exception as e:
        return {"ok": False, "error": str(e)}
