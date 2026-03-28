import re

def apply_patch(text: str, patch: str) -> str:
    """Apply a patch to text"""
    lines = text.splitlines()
    patch_lines = patch.splitlines()
    i = 0
    while i < len(patch_lines):
        line = patch_lines[i]
        if line.startswith('@@'):
            match = re.match(r'@@ -(\d+),?(\d*) \+(\d+),?(\d*) @@', line)
            if match:
                start_line = int(match.group(1)) - 1
                end_line = start_line + int(match.group(2) or 1)
                new_start = int(match.group(3)) - 1
                new_end = new_start + int(match.group(4) or 1)
                j = i + 1
                while j < len(patch_lines) and not patch_lines[j].startswith('@@'):
                    patch_line = patch_lines[j]
                    if patch_line.startswith('-'):
                        if start_line < len(lines):
                            lines.pop(start_line)
                            end_line -= 1
                    elif patch_line.startswith('+'):
                        lines.insert(start_line, patch_line[1:])
                        start_line += 1
                    else:
                        start_line += 1
                    j += 1
                i = j
            else:
                i += 1
        else:
            i += 1
    return '\n'.join(lines)

def handler(event):
    try:
        text = event.get("text", "") if isinstance(event, dict) else ""
        patch = event.get("patch", "") if isinstance(event, dict) else ""
        if not text:
            return {"ok": False, "error": "text is required"}
        if not patch:
            return {"ok": False, "error": "patch is required"}
        result = apply_patch(text, patch)
        return {"ok": True, "result": result}
    except Exception as e:
        return {"ok": False, "error": str(e)}
