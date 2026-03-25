import re


RESERVED_NAMES = {
    "CON", "PRN", "AUX", "NUL", "COM1", "COM2", "COM3", "COM4",
    "COM5", "COM6", "COM7", "COM8", "COM9", "LPT1", "LPT2", "LPT3",
    "LPT4", "LPT5", "LPT6", "LPT7", "LPT8", "LPT9"
}


def handler(event):
    value = event.get("value") if isinstance(event, dict) else None
    replacement = event.get("replacement", "_")
    max_length = event.get("max_length", 255)

    if not value:
        return {"ok": False, "error": "value is required"}
    if not isinstance(value, str):
        return {"ok": False, "error": "value must be a string"}

    # Remove/replace illegal characters
    sanitized = re.sub(r'[<>:"/\\|?*\x00-\x1f]', replacement, value)
    # Remove leading/trailing dots and spaces
    sanitized = sanitized.strip('. ')
    # Collapse multiple replacements
    if replacement:
        sanitized = re.sub(re.escape(replacement) + '+', replacement, sanitized)
    # Truncate
    if len(sanitized) > max_length:
        # Keep extension if present
        dot_idx = sanitized.rfind('.')
        if dot_idx > 0:
            ext = sanitized[dot_idx:]
            base = sanitized[:dot_idx]
            sanitized = base[:max_length - len(ext)] + ext
        else:
            sanitized = sanitized[:max_length]
    # Handle reserved names (Windows)
    name_no_ext = sanitized.split('.')[0].upper()
    if name_no_ext in RESERVED_NAMES:
        sanitized = replacement + sanitized

    if not sanitized:
        sanitized = "unnamed"

    return {"ok": True, "value": value, "result": sanitized, "changed": sanitized != value}
