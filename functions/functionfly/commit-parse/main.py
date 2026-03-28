import re


CONVENTIONAL_RE = re.compile(
    r'^(?P<type>feat|fix|docs|style|refactor|perf|test|build|ci|chore|revert)'
    r'(?:\((?P<scope>[^)]+)\))?'
    r'(?P<breaking>!)?'
    r':\s*(?P<subject>.+)$'
)


def handler(event):
    """Parse a git commit message into structured components."""
    try:
        message = event.get("message")
        if not message:
            return {"ok": False, "error": "message is required"}

        lines = message.strip().split("\n")
        header = lines[0].strip()

        # Try conventional commit format
        match = CONVENTIONAL_RE.match(header)
        if match:
            body_lines = []
            footer_lines = []
            in_footer = False

            for line in lines[2:]:
                if re.match(r'^[A-Z][A-Z-]+:', line) or line.startswith("BREAKING CHANGE:"):
                    in_footer = True
                if in_footer:
                    footer_lines.append(line)
                else:
                    body_lines.append(line)

            body = "\n".join(body_lines).strip()
            footer = "\n".join(footer_lines).strip()
            breaking = bool(match.group("breaking")) or "BREAKING CHANGE" in footer

            return {
                "ok": True,
                "conventional": True,
                "result": {
                    "type": match.group("type"),
                    "scope": match.group("scope"),
                    "subject": match.group("subject").strip(),
                    "body": body or None,
                    "footer": footer or None,
                    "breaking": breaking,
                    "header": header,
                },
            }
        else:
            # Non-conventional commit
            body = "\n".join(lines[2:]).strip() if len(lines) > 2 else None
            return {
                "ok": True,
                "conventional": False,
                "result": {
                    "type": None,
                    "scope": None,
                    "subject": header,
                    "body": body,
                    "footer": None,
                    "breaking": False,
                    "header": header,
                },
            }
    except Exception as e:
        return {"ok": False, "error": str(e)}
