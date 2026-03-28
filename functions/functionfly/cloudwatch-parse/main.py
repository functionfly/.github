import re
import json


def handler(event):
    """Parse an AWS CloudWatch log event."""
    try:
        cw_event = event.get("event")
        line = event.get("line")

        if cw_event:
            # CloudWatch event object
            timestamp = cw_event.get("timestamp")
            message = cw_event.get("message", "")
        elif line:
            message = line
            timestamp = None
        else:
            return {"ok": False, "error": "event or line is required"}

        result = {"timestamp": timestamp, "message": message}

        # Try to parse as JSON log
        if message.strip().startswith("{"):
            try:
                parsed = json.loads(message.strip())
                result["level"] = parsed.get("level") or parsed.get("severity")
                result["log_message"] = parsed.get("msg") or parsed.get("message")
                result["fields"] = parsed
                return {"ok": True, **result}
            except json.JSONDecodeError:
                pass

        # Lambda log format: TIMESTAMP\tLEVEL\tMESSAGE
        lambda_match = re.match(r'^(\S+)\t(\w+)\t(.*)', message, re.DOTALL)
        if lambda_match:
            result["timestamp"] = result["timestamp"] or lambda_match.group(1)
            result["level"] = lambda_match.group(2).lower()
            result["log_message"] = lambda_match.group(3)

        # Extract Lambda request ID
        req_match = re.search(r'RequestId:\s+([a-f0-9-]+)', message)
        if req_match:
            result["request_id"] = req_match.group(1)

        # Extract REPORT line metrics
        duration_match = re.search(r'Duration:\s+([\d.]+)\s+ms', message)
        if duration_match:
            result["duration_ms"] = float(duration_match.group(1))

        billed_match = re.search(r'Billed Duration:\s+(\d+)\s+ms', message)
        if billed_match:
            result["billed_duration_ms"] = int(billed_match.group(1))

        memory_match = re.search(r'Memory Size:\s+(\d+)\s+MB', message)
        if memory_match:
            result["memory_mb"] = int(memory_match.group(1))

        max_memory_match = re.search(r'Max Memory Used:\s+(\d+)\s+MB', message)
        if max_memory_match:
            result["max_memory_used_mb"] = int(max_memory_match.group(1))

        # Detect errors
        if re.search(r'\b(ERROR|Exception|Error|FATAL)\b', message):
            result["has_error"] = True

        return {"ok": True, **result}
    except Exception as e:
        return {"ok": False, "error": str(e)}
