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
        data_bytes = data.encode("ascii")
    except UnicodeEncodeError:
        return {"ok": False, "error": "data must be ASCII-safe base64 string"}

    try:
        decoded_bytes = base64.b64decode(data_bytes, validate=True)
        decoded_str = decoded_bytes.decode(encoding_lower)
        return {"ok": True, "result": decoded_str}
    except base64.binascii.Error as e:
        return {"ok": False, "error": f"invalid base64 data: {str(e)}"}
    except UnicodeDecodeError as e:
        return {"ok": False, "error": f"failed to decode data with {encoding} encoding: {str(e)}"}
    except Exception as e:
        return {"ok": False, "error": str(e)}
