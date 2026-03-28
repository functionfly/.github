import re


DEFAULT_TYPES = ["feat", "fix", "docs", "style", "refactor", "perf", "test", "build", "ci", "chore", "revert"]


def handler(event):
    """Validate a commit message against Conventional Commits spec."""
    try:
        message = event.get("message")
        if not message:
            return {"ok": False, "error": "message is required"}

        allowed_types = event.get("allowed_types", DEFAULT_TYPES)
        errors = []
        lines = message.strip().split("\n")
        header = lines[0].strip()

        # Check header length
        if len(header) > 100:
            errors.append(f"Header too long ({len(header)} chars, max 100)")

        # Parse header
        match = re.match(
            r'^(?P<type>[a-z]+)(?:\((?P<scope>[^)]+)\))?(?P<breaking>!)?:\s*(?P<subject>.+)$',
            header
        )

        if not match:
            errors.append("Header must follow format: type(scope): subject")
            return {"ok": True, "valid": False, "errors": errors, "parsed": None}

        ctype = match.group("type")
        scope = match.group("scope")
        subject = match.group("subject").strip()
        breaking = bool(match.group("breaking"))

        if ctype not in allowed_types:
            errors.append(f"Unknown type '{ctype}'. Allowed: {', '.join(allowed_types)}")

        if not subject:
            errors.append("Subject cannot be empty")
        elif subject[0].isupper():
            errors.append("Subject should not start with uppercase")
        elif subject.endswith("."):
            errors.append("Subject should not end with a period")

        parsed = {"type": ctype, "scope": scope, "subject": subject, "breaking": breaking}
        return {"ok": True, "valid": len(errors) == 0, "errors": errors, "parsed": parsed}
    except Exception as e:
        return {"ok": False, "error": str(e)}
