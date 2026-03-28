import re


def parse_version(v):
    v = str(v).strip().lstrip("v")
    match = re.match(r'^(\d+)\.(\d+)\.(\d+)', v)
    if not match:
        return None
    return int(match.group(1)), int(match.group(2)), int(match.group(3))


def version_gte(a, b):
    return a >= b


def version_lte(a, b):
    return a <= b


def version_gt(a, b):
    return a > b


def version_lt(a, b):
    return a < b


def handler(event):
    """Check if a version satisfies a range."""
    try:
        version = event.get("version")
        range_str = event.get("range")
        if not version or not range_str:
            return {"ok": False, "error": "version and range are required"}

        v = parse_version(version)
        if not v:
            return {"ok": False, "error": f"Invalid version: {version}"}

        range_str = range_str.strip()

        # Caret range
        if range_str.startswith("^"):
            rv = parse_version(range_str[1:])
            if not rv:
                return {"ok": False, "error": f"Invalid range: {range_str}"}
            major = rv[0]
            if major > 0:
                satisfies = v >= rv and v < (major + 1, 0, 0)
            elif rv[1] > 0:
                satisfies = v >= rv and v < (0, rv[1] + 1, 0)
            else:
                satisfies = v >= rv and v < (0, 0, rv[2] + 1)

        # Tilde range
        elif range_str.startswith("~"):
            rv = parse_version(range_str[1:])
            if not rv:
                return {"ok": False, "error": f"Invalid range: {range_str}"}
            satisfies = v >= rv and v < (rv[0], rv[1] + 1, 0)

        # Comparison operators
        elif range_str.startswith((">=", "<=", ">", "<", "=")):
            match = re.match(r'^(>=|<=|>|<|=)\s*(.+)$', range_str)
            if not match:
                return {"ok": False, "error": f"Invalid range: {range_str}"}
            op, rv_str = match.group(1), match.group(2)
            rv = parse_version(rv_str)
            if not rv:
                return {"ok": False, "error": f"Invalid version in range: {rv_str}"}
            ops = {">=": version_gte, "<=": version_lte, ">": version_gt, "<": version_lt, "=": lambda a, b: a == b}
            satisfies = ops[op](v, rv)

        # Exact version
        else:
            rv = parse_version(range_str)
            satisfies = v == rv if rv else False

        return {"ok": True, "result": satisfies, "satisfies": satisfies}
    except Exception as e:
        return {"ok": False, "error": str(e)}
