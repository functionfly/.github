def handler(event):
    content = event.get("content")
    from_encoding = event.get("from_encoding")
    to_encoding = event.get("to_encoding", "utf-8")
    errors = event.get("errors", "strict")  # strict, ignore, replace

    if content is None:
        return {"ok": False, "error": "content is required"}

    if not to_encoding:
        return {"ok": False, "error": "to_encoding is required"}

    valid_errors = ["strict", "ignore", "replace"]
    if errors not in valid_errors:
        return {"ok": False, "error": f"errors must be one of: {', '.join(valid_errors)}"}

    try:
        # Handle different input types
        if isinstance(content, str):
            # String input - assume it's already decoded
            if from_encoding:
                # Re-encode to bytes using from_encoding, then decode to target
                content_bytes = content.encode(from_encoding, errors=errors)
                result = content_bytes.decode(to_encoding, errors=errors)
            else:
                # Assume UTF-8 and convert to target encoding
                result = content.encode(to_encoding, errors=errors).decode(to_encoding)
        elif isinstance(content, bytes):
            # Bytes input - decode from source encoding, then encode to target
            source_encoding = from_encoding or "utf-8"
            decoded_string = content.decode(source_encoding, errors=errors)
            result = decoded_string.encode(to_encoding, errors=errors).decode(to_encoding)
        else:
            return {"ok": False, "error": "content must be a string or bytes"}

        conversion_info = {
            "original_type": "string" if isinstance(content, str) else "bytes",
            "original_length": len(content),
            "from_encoding": from_encoding or "auto-detected",
            "to_encoding": to_encoding,
            "errors_handling": errors,
            "converted_length": len(result),
            "was_converted": from_encoding != to_encoding if from_encoding else False
        }

        return {
            "ok": True,
            "result": result,
            "conversion_info": conversion_info
        }

    except (UnicodeDecodeError, UnicodeEncodeError, LookupError) as e:
        return {
            "ok": False,
            "error": f"encoding conversion failed: {str(e)}",
            "conversion_info": {
                "from_encoding": from_encoding,
                "to_encoding": to_encoding,
                "errors_handling": errors,
                "error_type": type(e).__name__
            }
        }
    except Exception as e:
        return {"ok": False, "error": f"failed to convert charset: {str(e)}"}