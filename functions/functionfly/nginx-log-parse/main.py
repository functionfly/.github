import re


# Nginx combined log format
COMBINED_RE = re.compile(
    r'^(?P<remote_addr>\S+)\s+'
    r'(?P<remote_user>\S+)\s+'
    r'(?P<auth_user>\S+)\s+'
    r'\[(?P<timestamp>[^\]]+)\]\s+'
    r'"(?P<method>\S+)\s+(?P<path>\S+)\s+(?P<protocol>[^"]+)"\s+'
    r'(?P<status>\d+)\s+'
    r'(?P<bytes_sent>\d+|-)\s+'
    r'"(?P<referer>[^"]*)"\s+'
    r'"(?P<user_agent>[^"]*)"'
)


def handler(event):
    """Parse an Nginx access log line."""
    try:
        line = event.get("line")
        if not line:
            return {"ok": False, "error": "line is required"}

        match = COMBINED_RE.match(line.strip())
        if not match:
            return {"ok": False, "error": "Line does not match Nginx combined log format"}

        d = match.groupdict()
        return {
            "ok": True,
            "remote_addr": d["remote_addr"],
            "remote_user": d["remote_user"] if d["remote_user"] != "-" else None,
            "timestamp": d["timestamp"],
            "method": d["method"],
            "path": d["path"],
            "protocol": d["protocol"],
            "status": int(d["status"]),
            "bytes_sent": int(d["bytes_sent"]) if d["bytes_sent"] != "-" else 0,
            "referer": d["referer"] if d["referer"] != "-" else None,
            "user_agent": d["user_agent"],
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
