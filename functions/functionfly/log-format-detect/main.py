import re
import json


def handler(event):
    """Detect the format of a log line."""
    try:
        line = event.get("line")
        if not line:
            return {"ok": False, "error": "line is required"}

        line = line.strip()

        # JSON log
        if line.startswith("{") and line.endswith("}"):
            try:
                json.loads(line)
                return {"ok": True, "format": "json", "confidence": 1.0}
            except json.JSONDecodeError:
                pass

        # Syslog format: <priority>timestamp hostname process[pid]: message
        if re.match(r'^<\d+>', line) or re.match(r'^\w{3}\s+\d+\s+\d{2}:\d{2}:\d{2}\s+\S+\s+\S+:', line):
            return {"ok": True, "format": "syslog", "confidence": 0.9}

        # Nginx access log: IP - - [date] "method path proto" status size
        if re.match(r'^\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}\s+-\s+-\s+\[', line):
            return {"ok": True, "format": "nginx_access", "confidence": 0.95}

        # Apache access log (similar to nginx)
        if re.match(r'^\S+\s+\S+\s+\S+\s+\[\d{2}/\w+/\d{4}', line):
            return {"ok": True, "format": "apache_access", "confidence": 0.9}

        # CloudWatch / AWS log: timestamp level message
        if re.match(r'^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}', line):
            return {"ok": True, "format": "iso8601", "confidence": 0.8}

        # Log4j / Java: timestamp level class - message
        if re.match(r'^\d{4}-\d{2}-\d{2}\s+\d{2}:\d{2}:\d{2}', line):
            return {"ok": True, "format": "log4j", "confidence": 0.75}

        # Logfmt: key=value pairs
        if re.search(r'\w+=(?:"[^"]*"|\S+)', line) and "=" in line:
            return {"ok": True, "format": "logfmt", "confidence": 0.7}

        return {"ok": True, "format": "unknown", "confidence": 0.0}
    except Exception as e:
        return {"ok": False, "error": str(e)}
