# Common extension -> MIME type map (subset of mimetypes)
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


def handler(event):
    """
    Detect MIME type from filename (extension) or optional content sniff.

    Input:
        - filename: Filename or path (used for extension)
        - content: Optional base64 or raw bytes for content sniff (not implemented; extension only)

    Returns:
        - ok: True on success
        - mime_type: Detected MIME type (e.g. image/png)
        - extension: Extension used
        - error: Message if ok is False
    """
    if isinstance(event, dict):
        filename = event.get("filename", event.get("path", event.get("name", "")))
    else:
        filename = str(event) if event else ""

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

    if not ext:
        return {"ok": True, "mime_type": "application/octet-stream", "extension": None}

    mime = EXTENSION_MAP.get(ext, "application/octet-stream")
    return {"ok": True, "mime_type": mime, "extension": ext}
