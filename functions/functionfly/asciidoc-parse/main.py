import re

def parse_asciidoc(asciidoc: str) -> dict:
    """Parse AsciiDoc to structured data"""
    lines = asciidoc.split('\n')
    title = ""
    sections = []
    current_section = None
    for line in lines:
        line = line.strip()
        if line.startswith('= '):
            title = line[2:].strip()
        elif line.startswith('== '):
            if current_section:
                sections.append(current_section)
            current_section = {"level": 1, "title": line[3:].strip(), "content": ""}
        elif line.startswith('=== '):
            if current_section:
                sections.append(current_section)
            current_section = {"level": 2, "title": line[4:].strip(), "content": ""}
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
        asciidoc = event.get("asciidoc", "") if isinstance(event, dict) else ""
        if not asciidoc:
            return {"ok": False, "error": "asciidoc is required"}
        result = parse_asciidoc(asciidoc)
        return {"ok": True, **result}
    except Exception as e:
        return {"ok": False, "error": str(e)}
