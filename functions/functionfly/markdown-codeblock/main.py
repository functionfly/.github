import re

def extract_code_blocks(markdown: str) -> list:
    """Extract code blocks from Markdown"""
    code_blocks = []
    pattern = r'```(\w*)\n(.*?)```'
    matches = re.findall(pattern, markdown, re.DOTALL)
    for language, code in matches:
        code_blocks.append({"language": language or "text", "code": code.strip()})
    return code_blocks

def handler(event):
    try:
        markdown = event.get("markdown", "") if isinstance(event, dict) else ""
        if not markdown:
            return {"ok": False, "error": "markdown is required"}
        code_blocks = extract_code_blocks(markdown)
        return {"ok": True, "code_blocks": code_blocks}
    except Exception as e:
        return {"ok": False, "error": str(e)}
