import copy


def _merge_patch(target, patch):
    """RFC 7396 JSON Merge Patch"""
    if not isinstance(patch, dict):
        return copy.deepcopy(patch)
    result = copy.deepcopy(target) if isinstance(target, dict) else {}
    for key, value in patch.items():
        if value is None:
            result.pop(key, None)
        else:
            result[key] = _merge_patch(result.get(key), value)
    return result


def handler(event):
    target = event.get("target") if isinstance(event, dict) else None
    patch = event.get("patch")
    if target is None:
        return {"ok": False, "error": "target is required"}
    if patch is None:
        return {"ok": False, "error": "patch is required"}
    try:
        result = _merge_patch(target, patch)
        return {"ok": True, "result": result}
    except Exception as e:
        return {"ok": False, "error": str(e)}
