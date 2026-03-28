import re


def handler(event):
    """Extract error codes from log text."""
    try:
        text = event.get("text")
        if not text:
            return {"ok": False, "error": "text is required"}

        # HTTP status codes
        http_codes = [int(m) for m in re.findall(r'\b([1-5]\d{2})\b', text)]
        http_codes = list(set(http_codes))

        # Named error codes (E_*, ERR_*, ENOENT, etc.)
        error_codes = re.findall(r'\b(E[A-Z_]{2,}|ERR_[A-Z_]+|ENOENT|EACCES|ECONNREFUSED|ETIMEDOUT)\b', text)
        error_codes = list(set(error_codes))

        # errno values
        errno_matches = re.findall(r'errno\s+(\d+)', text, re.IGNORECASE)

        # Exit codes
        exit_codes = re.findall(r'exit\s+(?:code\s+)?(\d+)', text, re.IGNORECASE)

        all_codes = list(set(http_codes + error_codes + errno_matches + exit_codes))

        return {
            "ok": True,
            "codes": all_codes,
            "http_codes": http_codes,
            "error_codes": error_codes,
            "errno_values": errno_matches,
            "exit_codes": exit_codes,
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
