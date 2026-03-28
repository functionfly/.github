import json


TIMESTAMP_FIELDS = ["time", "timestamp", "ts", "@timestamp", "datetime", "date"]
LEVEL_FIELDS = ["level", "severity", "log_level", "loglevel", "lvl"]
MESSAGE_FIELDS = ["msg", "message", "text", "log", "body"]


def handler(event):
    """Parse a JSON-formatted log line."""
    try:
        line = event.get("line")
        if not line:
            return {"ok": False, "error": "line is required"}

        field_map = event.get("field_map", {})

        try:
            data = json.loads(line.strip())
        except json.JSONDecodeError as e:
            return {"ok": False, "error": f"Invalid JSON: {e}"}

        # Extract standard fields
        timestamp = None
        for f in [field_map.get("timestamp")] + TIMESTAMP_FIELDS:
            if f and f in data:
                timestamp = data[f]
                break

        level = None
        for f in [field_map.get("level")] + LEVEL_FIELDS:
            if f and f in data:
                level = str(data[f]).lower()
                break

        message = None
        for f in [field_map.get("message")] + MESSAGE_FIELDS:
            if f and f in data:
                message = data[f]
                break

        # Remaining fields
        known_fields = set(TIMESTAMP_FIELDS + LEVEL_FIELDS + MESSAGE_FIELDS + list(field_map.values()))
        fields = {k: v for k, v in data.items() if k not in known_fields}

        return {
            "ok": True,
            "timestamp": timestamp,
            "level": level,
            "message": message,
            "fields": fields,
            "raw": data,
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
