import re


def parse_version(v):
    v = str(v).strip().lstrip("v")
    match = re.match(r'^(\d+)\.(\d+)\.(\d+)(?:-([a-zA-Z0-9\-\.]+))?', v)
    if not match:
        raise ValueError(f"Invalid version: {v}")
    major, minor, patch = int(match.group(1)), int(match.group(2)), int(match.group(3))
    pre = match.group(4)
    return (major, minor, patch, pre)


def compare_pre(a, b):
    if a is None and b is None:
        return 0
    if a is None:
        return 1  # release > prerelease
    if b is None:
        return -1
    a_parts = a.split(".")
    b_parts = b.split(".")
    for ap, bp in zip(a_parts, b_parts):
        if ap.isdigit() and bp.isdigit():
            if int(ap) != int(bp):
                return 1 if int(ap) > int(bp) else -1
        else:
            if ap != bp:
                return 1 if ap > bp else -1
    if len(a_parts) != len(b_parts):
        return 1 if len(a_parts) > len(b_parts) else -1
    return 0


def handler(event):
    """Compare two semantic version strings."""
    try:
        va = event.get("version_a")
        vb = event.get("version_b")
        if not va or not vb:
            return {"ok": False, "error": "version_a and version_b are required"}

        a = parse_version(va)
        b = parse_version(vb)

        for i in range(3):
            if a[i] != b[i]:
                result = 1 if a[i] > b[i] else -1
                op = ">" if result == 1 else "<"
                return {"ok": True, "result": result, "comparison": f"{va} {op} {vb}"}

        result = compare_pre(a[3], b[3])
        if result == 0:
            return {"ok": True, "result": 0, "comparison": f"{va} == {vb}"}
        op = ">" if result == 1 else "<"
        return {"ok": True, "result": result, "comparison": f"{va} {op} {vb}"}
    except Exception as e:
        return {"ok": False, "error": str(e)}
