import http.cookies
from datetime import datetime, timedelta


def handler(event):
    cookies = event.get("cookies", [])
    format_output = event.get("format", "string")  # "string", "object", "array"

    if not isinstance(cookies, list):
        return {"ok": False, "error": "cookies must be an array"}

    try:
        cookie_objects = []

        for cookie_data in cookies:
            if not isinstance(cookie_data, dict):
                return {"ok": False, "error": "each cookie must be an object"}

            name = cookie_data.get("name")
            value = cookie_data.get("value")

            if not name or value is None:
                return {"ok": False, "error": "each cookie must have name and value"}

            morsel = http.cookies.Morsel()
            morsel.set(name, str(value), str(value))

            # Set optional attributes
            if cookie_data.get("path"):
                morsel["path"] = cookie_data["path"]
            if cookie_data.get("domain"):
                morsel["domain"] = cookie_data["domain"]
            if cookie_data.get("max_age"):
                morsel["max-age"] = str(cookie_data["max_age"])
            if cookie_data.get("expires"):
                expires = cookie_data["expires"]
                if isinstance(expires, (int, float)):
                    # Convert timestamp to date string
                    expires_date = datetime.fromtimestamp(expires)
                    morsel["expires"] = expires_date.strftime("%a, %d %b %Y %H:%M:%S GMT")
                else:
                    morsel["expires"] = expires
            if cookie_data.get("secure"):
                morsel["secure"] = True
            if cookie_data.get("httponly"):
                morsel["httponly"] = True
            if cookie_data.get("samesite"):
                morsel["samesite"] = cookie_data["samesite"]

            cookie_objects.append(morsel)

        result = {}

        if format_output == "string" or format_output == "all":
            cookie_strings = []
            for morsel in cookie_objects:
                cookie_strings.append(morsel.OutputString())
            result["cookies_string"] = "; ".join(cookie_strings)

        if format_output == "object" or format_output == "all":
            cookie_dict = {}
            for morsel in cookie_objects:
                cookie_dict[morsel.key] = morsel.value
            result["cookies_object"] = cookie_dict

        if format_output == "array" or format_output == "all":
            result["cookies_array"] = cookies.copy()

        result["count"] = len(cookie_objects)

        # Add computed fields
        has_secure = any(cookie_data.get("secure", False) for cookie_data in cookies)
        has_httponly = any(cookie_data.get("httponly", False) for cookie_data in cookies)
        has_expires = any(cookie_data.get("expires") or cookie_data.get("max_age") for cookie_data in cookies)

        result["has_secure"] = has_secure
        result["has_httponly"] = has_httponly
        result["has_expires"] = has_expires

        return {
            "ok": True,
            "result": result
        }

    except Exception as e:
        return {"ok": False, "error": f"failed to build cookies: {str(e)}"}