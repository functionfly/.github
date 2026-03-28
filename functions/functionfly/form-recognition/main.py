import re

COMMON_FIELD_PATTERNS = {
    "name": [
        r"(?:full\s+)?name\s*[:=]\s*(.+)",
        r"(?:first|last)\s+name\s*[:=]\s*(.+)"
    ],
    "email": [
        r"e-?mail\s*(?:address)?\s*[:=]\s*(\S+@\S+\.\S+)",
        r"(\S+@\S+\.\S+)"
    ],
    "phone": [
        r"(?:phone|tel|mobile|cell)\s*(?:number)?\s*[:=]\s*([\d\s\-\+\(\)]+)",
        r"(\+?[\d\s\-\(\)]{7,})"
    ],
    "address": [
        r"(?:street\s+)?address\s*[:=]\s*(.+)",
        r"(?:city|state|zip|postal)\s*[:=]\s*(.+)"
    ],
    "date": [
        r"date\s*[:=]\s*(\d{1,2}[/\-\.]\d{1,2}[/\-\.]\d{2,4})",
        r"(\d{4}-\d{2}-\d{2})"
    ],
    "amount": [
        r"(?:amount|total|price|cost|fee)\s*[:=]\s*\$?([\d,]+(?:\.\d{2})?)",
        r"\$\s*([\d,]+(?:\.\d{2})?)"
    ],
    "id": [
        r"(?:id|number|no|#)\s*[:=]\s*([A-Z0-9\-]+)",
        r"(?:invoice|order|ref)\s*#?\s*[:=]?\s*([A-Z0-9\-]+)"
    ],
    "signature": [
        r"signature\s*[:=]\s*(.+)",
        r"signed\s+by\s*[:=]\s*(.+)"
    ],
}


def handler(event):
    text = event.get("text") if isinstance(event, dict) else None
    if not text or not isinstance(text, str):
        return {"ok": False, "error": "text (string) is required"}
    try:
        custom_fields = event.get("fields", [])
        extracted = {}
        # Extract using common patterns
        for field, patterns in COMMON_FIELD_PATTERNS.items():
            for pattern in patterns:
                m = re.search(pattern, text, re.IGNORECASE)
                if m:
                    extracted[field] = {"value": m.group(1).strip(), "confidence": 0.85, "pattern": "auto"}
                    break
        # Extract custom fields
        for field in custom_fields:
            if isinstance(field, str):
                pattern = re.escape(field) + r"\s*[:=]\s*(.+)"
                m = re.search(pattern, text, re.IGNORECASE)
                if m:
                    extracted[field] = {"value": m.group(1).strip(), "confidence": 0.9, "pattern": "custom"}
        # Also extract key-value pairs generically
        kv_pairs = re.findall(r"([A-Za-z][A-Za-z ]{1,30})\s*[:=]\s*([^\n]+)", text)
        for key, value in kv_pairs:
            key_clean = key.strip().lower().replace(" ", "_")
            if key_clean not in extracted:
                extracted[key_clean] = {"value": value.strip(), "confidence": 0.7, "pattern": "generic"}
        return {
            "ok": True,
            "result": extracted,
            "fields": extracted,
            "field_count": len(extracted),
            "note": "Rule-based form recognition — for production use, integrate Azure Form Recognizer or similar"
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
