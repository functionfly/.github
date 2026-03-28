import re

def parse_org_mode(org: str) -> dict:
    """Parse Org-mode to structured data"""
    lines = org.split('\n')
    title = ""
    sections = []
    current_section = None
    for line in lines:
        line = line.strip()
        if line.startswith('#+TITLE:'):
            title = line[8:].strip()
        elif line.startswith('* '):
            if current_section:
                sections.append(current_section)
            current_section = {"level": 1, "title": line[2:].strip(), "content": ""}
        elif line.startswith('** '):
            if current_section:
                sections.append(current_section)
            current_section = {"level": 2, "title": line[3:].strip(), "content": ""}
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
        org = event.get("org", "") if isinstance(event, dict) else ""
        if not org:
            return {"ok": False, "error": "org is required"}
        result = parse_org_mode(org)
        return {"ok": True, **result}
    except Exception as e:
        return {"ok": False, "error": str(e)}
