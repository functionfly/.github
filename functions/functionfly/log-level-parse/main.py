import re


LEVEL_PATTERNS = [
    (r'\b(TRACE|trace)\b', "trace", 0),
    (r'\b(DEBUG|debug|DBG|dbg)\b', "debug", 1),
    (r'\b(INFO|info|INFORMATION|information)\b', "info", 2),
    (r'\b(WARN|warn|WARNING|warning)\b', "warn", 3),
    (r'\b(ERROR|error|ERR|err)\b', "error", 4),
    (r'\b(FATAL|fatal|CRITICAL|critical|CRIT|crit)\b', "fatal", 5),
]


def handler(event):
    """Extract and normalize the log level from a log line."""
    try:
        line = event.get("line")
        if not line:
            return {"ok": False, "error": "line is required"}

        for pattern, normalized, severity in LEVEL_PATTERNS:
            match = re.search(pattern, line)
            if match:
                return {
                    "ok": True,
                    "level": match.group(1),
                    "normalized": normalized,
                    "severity": severity,
                }

        return {"ok": True, "level": None, "normalized": "unknown", "severity": -1}
    except Exception as e:
        return {"ok": False, "error": str(e)}
