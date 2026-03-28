def diff_objects(base, target, path=""):
    """Recursively diff two objects."""
    added = {}
    removed = {}
    changed = {}
    unchanged = {}

    all_keys = set(list(base.keys()) + list(target.keys()))

    for key in all_keys:
        full_key = f"{path}.{key}" if path else key
        if key not in base:
            added[full_key] = target[key]
        elif key not in target:
            removed[full_key] = base[key]
        elif isinstance(base[key], dict) and isinstance(target[key], dict):
            sub = diff_objects(base[key], target[key], full_key)
            added.update(sub["added"])
            removed.update(sub["removed"])
            changed.update(sub["changed"])
            unchanged.update(sub["unchanged"])
        elif base[key] != target[key]:
            changed[full_key] = {"from": base[key], "to": target[key]}
        else:
            unchanged[full_key] = base[key]

    return {"added": added, "removed": removed, "changed": changed, "unchanged": unchanged}


def handler(event):
    """Compare two configuration objects."""
    try:
        base = event.get("base")
        target = event.get("target")
        if base is None or target is None:
            return {"ok": False, "error": "base and target are required"}

        result = diff_objects(base, target)
        has_changes = bool(result["added"] or result["removed"] or result["changed"])

        return {"ok": True, "result": result, "has_changes": has_changes}
    except Exception as e:
        return {"ok": False, "error": str(e)}
