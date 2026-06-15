import json
import re


def handler(event):
    if isinstance(event, dict):
        pattern = event.get("pattern", "")
        text = event.get("text", "")
        flags = event.get("flags", "")
    else:
        pattern = ""
        text = ""
        flags = ""

    if not pattern:
        return {"ok": False, "error": "pattern is required"}

    if not isinstance(pattern, str):
        return {"ok": False, "error": "pattern must be a string"}

    if not isinstance(text, str):
        return {"ok": False, "error": "text must be a string"}

    if not isinstance(flags, str):
        return {"ok": False, "error": "flags must be a string"}

    flag_value = 0
    valid_flags = set("imsx")
    for f in flags.lower():
        if f == "i":
            flag_value |= re.IGNORECASE
        elif f == "m":
            flag_value |= re.MULTILINE
        elif f == "s":
            flag_value |= re.DOTALL
        elif f == "x":
            flag_value |= re.VERBOSE
        else:
            return {"ok": False, "error": f"invalid flag: {f}. Valid flags: i, m, s, x"}

    try:
        regex = re.compile(pattern, flag_value)
    except re.error as e:
        return {"ok": False, "error": f"invalid regex pattern: {str(e)}"}

    matches = []
    named_groups = None

    for match in regex.finditer(text):
        groups = match.groups()
        group_list = []
        for g in groups:
            group_list.append(g if g is not None else "")

        if match.groupdict():
            named_groups = match.groupdict()

        matches.append(group_list)

    if named_groups:
        return {
            "ok": True,
            "matched": len(matches) > 0,
            "matches": matches,
            "groups": named_groups
        }

    return {
        "ok": True,
        "matched": len(matches) > 0,
        "matches": matches
    }
