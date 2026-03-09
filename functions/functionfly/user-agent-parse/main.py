import re


def _simple_parse(ua):
    ua = (ua or "").strip()
    browser = "Unknown"
    os = "Unknown"
    device = "Desktop"
    is_bot = bool(re.search(r"bot|crawler|spider|slurp|googlebot|bingbot", ua, re.I))

    if not ua:
        return {"browser": browser, "os": os, "device": device, "is_bot": is_bot}

    # Browsers
    if "Edg/" in ua or "Edge/" in ua:
        browser = "Edge"
    elif "Chrome/" in ua and "Chromium" not in ua:
        browser = "Chrome"
    elif "Firefox/" in ua or "FxiOS/" in ua:
        browser = "Firefox"
    elif "Safari/" in ua and "Chrome" not in ua:
        browser = "Safari"
    elif "OPR/" in ua or "Opera" in ua:
        browser = "Opera"

    # OS
    if "Windows NT" in ua:
        os = "Windows"
    elif "Mac OS" in ua or "Macintosh" in ua:
        os = "macOS"
    elif "Linux" in ua and "Android" not in ua:
        os = "Linux"
    elif "Android" in ua:
        os = "Android"
    elif "iPhone" in ua or "iPad" in ua:
        os = "iOS"

    # Device
    if "Mobile" in ua and "iPad" not in ua or "Android" in ua:
        device = "Mobile"
    elif "Tablet" in ua or "iPad" in ua:
        device = "Tablet"

    return {"browser": browser, "os": os, "device": device, "is_bot": is_bot}


def handler(event):
    if isinstance(event, dict):
        ua = event.get("user_agent", event.get("ua", ""))
    else:
        ua = ""

    if ua is None:
        ua = ""

    try:
        import user_agents
        parsed = user_agents.parse(ua)
        return {"ok": True, "browser": parsed.browser.family or "Unknown", "os": parsed.os.family or "Unknown", "device": "Mobile" if parsed.is_mobile else ("Tablet" if parsed.is_tablet else "Desktop"), "is_bot": parsed.is_bot, "raw": ua}
    except ImportError:
        out = _simple_parse(ua)
        return {"ok": True, "browser": out["browser"], "os": out["os"], "device": out["device"], "is_bot": out["is_bot"], "raw": ua}
    except Exception as e:
        return {"ok": False, "error": str(e)}
