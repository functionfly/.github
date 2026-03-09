def handler(event):
    headers = event.get("headers", {})
    format_output = event.get("format", "object")  # "object", "string", "array"

    if not isinstance(headers, dict):
        return {"ok": False, "error": "headers must be an object"}

    try:
        # Normalize header names (HTTP headers are case-insensitive)
        normalized_headers = {}
        for key, value in headers.items():
            # Convert to title case for HTTP header convention
            normalized_key = '-'.join(word.capitalize() for word in key.replace('_', '-').split('-'))
            normalized_headers[normalized_key] = value

        result = {}

        if format_output == "object" or format_output == "all":
            result["headers_object"] = normalized_headers

        if format_output == "string" or format_output == "all":
            header_lines = []
            for key, value in normalized_headers.items():
                if isinstance(value, list):
                    # Multiple values for same header
                    for v in value:
                        header_lines.append(f"{key}: {v}")
                else:
                    header_lines.append(f"{key}: {value}")
            result["headers_string"] = '\n'.join(header_lines)

        if format_output == "array" or format_output == "all":
            header_array = []
            for key, value in normalized_headers.items():
                if isinstance(value, list):
                    for v in value:
                        header_array.append([key, str(v)])
                else:
                    header_array.append([key, str(value)])
            result["headers_array"] = header_array

        # Add some useful computed fields
        content_type = normalized_headers.get('Content-Type', '')
        if content_type and ';' not in content_type and headers.get('charset'):
            # Add charset to content-type if not present
            charset = headers.get('charset')
            content_type = f"{content_type}; charset={charset}"

        result["content_type_header"] = content_type
        result["has_authorization"] = 'Authorization' in normalized_headers
        result["header_count"] = len(normalized_headers)

        return {
            "ok": True,
            "result": result
        }

    except Exception as e:
        return {"ok": False, "error": f"failed to build headers: {str(e)}"}