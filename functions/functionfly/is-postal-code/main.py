import re

POSTAL_CODES = {
    "US": re.compile(r'^\d{5}(-\d{4})?$'),
    "GB": re.compile(r'^[A-Z]{1,2}[0-9][0-9A-Z]?\s?[0-9][A-Z]{2}$', re.I),
    "CA": re.compile(r'^[ABCEGHJKLMNPRSTVXY]\d[A-Z]\s?\d[A-Z]\d$', re.I),
    "DE": re.compile(r'^\d{5}$'),
    "FR": re.compile(r'^\d{5}$'),
    "AU": re.compile(r'^\d{4}$'),
    "JP": re.compile(r'^\d{3}-?\d{4}$'),
    "IN": re.compile(r'^\d{6}$'),
    "BR": re.compile(r'^\d{5}-?\d{3}$'),
    "NL": re.compile(r'^\d{4}\s?[A-Z]{2}$', re.I),
    "PL": re.compile(r'^\d{2}-\d{3}$'),
    "IT": re.compile(r'^\d{5}$'),
    "ES": re.compile(r'^\d{5}$'),
    "SE": re.compile(r'^\d{3}\s?\d{2}$'),
    "NO": re.compile(r'^\d{4}$'),
    "DK": re.compile(r'^\d{4}$'),
    "FI": re.compile(r'^\d{5}$'),
    "CH": re.compile(r'^\d{4}$'),
    "AT": re.compile(r'^\d{4}$'),
    "MX": re.compile(r'^\d{5}$'),
}

GENERIC_RE = re.compile(r'^[A-Z0-9][A-Z0-9\s\-]{2,9}[A-Z0-9]$', re.I)


def handler(event):
    value = event.get("value") if isinstance(event, dict) else None
    country = event.get("country", "").upper()

    if value is None:
        return {"ok": False, "error": "value is required"}

    val = str(value).strip()

    if country and country in POSTAL_CODES:
        result = bool(POSTAL_CODES[country].match(val))
        return {"ok": True, "value": value, "result": result, "country": country}

    # Generic fallback
    result = bool(GENERIC_RE.match(val))
    return {"ok": True, "value": value, "result": result, "country": country or None}
