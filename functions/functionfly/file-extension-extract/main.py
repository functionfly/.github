import os


def handler(event):
    """
    Extract file extension from a filename or path.

    Input:
        - filename: Filename or path (required)
        - include_dot: If true, include the leading dot (default: true)

    Returns:
        - ok: True on success
        - extension: Extracted extension (e.g. .png or png)
        - basename: Filename without path
        - error: Message if ok is False
    """
    if isinstance(event, dict):
        filename = event.get("filename", event.get("path", event.get("name", "")))
        include_dot = event.get("include_dot", True)
    else:
        filename = str(event) if event else ""
        include_dot = True

    if not filename or not str(filename).strip():
        return {"ok": False, "error": "Input 'filename' is required"}

    name = os.path.basename(str(filename).strip())
    _, ext = os.path.splitext(name)
    if not include_dot and ext.startswith("."):
        ext = ext[1:]
    return {"ok": True, "extension": ext, "basename": name}
