import mimetypes

MAGIC_SIGNATURES = [
    (b"\x89PNG\r\n\x1a\n", "image/png"),
    (b"\xff\xd8\xff", "image/jpeg"),
    (b"GIF87a", "image/gif"),
    (b"GIF89a", "image/gif"),
    (b"RIFF", "audio/wav"),
    (b"\x1f\x8b", "application/gzip"),
    (b"%PDF", "application/pdf"),
    (b"PK\x03\x04", "application/zip"),
    (b"\x00\x00\x01\x00", "image/x-icon"),
    (b"\x00\x00\x02\x00", "image/x-icon"),
    (b"\x00\x00\x00", "image/webp"),
    (b"fLaC", "audio/flac"),
    (b"ID3", "audio/mpeg"),
    (b"\xff\xfb", "audio/mpeg"),
    (b"\xff\xf3", "audio/mpeg"),
    (b"\xff\xf2", "audio/mpeg"),
    (b"RIFF", "video/webm"),
    (b"\x1aE\xdf\xa3", "video/webm"),
    (b"\x00\x00\x00", "video/mp4"),
    (b"ftyp", "video/mp4"),
    (b"ftyp", "video/quicktime"),
    (b"<svg", "image/svg+xml"),
    (b"<?xml", "application/xml"),
    (b"<!DOCTYPE html", "text/html"),
    (b"<html", "text/html"),
    (b"\xef\xbb\xbf", "text/csv"),
    (b"\xfe\xff", "text/csv"),
    (b"PK\x03\x04", "application/vnd.openxmlformats-officedocument.wordprocessingml.document"),
]

EXTENSION_MAP = {
    ".txt": "text/plain",
    ".html": "text/html",
    ".htm": "text/html",
    ".css": "text/css",
    ".js": "application/javascript",
    ".json": "application/json",
    ".xml": "application/xml",
    ".pdf": "application/pdf",
    ".zip": "application/zip",
    ".gz": "application/gzip",
    ".tar": "application/x-tar",
    ".png": "image/png",
    ".jpg": "image/jpeg",
    ".jpeg": "image/jpeg",
    ".gif": "image/gif",
    ".webp": "image/webp",
    ".svg": "image/svg+xml",
    ".ico": "image/x-icon",
    ".mp3": "audio/mpeg",
    ".wav": "audio/wav",
    ".mp4": "video/mp4",
    ".webm": "video/webm",
    ".woff": "font/woff",
    ".woff2": "font/woff2",
    ".ttf": "font/ttf",
    ".csv": "text/csv",
    ".yaml": "text/yaml",
    ".yml": "text/yaml",
    ".md": "text/markdown",
    ".rss": "application/rss+xml",
}


def _sniff_content(content_bytes, ext):
    for magic, mime in MAGIC_SIGNATURES:
        if content_bytes.startswith(magic):
            return mime
    if content_bytes.startswith(b"<?xml"):
        return "application/xml"
    if content_bytes.startswith(b"<svg"):
        return "image/svg+xml"
    if content_bytes.startswith(b"<!DOCTYPE html") or content_bytes.startswith(b"<html"):
        return "text/html"
    if ext:
        detected = mimetypes.guess_type("file." + ext, strict=False)[0]
        if detected:
            return detected
    return None


def handler(event):
    """
    Detect MIME type from filename (extension) or content sniff.

    Input:
        - filename: Filename or path (used for extension)
        - content: Optional base64-encoded content for sniffing

    Returns:
        - ok: True on success
        - mime_type: Detected MIME type (e.g. image/png)
        - extension: Extension used
        - source: 'content' or 'extension' indicating what was used
        - error: Message if ok is False
    """
    if isinstance(event, dict):
        filename = event.get("filename", event.get("path", event.get("name", "")))
        content_b64 = event.get("content")
    else:
        filename = str(event) if event else ""
        content_b64 = None

    if not filename or not str(filename).strip():
        return {"ok": False, "error": "Input 'filename' is required"}

    name = str(filename).strip()
    ext = ""
    for i in range(len(name) - 1, -1, -1):
        if name[i] == ".":
            ext = name[i:].lower()
            break
        if name[i] in "/\\":
            break

    mime = None
    source = "extension"

    if content_b64:
        import base64
        try:
            content_bytes = base64.b64decode(content_b64)
            mime = _sniff_content(content_bytes, ext)
            if mime:
                source = "content"
        except Exception:
            pass

    if not mime:
        if not ext:
            return {"ok": True, "mime_type": "application/octet-stream", "extension": None, "source": "extension"}
        mime = EXTENSION_MAP.get(ext, "application/octet-stream")

    return {"ok": True, "mime_type": mime, "extension": ext, "source": source}
