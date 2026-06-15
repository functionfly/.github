import ast
import json
import re


def handler(event):
    if isinstance(event, dict):
        code = event.get("code", "")
        language = event.get("language", "python")
    else:
        code = ""
        language = "python"

    if not code:
        return {"ok": False, "error": "code is required"}

    if not isinstance(code, str):
        return {"ok": False, "error": "code must be a string"}

    language_lower = language.lower()
    valid_languages = {"python", "javascript", "typescript", "html", "css", "json"}
    if language_lower not in valid_languages:
        return {"ok": False, "error": f"unsupported language: {language}. Supported: {', '.join(sorted(valid_languages))}"}

    try:
        if language_lower == "python":
            formatted = format_python(code)
        elif language_lower == "json":
            formatted = format_json(code)
        else:
            formatted = basic_cleanup(code)

        return {"ok": True, "result": formatted}
    except Exception as e:
        return {"ok": False, "error": str(e)}


def format_python(code):
    try:
        tree = ast.parse(code)
    except SyntaxError as e:
        raise ValueError(f"Python syntax error: {str(e)}")

    formatted_lines = []

    for node in ast.walk(tree):
        if isinstance(node, ast.Module):
            for child in node.body:
                formatted_lines.append(ast.unparse(child))

    return "\n\n".join(formatted_lines)


def format_json(code):
    try:
        parsed = json.loads(code)
        return json.dumps(parsed, indent=4)
    except json.JSONDecodeError as e:
        raise ValueError(f"JSON parse error: {str(e)}")


def basic_cleanup(code):
    lines = code.split("\n")
    cleaned_lines = []
    for line in lines:
        cleaned_lines.append(line.rstrip())
    while cleaned_lines and not cleaned_lines[0]:
        cleaned_lines.pop(0)
    while cleaned_lines and not cleaned_lines[-1]:
        cleaned_lines.pop()
    return "\n".join(cleaned_lines)
