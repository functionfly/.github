import re
from datetime import datetime


def handler(event):
    code = event.get("code") if isinstance(event, dict) else None
    rules = event.get("rules", {})
    cart_total = event.get("cart_total")
    if not code:
        return {"ok": False, "error": "code is required"}
    try:
        c = str(code).upper().strip()
        valid = True
        errors = []
        # Pattern check
        pattern = rules.get("pattern")
        if pattern and not re.match(pattern, c):
            valid = False; errors.append("code format invalid")
        # Prefix check
        prefix = rules.get("prefix")
        if prefix and not c.startswith(prefix.upper()):
            valid = False; errors.append(f"code must start with {prefix}")
        # Expiry check
        expires_at = rules.get("expires_at")
        if expires_at:
            exp = datetime.fromisoformat(str(expires_at))
            if datetime.utcnow() > exp:
                valid = False; errors.append("coupon expired")
        # Min cart total
        min_cart = rules.get("min_cart_total")
        if min_cart and cart_total is not None and float(cart_total) < float(min_cart):
            valid = False; errors.append(f"minimum cart total {min_cart} not met")
        return {"ok": True, "result": valid, "valid": valid, "code": c, "errors": errors}
    except Exception as e:
        return {"ok": False, "error": str(e)}
