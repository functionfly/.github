import difflib

def handler(event):
    try:
        text1 = event.get("text1", "") if isinstance(event, dict) else ""
        text2 = event.get("text2", "") if isinstance(event, dict) else ""
        filename1 = event.get("filename1", "a/file.txt") if isinstance(event, dict) else "a/file.txt"
        filename2 = event.get("filename2", "b/file.txt") if isinstance(event, dict) else "b/file.txt"
        if not text1:
            return {"ok": False, "error": "text1 is required"}
        if not text2:
            return {"ok": False, "error": "text2 is required"}
        lines1 = text1.splitlines()
        lines2 = text2.splitlines()
        diff = list(difflib.unified_diff(lines1, lines2, fromfile=filename1, tofile=filename2, lineterm=''))
        return {"ok": True, "diff": '\n'.join(diff)}
    except Exception as e:
        return {"ok": False, "error": str(e)}
