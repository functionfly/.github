from urllib.parse import parse_qs
import http.cookies


def handler(event):
    cookie_input = event.get("cookies")
    format_output = event.get("format", "object")  # "object" or "array"

    if not cookie_input:
        return {"ok": False, "error": "cookies is required"}

    if not isinstance(cookie_input, str):
        return {"ok": False, "error": "cookies must be a string"}

    try:
        # Parse cookies using http.cookies module
        cookie = http.cookies.SimpleCookie()
        cookie.load(cookie_input)

        parsed_cookies = {}
        cookie_array = []

        for name, morsel in cookie.items():
            cookie_info = {
                "name": name,
                "value": morsel.value,
                "path": morsel.get("path", ""),
                "domain": morsel.get("domain", ""),
                "secure": "secure" in morsel,
                "httponly": "httponly" in morsel,
                "max_age": morsel.get("max-age"),
                "expires": morsel.get("expires", ""),
                "samesite": morsel.get("samesite", "")
            }
            parsed_cookies[name] = cookie_info
            cookie_array.append(cookie_info)

        result = {
            "cookies": parsed_cookies if format_output == "object" else cookie_array,
            "count": len(parsed_cookies)
        }

        # Add some useful computed fields
        session_cookies = []
        persistent_cookies = []

        for name, info in parsed_cookies.items():
            if info.get("max_age") or info.get("expires"):
                persistent_cookies.append(name)
            else:
                session_cookies.append(name)

        result["session_cookies"] = session_cookies
        result["persistent_cookies"] = persistent_cookies
        result["has_secure"] = any(info.get("secure", False) for info in parsed_cookies.values())
        result["has_httponly"] = any(info.get("httponly", False) for info in parsed_cookies.values())

        return {
            "ok": True,
            "result": result
        }

    except Exception as e:
        return {"ok": False, "error": f"failed to parse cookies: {str(e)}"}