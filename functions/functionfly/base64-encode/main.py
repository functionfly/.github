import base64
import json


def handler(event):
    if isinstance(event, dict):
        data = event.get("data", "")
        encoding = event.get("encoding", "utf-8")
    else:
        data = ""
        encoding = "utf-8"

    if not data:
        return {"ok": False, "error": "data is required"}

    if not isinstance(data, str):
        return {"ok": False, "error": "data must be a string"}

    valid_encodings = {"utf-8", "utf-16", "utf-32", "ascii", "latin-1", "iso-8859-1"}
    encoding_lower = encoding.lower()
    if encoding_lower not in valid_encodings:
        return {"ok": False, "error": f"unsupported encoding: {encoding}. Supported: {', '.join(sorted(valid_encodings))}"}

    try:
        data_bytes = data.encode(encoding_lower)
        encoded_bytes = base64.b64encode(data_bytes)
        encoded_str = encoded_bytes.decode("ascii")
        return {"ok": True, "result": encoded_str}
    except UnicodeEncodeError as e:
        return {"ok": False, "error": f"data contains characters not encodable in {encoding}: {str(e)}"}
    except Exception as e:
        return {"ok": False, "error": str(e)}
