import re


def handler(event):
    """Parse a .env file string into a key-value object."""
    try:
        content = event.get("content")
        if content is None:
            return {"ok": False, "error": "content is required"}

        result = {}
        for line in content.splitlines():
            line = line.strip()
            # Skip empty lines and comments
            if not line or line.startswith("#"):
                continue
            # Match KEY=VALUE or KEY="VALUE" or KEY='VALUE'
            match = re.match(r'^([A-Za-z_][A-Za-z0-9_]*)\s*=\s*(.*)', line)
            if not match:
                continue
            key = match.group(1)
            value = match.group(2).strip()
            # Remove inline comments (not inside quotes)
            if not (value.startswith('"') or value.startswith("'")):
                value = re.sub(r'\s+#.*$', '', value).strip()
            # Remove surrounding quotes
            if len(value) >= 2:
                if (value.startswith('"') and value.endswith('"')) or \
                   (value.startswith("'") and value.endswith("'")):
                    value = value[1:-1]
            result[key] = value

        return {"ok": True, "result": result, "count": len(result)}
    except Exception as e:
        return {"ok": False, "error": str(e)}
