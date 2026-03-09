import re


def _strip_tags_fallback(html, allowed=None):
    if not allowed:
        return re.sub(r"<[^>]+>", "", html or "")
    allowed_set = set((allowed or []) if isinstance(allowed, (list, tuple)) else [])
    pattern = re.compile(r"</?([a-zA-Z][a-zA-Z0-9]*)\b[^>]*>")
    def repl(m):
        tag = m.group(1).lower()
        if tag in allowed_set:
            return m.group(0)
        return ""
    return pattern.sub(repl, html or "")


def handler(event):
    if isinstance(event, dict):
        html = event.get("html", event.get("content", ""))
        allowed_tags = event.get("allowed_tags", event.get("tags"))
        allowed_attrs = event.get("allowed_attrs")
    else:
        html, allowed_tags, allowed_attrs = "", None, None

    if html is None:
        return {"ok": False, "error": "Input 'html' is required"}

    try:
        import bleach
        tags = allowed_tags if isinstance(allowed_tags, (list, tuple)) else ["b", "i", "em", "strong", "a", "p", "br", "ul", "ol", "li"]
        attrs = allowed_attrs if isinstance(allowed_attrs, dict) else {"a": ["href", "title"]}
        clean = bleach.clean(html or "", tags=tags, attributes=attrs, strip=True)
        return {"ok": True, "sanitized": clean}
    except ImportError:
        clean = _strip_tags_fallback(html, allowed_tags)
        return {"ok": True, "sanitized": clean}
    except Exception as e:
        return {"ok": False, "error": str(e)}
