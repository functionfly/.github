import xml.etree.ElementTree as ET


def handler(event):
    value = event.get("value") if isinstance(event, dict) else None
    if value is None:
        return {"ok": False, "error": "value is required"}
    try:
        root = ET.fromstring(str(value))
        return {"ok": True, "value": value, "result": True, "root_tag": root.tag}
    except ET.ParseError as e:
        return {"ok": True, "value": value, "result": False, "parse_error": str(e)}
    except Exception as e:
        return {"ok": True, "value": value, "result": False, "parse_error": str(e)}
