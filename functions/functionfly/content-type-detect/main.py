import mimetypes
import json
import xml.etree.ElementTree as ET


def _is_json(content):
    """Check if content is valid JSON"""
    try:
        json.loads(content)
        return True
    except (ValueError, TypeError):
        return False


def _is_xml(content):
    """Check if content is valid XML"""
    try:
        ET.fromstring(content)
        return True
    except ET.ParseError:
        return False


def _is_html(content):
    """Check if content looks like HTML"""
    content_lower = content.lower().strip()
    return content_lower.startswith('<!doctype html') or \
           content_lower.startswith('<html') or \
           ('<head>' in content_lower and '<body>' in content_lower)


def _is_css(content):
    """Check if content looks like CSS"""
    content_lower = content.lower().strip()
    return content_lower.startswith('@') or \
           ('{' in content and '}' in content and ':' in content)


def _is_javascript(content):
    """Check if content looks like JavaScript"""
    content_lower = content.lower().strip()
    js_keywords = ['function', 'var ', 'let ', 'const ', 'if ', 'for ', 'while ']
    return any(keyword in content_lower for keyword in js_keywords) or \
           content_lower.startswith('//') or \
           content_lower.startswith('/*')


def handler(event):
    content = event.get("content")
    filename = event.get("filename")
    analyze_content = event.get("analyze_content", True)

    if content is None:
        return {"ok": False, "error": "content is required"}

    if not isinstance(content, str):
        return {"ok": False, "error": "content must be a string"}

    try:
        detected_types = []

        # Method 1: Filename-based detection
        if filename:
            mime_type, encoding = mimetypes.guess_type(filename)
            if mime_type:
                detected_types.append({
                    "method": "filename",
                    "mime_type": mime_type,
                    "encoding": encoding,
                    "confidence": "high"
                })

        # Method 2: Content analysis
        if analyze_content and content.strip():
            content_sample = content.strip()[:1000]  # Analyze first 1000 chars

            if _is_json(content_sample):
                detected_types.append({
                    "method": "content_analysis",
                    "mime_type": "application/json",
                    "confidence": "high"
                })
            elif _is_xml(content_sample):
                detected_types.append({
                    "method": "content_analysis",
                    "mime_type": "application/xml",
                    "confidence": "high"
                })
            elif _is_html(content_sample):
                detected_types.append({
                    "method": "content_analysis",
                    "mime_type": "text/html",
                    "confidence": "high"
                })
            elif _is_css(content_sample):
                detected_types.append({
                    "method": "content_analysis",
                    "mime_type": "text/css",
                    "confidence": "medium"
                })
            elif _is_javascript(content_sample):
                detected_types.append({
                    "method": "content_analysis",
                    "mime_type": "application/javascript",
                    "confidence": "medium"
                })
            else:
                # Fallback to text/plain
                detected_types.append({
                    "method": "content_analysis",
                    "mime_type": "text/plain",
                    "confidence": "low"
                })

        # Determine the best match
        best_match = None
        if detected_types:
            # Prioritize filename detection, then content analysis
            filename_matches = [t for t in detected_types if t["method"] == "filename"]
            content_matches = [t for t in detected_types if t["method"] == "content_analysis"]

            if filename_matches:
                best_match = filename_matches[0]
            elif content_matches:
                best_match = content_matches[0]

        result = {
            "detected_types": detected_types,
            "best_match": best_match,
            "content_length": len(content),
            "has_content": bool(content.strip())
        }

        if best_match:
            result["mime_type"] = best_match["mime_type"]
            result["confidence"] = best_match["confidence"]

        return {
            "ok": True,
            "result": result
        }

    except Exception as e:
        return {"ok": False, "error": f"failed to detect content type: {str(e)}"}