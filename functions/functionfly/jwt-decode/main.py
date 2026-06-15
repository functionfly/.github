import base64
import json


def handler(event):
    if isinstance(event, dict):
        token = event.get("token", "")
    else:
        token = ""

    if not token:
        return {"ok": False, "error": "token is required"}

    if not isinstance(token, str):
        return {"ok": False, "error": "token must be a string"}

    parts = token.strip().split(".")
    if len(parts) != 3:
        return {"ok": False, "error": "invalid JWT format: expected 3 parts separated by dots"}

    try:
        def decode_base64url(data):
            data = data.replace("-", "+").replace("_", "/")
            padding = len(data) % 4
            if padding:
                data += "=" * (4 - padding)
            return base64.b64decode(data)

        header_bytes = decode_base64url(parts[0])
        header = json.loads(header_bytes.decode("utf-8"))

        payload_bytes = decode_base64url(parts[1])
        payload = json.loads(payload_bytes.decode("utf-8"))

        signature = parts[2]

        result = {
            "header": header,
            "payload": payload,
            "signature": signature
        }
        return {"ok": True, "result": result}
    except json.JSONDecodeError as e:
        return {"ok": False, "error": f"failed to decode JWT payload: {str(e)}"}
    except base64.binascii.Error as e:
        return {"ok": False, "error": f"invalid base64url encoding: {str(e)}"}
    except UnicodeDecodeError as e:
        return {"ok": False, "error": f"invalid UTF-8 in JWT: {str(e)}"}
    except Exception as e:
        return {"ok": False, "error": str(e)}
