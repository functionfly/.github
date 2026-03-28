import re

def handler(event):
    text = event.get("text") if isinstance(event, dict) else None
    if not text or not isinstance(text, str):
        return {"ok": False, "error": "text (string) is required"}
    try:
        structure = {"headings": [], "paragraphs": [], "lists": [], "tables": [], "metadata": {}}
        lines = text.split("\n")
        current_para = []
        current_list = []
        in_list = False
        for i, line in enumerate(lines):
            stripped = line.strip()
            if not stripped:
                if current_para:
                    structure["paragraphs"].append({"text": " ".join(current_para), "line": i})
                    current_para = []
                if current_list:
                    structure["lists"].append({"items": current_list, "line": i})
                    current_list = []
                    in_list = False
                continue
            # Detect headings (markdown or ALL CAPS short lines)
            if re.match(r"^#{1,6}\s+", stripped):
                level = len(re.match(r"^(#+)", stripped).group(1))
                heading_text = re.sub(r"^#+\s+", "", stripped)
                structure["headings"].append({"text": heading_text, "level": level, "line": i})
            elif stripped.isupper() and len(stripped.split()) <= 8 and len(stripped) > 3:
                structure["headings"].append({"text": stripped, "level": 1, "line": i})
            # Detect list items
            elif re.match(r"^[-*•]\s+", stripped) or re.match(r"^\d+[.)]\s+", stripped):
                item_text = re.sub(r"^[-*•\d.)\s]+", "", stripped)
                current_list.append(item_text)
                in_list = True
            # Detect table rows
            elif "|" in stripped and stripped.count("|") >= 2:
                cells = [c.strip() for c in stripped.split("|") if c.strip()]
                if cells:
                    structure["tables"].append({"cells": cells, "line": i})
            else:
                if in_list:
                    current_list.append(stripped)
                else:
                    current_para.append(stripped)
        if current_para:
            structure["paragraphs"].append({"text": " ".join(current_para), "line": len(lines)})
        if current_list:
            structure["lists"].append({"items": current_list, "line": len(lines)})
        structure["metadata"] = {
            "total_lines": len(lines),
            "heading_count": len(structure["headings"]),
            "paragraph_count": len(structure["paragraphs"]),
            "list_count": len(structure["lists"]),
            "table_row_count": len(structure["tables"]),
            "word_count": len(text.split())
        }
        return {"ok": True, "result": structure, "structure": structure}
    except Exception as e:
        return {"ok": False, "error": str(e)}
