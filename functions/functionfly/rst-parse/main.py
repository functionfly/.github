import re

def parse_rst(rst: str) -> dict:
    """Parse reStructuredText to structured data"""
    lines = rst.split('\n')
    title = ""
    sections = []
    current_section = None
    for i, line in enumerate(lines):
        line = line.strip()
        if i > 0 and lines[i-1].strip() and all(c == '=' for c in line):
            title = lines[i-1].strip()
        elif i > 0 and lines[i-1].strip() and all(c == '-' for c in line):
            if current_section:
                sections.append(current_section)
            current_section = {"level": 1, "title": lines[i-1].strip(), "content": ""}
        elif current_section:
            if current_section["content"]:
                current_section["content"] += "\n" + line
            else:
                current_section["content"] = line
    if current_section:
        sections.append(current_section)
    return {"title": title, "sections": sections}

def handler(event):
    try:
        rst = event.get("rst", "") if isinstance(event, dict) else ""
        if not rst:
            return {"ok": False, "error": "rst is required"}
        result = parse_rst(rst)
        return {"ok": True, **result}
    except Exception as e:
        return {"ok": False, "error": str(e)}
