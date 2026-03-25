CATEGORIES = {
    1: {"name": "informational", "description": "Request received, continuing process"},
    2: {"name": "success", "description": "Request successfully received, understood, and accepted"},
    3: {"name": "redirection", "description": "Further action needs to be taken to complete the request"},
    4: {"name": "client_error", "description": "Request contains bad syntax or cannot be fulfilled"},
    5: {"name": "server_error", "description": "Server failed to fulfill an apparently valid request"},
}


def handler(event):
    status_code = event.get("status_code") if isinstance(event, dict) else None

    if status_code is None:
        return {"ok": False, "error": "status_code is required"}

    try:
        code = int(status_code)
    except (TypeError, ValueError):
        return {"ok": False, "error": "status_code must be an integer"}

    if code < 100 or code > 599:
        return {"ok": False, "error": f"Invalid HTTP status code: {code}"}

    digit = code // 100
    cat = CATEGORIES.get(digit)
    if not cat:
        return {"ok": False, "error": f"Unknown status code category for: {code}"}

    return {
        "ok": True,
        "status_code": code,
        "category": cat["name"],
        "description": cat["description"],
        "is_informational": digit == 1,
        "is_success": digit == 2,
        "is_redirection": digit == 3,
        "is_client_error": digit == 4,
        "is_server_error": digit == 5,
    }
