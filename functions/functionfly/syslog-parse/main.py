import re


SEVERITY_NAMES = ["emerg", "alert", "crit", "err", "warning", "notice", "info", "debug"]
FACILITY_NAMES = ["kern", "user", "mail", "daemon", "auth", "syslog", "lpr", "news", "uucp", "cron", "authpriv", "ftp"]


def handler(event):
    """Parse a syslog format log line."""
    try:
        line = event.get("line")
        if not line:
            return {"ok": False, "error": "line is required"}

        line = line.strip()
        result = {}

        # Parse priority: <PRI>
        pri_match = re.match(r'^<(\d+)>(.*)', line)
        if pri_match:
            pri = int(pri_match.group(1))
            result["facility"] = pri >> 3
            result["severity"] = pri & 7
            result["severity_name"] = SEVERITY_NAMES[result["severity"]] if result["severity"] < len(SEVERITY_NAMES) else "unknown"
            result["facility_name"] = FACILITY_NAMES[result["facility"]] if result["facility"] < len(FACILITY_NAMES) else "unknown"
            rest = pri_match.group(2).strip()
        else:
            rest = line

        # RFC 3164: MMM DD HH:MM:SS hostname process[pid]: message
        rfc3164 = re.match(
            r'^(\w{3}\s+\d+\s+\d{2}:\d{2}:\d{2})\s+(\S+)\s+(\S+?)(?:\[(\d+)\])?:\s*(.*)',
            rest
        )
        if rfc3164:
            result["timestamp"] = rfc3164.group(1)
            result["hostname"] = rfc3164.group(2)
            result["process"] = rfc3164.group(3)
            result["pid"] = int(rfc3164.group(4)) if rfc3164.group(4) else None
            result["message"] = rfc3164.group(5)
        else:
            result["message"] = rest

        return {"ok": True, **result}
    except Exception as e:
        return {"ok": False, "error": str(e)}
