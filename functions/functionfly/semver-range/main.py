import re


def parse_version(v):
    v = str(v).strip()
    match = re.match(r'^(\d+)(?:\.(\d+))?(?:\.(\d+))?', v)
    if not match:
        return None
    major = int(match.group(1))
    minor = int(match.group(2)) if match.group(2) else 0
    patch = int(match.group(3)) if match.group(3) else 0
    return major, minor, patch


def handler(event):
    """Parse a semantic version range expression."""
    try:
        range_str = event.get("range")
        if not range_str:
            return {"ok": False, "error": "range is required"}

        range_str = range_str.strip()

        # Caret range: ^1.2.3 -> >=1.2.3 <2.0.0
        if range_str.startswith("^"):
            v = range_str[1:].strip()
            parts = parse_version(v)
            if not parts:
                return {"ok": False, "error": f"Invalid version in range: {v}"}
            major, minor, patch = parts
            if major > 0:
                max_v = f"{major + 1}.0.0"
            elif minor > 0:
                max_v = f"0.{minor + 1}.0"
            else:
                max_v = f"0.0.{patch + 1}"
            return {"ok": True, "result": {"operator": "^", "version": v, "min": v, "max": max_v, "description": f"Compatible with {v} (>={v} <{max_v})"}}

        # Tilde range: ~1.2.3 -> >=1.2.3 <1.3.0
        elif range_str.startswith("~"):
            v = range_str[1:].strip()
            parts = parse_version(v)
            if not parts:
                return {"ok": False, "error": f"Invalid version in range: {v}"}
            major, minor, patch = parts
            max_v = f"{major}.{minor + 1}.0"
            return {"ok": True, "result": {"operator": "~", "version": v, "min": v, "max": max_v, "description": f"Approximately {v} (>={v} <{max_v})"}}

        # Comparison operators
        elif range_str.startswith((">=", "<=", ">", "<", "=")):
            match = re.match(r'^(>=|<=|>|<|=)\s*(.+)$', range_str)
            if match:
                op, v = match.group(1), match.group(2).strip()
                return {"ok": True, "result": {"operator": op, "version": v, "description": f"{op}{v}"}}

        # Exact version
        elif re.match(r'^\d', range_str):
            return {"ok": True, "result": {"operator": "=", "version": range_str, "description": f"Exactly {range_str}"}}

        return {"ok": True, "result": {"operator": "unknown", "version": range_str, "description": range_str}}
    except Exception as e:
        return {"ok": False, "error": str(e)}
