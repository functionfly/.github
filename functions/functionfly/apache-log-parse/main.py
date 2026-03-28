import re


# Apache combined log format (same as nginx combined)
COMBINED_RE = re.compile(
    r'^(?P<remote_addr>\S+)\s+'
    r'(?P<ident>\S+)\s+'
    r'(?P<auth_user>\S+)\s+'
    r'\[(?P<timestamp>[^\]]+)\]\s+'
    r'"(?P<method>\S+)\s+(?P<path>\S+)\s+(?P<protocol>[^"]+)"\s+'
    r'(?P<status>\d+)\s+'
    r'(?P<bytes_sent>\d+|-)'
    r'(?:\s+"(?P<referer>[^"]*)"\s+"(?P<user_agent>[^"]*)")?'
)


def handler(event):
    """Parse an Apache access log line."""
    try:
        line = event.get("line")
        if not line:
            return {"ok": False, "error": "line is required"}

        match = COMBINED_RE.match(line.strip())
        if not match:
            return {"ok": False, "error": "Line does not match Apache combined log format"}

        d = match.groupdict()
        return {
            "ok": True,
            "remote_addr": d["remote_addr"],
            "ident": d["ident"] if d["ident"] != "-" else None,
            "auth_user": d["auth_user"] if d["auth_user"] != "-" else None,
            "timestamp": d["timestamp"],
            "method": d["method"],
            "path": d["path"],
            "protocol": d["protocol"],
            "status": int(d["status"]),
            "bytes_sent": int(d["bytes_sent"]) if d["bytes_sent"] != "-" else 0,
            "referer": d.get("referer") if d.get("referer") != "-" else None,
            "user_agent": d.get("user_agent"),
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
