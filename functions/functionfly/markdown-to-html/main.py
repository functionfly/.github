import re


def handler(event):
    """
    Convert Markdown to HTML (basic conversion using stdlib).

    Input:
        - markdown: Markdown string to convert (required)

    Returns:
        - ok: True on success
        - html: Converted HTML string
        - error: Message if ok is False
    """
    if isinstance(event, dict):
        md = event.get("markdown", event.get("data", ""))
    else:
        md = str(event) if event is not None else ""

    if md is None or (isinstance(md, str) and not md.strip()):
        return {"ok": False, "error": "Input 'markdown' is required"}

    text = str(md).strip()
    lines = text.split("\n")
    out = []
    i = 0

    while i < len(lines):
        line = lines[i]
        stripped = line.strip()

        # Headers # ## ### etc
        if stripped.startswith("#"):
            level = 0
            while level < len(stripped) and stripped[level] == "#":
                level += 1
            content = stripped[level:].strip()
            content = _escape_html(content)
            out.append(f"<h{level}>{content}</h{level}>")
            i += 1
            continue

        # Bold **text**
        line = re.sub(r"\*\*([^*]+)\*\*", r"<strong>\1</strong>", line)
        line = re.sub(r"__([^_]+)__", r"<strong>\1</strong>", line)
        # Italic *text*
        line = re.sub(r"\*([^*]+)\*", r"<em>\1</em>", line)
        line = re.sub(r"_([^_]+)_", r"<em>\1</em>", line)
        # Links [text](url)
        line = re.sub(
            r"\[([^\]]+)\]\(([^)]+)\)",
            r'<a href="\2">\1</a>',
            line,
        )
        line = _escape_html(line)

        if stripped.startswith("- ") or stripped.startswith("* "):
            if out and not out[-1].startswith("<ul>"):
                out.append("<ul>")
            content = stripped[2:].strip()
            out.append(f"<li>{content}</li>")
            i += 1
            continue

        if out and out[-1] == "<ul>":
            out.append("</ul>")
        if stripped:
            out.append(f"<p>{line}</p>")
        else:
            out.append("<br/>")
        i += 1

    if out and out[-1] == "<ul>":
        out.append("</ul>")

    html = "\n".join(out)
    return {"ok": True, "html": html}


def _escape_html(s):
    s = s.replace("&", "&amp;")
    s = s.replace("<", "&lt;")
    s = s.replace(">", "&gt;")
    s = s.replace('"', "&quot;")
    return s
