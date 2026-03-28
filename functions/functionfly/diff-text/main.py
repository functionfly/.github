import difflib

def handler(event):
    try:
        text1 = event.get("text1", "") if isinstance(event, dict) else ""
        text2 = event.get("text2", "") if isinstance(event, dict) else ""
        if not text1:
            return {"ok": False, "error": "text1 is required"}
        if not text2:
            return {"ok": False, "error": "text2 is required"}
        lines1 = text1.splitlines()
        lines2 = text2.splitlines()
        diff = list(difflib.unified_diff(lines1, lines2, lineterm=''))
        added = sum(1 for line in diff if line.startswith('+') and not line.startswith('+++'))
        removed = sum(1 for line in diff if line.startswith('-') and not line.startswith('---'))
        return {"ok": True, "diff": '\n'.join(diff), "added": added, "removed": removed}
    except Exception as e:
        return {"ok": False, "error": str(e)}
