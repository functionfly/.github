import chardet


def handler(event):
    content = event.get("content")
    confidence_threshold = event.get("confidence_threshold", 0.7)

    if content is None:
        return {"ok": False, "error": "content is required"}

    # Convert bytes to string if needed
    if isinstance(content, bytes):
        # For bytes input, we'll try to detect encoding
        pass
    elif isinstance(content, str):
        # For string input, assume it's already properly decoded
        return {
            "ok": True,
            "result": {
                "encoding": "utf-8",
                "confidence": 1.0,
                "is_valid_utf8": True,
                "content_type": "string",
                "bytes_length": len(content.encode('utf-8'))
            }
        }
    else:
        return {"ok": False, "error": "content must be a string or bytes"}

    try:
        # Use chardet to detect encoding
        detection = chardet.detect(content)

        encoding = detection.get('encoding', 'unknown').lower()
        confidence = detection.get('confidence', 0.0)
        language = detection.get('language')

        # Additional validation
        is_valid_utf8 = False
        try:
            content.decode('utf-8')
            is_valid_utf8 = True
        except UnicodeDecodeError:
            pass

        is_valid_latin1 = False
        try:
            content.decode('latin-1')
            is_valid_latin1 = True
        except UnicodeDecodeError:
            pass

        result = {
            "encoding": encoding,
            "confidence": confidence,
            "language": language,
            "is_confident": confidence >= confidence_threshold,
            "is_valid_utf8": is_valid_utf8,
            "is_valid_latin1": is_valid_latin1,
            "bytes_length": len(content),
            "content_type": "bytes"
        }

        # Add suggestions for alternative encodings
        suggestions = []
        if not is_valid_utf8:
            if is_valid_latin1 and encoding != 'iso-8859-1':
                suggestions.append({
                    "encoding": "iso-8859-1",
                    "reason": "content is valid latin-1"
                })
            if confidence < confidence_threshold:
                suggestions.append({
                    "encoding": "utf-8",
                    "reason": "fallback to utf-8"
                })

        result["suggestions"] = suggestions

        return {
            "ok": True,
            "result": result
        }

    except Exception as e:
        return {"ok": False, "error": f"failed to detect encoding: {str(e)}"}