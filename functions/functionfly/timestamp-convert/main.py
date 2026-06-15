from datetime import datetime, timezone
from typing import Any


def handler(event: dict[str, Any]) -> dict[str, Any]:
    try:
        timestamp = event.get("timestamp")
        from_zone = event.get("from_zone", "UTC")
        to_zone = event.get("to_zone", "UTC")
        output_format = event.get("output_format", "iso").lower().strip()
        
        if timestamp is None:
            return {"ok": False, "error": "timestamp is required (unix integer or ISO string)"}
        
        valid_formats = ["iso", "unix", "human"]
        if output_format not in valid_formats:
            return {"ok": False, "error": f"output_format must be one of: {', '.join(valid_formats)}"}
        
        try:
            from zoneinfo import ZoneInfo
        except ImportError:
            return {"ok": False, "error": "zoneinfo module not available (Python 3.9+ required)"}
        
        try:
            from_tz = ZoneInfo(from_zone)
        except Exception:
            return {"ok": False, "error": f"Invalid from_zone: {from_zone}"}
        
        try:
            to_tz = ZoneInfo(to_zone)
        except Exception:
            return {"ok": False, "error": f"Invalid to_zone: {to_zone}"}
        
        if isinstance(timestamp, (int, float)):
            dt = datetime.fromtimestamp(timestamp, tz=from_tz)
        elif isinstance(timestamp, str):
            timestamp_str = timestamp.strip()
            
            if timestamp_str.isdigit():
                dt = datetime.fromtimestamp(int(timestamp_str), tz=from_tz)
            else:
                try:
                    dt = datetime.fromisoformat(timestamp_str.replace('Z', '+00:00'))
                except ValueError:
                    try:
                        dt = datetime.strptime(timestamp_str, "%Y-%m-%d %H:%M:%S")
                        dt = dt.replace(tzinfo=from_tz)
                    except ValueError:
                        try:
                            dt = datetime.strptime(timestamp_str, "%Y-%m-%d")
                            dt = dt.replace(tzinfo=from_tz)
                        except ValueError:
                            return {"ok": False, "error": f"Could not parse timestamp: {timestamp}"}
        else:
            return {"ok": False, "error": "timestamp must be a unix timestamp (int) or ISO string"}
        
        converted_dt = dt.astimezone(to_tz)
        
        unix_timestamp = int(converted_dt.timestamp())
        
        if output_format == "iso":
            converted_timestamp = converted_dt.isoformat()
        elif output_format == "unix":
            converted_timestamp = str(unix_timestamp)
        else:
            converted_timestamp = converted_dt.strftime("%Y-%m-%d %H:%M:%S %Z")
        
        return {
            "ok": True,
            "converted_timestamp": converted_timestamp,
            "from_zone": from_zone,
            "to_zone": to_zone,
            "unix_timestamp": unix_timestamp,
            "output_format": output_format
        }
        
    except Exception as e:
        return {"ok": False, "error": str(e)}
