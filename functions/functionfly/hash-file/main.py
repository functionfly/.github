import base64
import hashlib


def handler(event):
    if isinstance(event, dict):
        content = event.get("content")
        content_b64 = event.get("content_base64")
        algorithm = (event.get("algorithm") or "sha256").lower().replace("-", "")
    else:
        content, content_b64, algorithm = None, None, "sha256"

    if content is None and content_b64 is None:
        return {"ok": False, "error": "Provide 'content' or 'content_base64'"}

    if content_b64:
        try:
            data = base64.b64decode(content_b64)
        except Exception as e:
            return {"ok": False, "error": f"Invalid base64: {e}"}
    else:
        data = content.encode("utf-8") if isinstance(content, str) else content
    if not isinstance(data, bytes):
        data = bytes(data)

    alg_map = {"sha256": hashlib.sha256, "sha512": hashlib.sha512, "sha1": hashlib.sha1, "md5": hashlib.md5}
    hasher = alg_map.get(algorithm, hashlib.sha256)
    h = hasher(data).hexdigest()
    return {"ok": True, "hash": h, "algorithm": algorithm}
