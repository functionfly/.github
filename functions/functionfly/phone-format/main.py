import re


def handler(event):
    """
    Format phone numbers toward E.164 (digits only, optional + prefix).
    Does not validate country code; strips non-digits and optionally adds +.

    Input:
        - phone: Phone number string (required)
        - e164: If true, return with leading + (default: true)

    Returns:
        - ok: True on success
        - formatted: E.164-style string (digits only or +digits)
        - digits: Digits only
        - error: Message if ok is False
    """
    if isinstance(event, dict):
        phone = event.get("phone", event.get("number", ""))
        e164 = event.get("e164", True)
    else:
        phone = str(event) if event else ""
        e164 = True

    if phone is None or (isinstance(phone, str) and not phone.strip()):
        return {"ok": False, "error": "Input 'phone' is required"}

    digits = re.sub(r"\D", "", str(phone))
    if not digits:
        return {"ok": False, "error": "No digits found in phone number"}

    if len(digits) > 15:
        return {"ok": False, "error": "Phone number too long for E.164"}

    formatted = "+" + digits if e164 else digits
    return {"ok": True, "formatted": formatted, "digits": digits}
