import hashlib


def handler(event):
    content = event.get("content") if isinstance(event, dict) else None
    weak = event.get("weak", False)

    if content is None:
        return {"ok": False, "error": "content is required"}

    if isinstance(content, dict) or isinstance(content, list):
        import json
        content_str = json.dumps(content, sort_keys=True, separators=(',', ':'))
    else:
        content_str = str(content)

    digest = hashlib.md5(content_str.encode('utf-8')).hexdigest()

    if weak:
        etag = f'W/"{digest}"'
    else:
        etag = f'"{digest}"'

    return {"ok": True, "etag": etag, "weak": weak}
