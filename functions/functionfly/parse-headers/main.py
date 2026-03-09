from urllib.parse import parse_qs
import email.utils


def handler(event):
    headers_input = event.get("headers")
    format_output = event.get("format", "object")  # "object" or "array"

    if not headers_input:
        return {"ok": False, "error": "headers is required"}

    if not isinstance(headers_input, (str, dict, list)):
        return {"ok": False, "error": "headers must be a string, object, or array"}

    try:
        parsed_headers = {}

        if isinstance(headers_input, str):
            # Parse raw HTTP headers string
            lines = headers_input.strip().split('\n')
            for line in lines:
                line = line.strip()
                if ':' in line:
                    key, value = line.split(':', 1)
                    key = key.strip()
                    value = value.strip()
                    # Handle multiple headers with same name
                    if key in parsed_headers:
                        if not isinstance(parsed_headers[key], list):
                            parsed_headers[key] = [parsed_headers[key]]
                        parsed_headers[key].append(value)
                    else:
                        parsed_headers[key] = value

        elif isinstance(headers_input, dict):
            # Headers already in object format
            parsed_headers = headers_input.copy()

        elif isinstance(headers_input, list):
            # Headers as array of [key, value] pairs
            for item in headers_input:
                if isinstance(item, list) and len(item) >= 2:
                    key, value = item[0], item[1]
                    if key in parsed_headers:
                        if not isinstance(parsed_headers[key], list):
                            parsed_headers[key] = [parsed_headers[key]]
                        parsed_headers[key].append(str(value))
                    else:
                        parsed_headers[key] = str(value)

        # Normalize header names to lowercase for consistency
        normalized_headers = {}
        for key, value in parsed_headers.items():
            normalized_key = key.lower()
            normalized_headers[normalized_key] = value

        # Add some useful computed fields
        content_type = normalized_headers.get('content-type', '')
        content_length = normalized_headers.get('content-length')
        user_agent = normalized_headers.get('user-agent', '')
        authorization = normalized_headers.get('authorization', '')

        # Parse content-type
        media_type = ''
        charset = ''
        if content_type:
            if ';' in content_type:
                media_type, params = content_type.split(';', 1)
                media_type = media_type.strip()
                if 'charset=' in params:
                    charset_part = [p.strip() for p in params.split('charset=') if p.strip()]
                    if charset_part:
                        charset = charset_part[0].split(';')[0].strip().strip('"\'')
            else:
                media_type = content_type.strip()

        # Parse authorization
        auth_type = ''
        auth_credentials = ''
        if authorization:
            parts = authorization.split(' ', 1)
            if len(parts) == 2:
                auth_type = parts[0]
                auth_credentials = parts[1]

        result = {
            "headers": normalized_headers,
            "content_type": media_type,
            "charset": charset,
            "content_length": int(content_length) if content_length and content_length.isdigit() else None,
            "user_agent": user_agent,
            "authorization": {
                "type": auth_type,
                "credentials": auth_credentials
            } if auth_type else None
        }

        if format_output == "array":
            result["headers_array"] = [[key, value] for key, value in normalized_headers.items()]

        return {
            "ok": True,
            "result": result,
            "header_count": len(normalized_headers)
        }

    except Exception as e:
        return {"ok": False, "error": f"failed to parse headers: {str(e)}"}