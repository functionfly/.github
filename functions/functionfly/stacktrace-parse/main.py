import re


def parse_python(text):
    frames = []
    exception = None
    message = None
    lines = text.strip().splitlines()

    for i, line in enumerate(lines):
        # Python frame: File "path", line N, in func
        match = re.match(r'\s+File "([^"]+)", line (\d+), in (.+)', line)
        if match:
            frames.append({
                "file": match.group(1),
                "line": int(match.group(2)),
                "function": match.group(3).strip(),
            })
        # Exception line (last line)
        elif i == len(lines) - 1 and ":" in line and not line.startswith(" "):
            parts = line.split(":", 1)
            exception = parts[0].strip()
            message = parts[1].strip() if len(parts) > 1 else None

    return frames, exception, message, "python"


def parse_java(text):
    frames = []
    exception = None
    message = None
    lines = text.strip().splitlines()

    for i, line in enumerate(lines):
        # Java frame: at package.Class.method(File.java:N)
        match = re.match(r'\s+at ([\w\.$]+)\(([\w\.]+):(\d+)\)', line)
        if match:
            frames.append({
                "function": match.group(1),
                "file": match.group(2),
                "line": int(match.group(3)),
            })
        elif i == 0 and ":" in line:
            parts = line.split(":", 1)
            exception = parts[0].strip()
            message = parts[1].strip() if len(parts) > 1 else None

    return frames, exception, message, "java"


def parse_javascript(text):
    frames = []
    exception = None
    message = None
    lines = text.strip().splitlines()

    for i, line in enumerate(lines):
        # JS frame: at Function (file:line:col) or at file:line:col
        match = re.match(r'\s+at (?:(.+?) \()?(.+?):(\d+):(\d+)\)?', line)
        if match:
            frames.append({
                "function": match.group(1) or "<anonymous>",
                "file": match.group(2),
                "line": int(match.group(3)),
                "column": int(match.group(4)),
            })
        elif i == 0:
            parts = line.split(":", 1)
            exception = parts[0].strip()
            message = parts[1].strip() if len(parts) > 1 else None

    return frames, exception, message, "javascript"


def detect_language(text):
    if "Traceback (most recent call last)" in text or 'File "' in text:
        return "python"
    if "\tat " in text and ".java:" in text:
        return "java"
    if "\n    at " in text or "Error:" in text:
        return "javascript"
    return "unknown"


def handler(event):
    """Parse a stack trace into structured frames."""
    try:
        stacktrace = event.get("stacktrace")
        if not stacktrace:
            return {"ok": False, "error": "stacktrace is required"}

        language = event.get("language", "auto")
        if language == "auto":
            language = detect_language(stacktrace)

        if language == "python":
            frames, exception, message, lang = parse_python(stacktrace)
        elif language == "java":
            frames, exception, message, lang = parse_java(stacktrace)
        elif language in ("javascript", "js", "node"):
            frames, exception, message, lang = parse_javascript(stacktrace)
        else:
            frames, exception, message, lang = [], None, None, language

        return {
            "ok": True,
            "language": lang,
            "frames": frames,
            "exception": exception,
            "message": message,
            "frame_count": len(frames),
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
