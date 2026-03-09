import re


def handler(event):
    if isinstance(event, dict):
        text = event.get("text", event.get("data", ""))
        mask_char = event.get("mask_char", "*")
        show_last = event.get("show_last", 4)
    else:
        text = str(event) if event is not None else ""
        mask_char = "*"
        show_last = 4

    if text is None:
        return {"ok": False, "error": "Input 'text' is required"}

    try:
        show_last = max(0, min(int(show_last), 4))
    except (TypeError, ValueError):
        show_last = 4

    char = str(mask_char)[:1] if mask_char else "*"
    s = str(text)

    # Credit card: 13-19 digits, optional spaces/dashes
    def mask_digits(match):
        digits = re.sub(r"\D", "", match.group(0))
        if len(digits) < 13:
            return match.group(0)
        visible = digits[-show_last:] if show_last else ""
        return char * (len(digits) - len(visible)) + visible

    s = re.sub(r"[\d\s\-]{13,}", lambda m: mask_digits(m) if re.sub(r"\D", "", m.group(0)) else m.group(0), s)

    # SSN: 3-2-4 digits
    def mask_ssn(m):
        g3 = m.group(3)
        visible = g3[-show_last:] if show_last else ""
        return char * 3 + "-" + char * 2 + "-" + char * (4 - len(visible)) + visible

    s = re.sub(r"\b(\d{3})[\-\s]?(\d{2})[\-\s]?(\d{4})\b", mask_ssn, s)

    return {"ok": True, "result": s}

