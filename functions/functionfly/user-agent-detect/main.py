try:
    from user_agents import parse
    HAS_USER_AGENTS = True
except ImportError:
    HAS_USER_AGENTS = False


def _fallback_parse(user_agent):
    """Fallback parser when user-agents library is not available"""
    ua = user_agent.lower()

    # Basic detection
    browser = "Unknown"
    browser_version = ""
    os = "Unknown"
    device = "Unknown"

    # Browser detection
    if 'chrome' in ua and 'edg' not in ua:
        browser = "Chrome"
        if 'chrome/' in ua:
            try:
                browser_version = ua.split('chrome/')[1].split(' ')[0].split('.')[0]
            except:
                pass
    elif 'firefox' in ua:
        browser = "Firefox"
        if 'firefox/' in ua:
            try:
                browser_version = ua.split('firefox/')[1].split(' ')[0].split('.')[0]
            except:
                pass
    elif 'safari' in ua and 'chrome' not in ua:
        browser = "Safari"
        if 'version/' in ua:
            try:
                browser_version = ua.split('version/')[1].split(' ')[0].split('.')[0]
            except:
                pass
    elif 'edg' in ua:
        browser = "Edge"
        if 'edg/' in ua:
            try:
                browser_version = ua.split('edg/')[1].split(' ')[0].split('.')[0]
            except:
                pass

    # OS detection
    if 'windows' in ua:
        os = "Windows"
    elif 'mac os x' in ua or 'macos' in ua:
        os = "macOS"
    elif 'linux' in ua:
        os = "Linux"
    elif 'android' in ua:
        os = "Android"
    elif 'ios' in ua or 'iphone' in ua or 'ipad' in ua:
        os = "iOS"

    # Device detection
    if 'mobile' in ua or 'android' in ua or 'iphone' in ua:
        device = "Mobile"
    elif 'tablet' in ua or 'ipad' in ua:
        device = "Tablet"
    else:
        device = "Desktop"

    return {
        "browser": browser,
        "browser_version": browser_version,
        "os": os,
        "device": device,
        "is_mobile": device == "Mobile",
        "is_tablet": device == "Tablet",
        "is_desktop": device == "Desktop",
        "is_bot": any(bot in ua for bot in ['bot', 'crawler', 'spider', 'scraper'])
    }


def handler(event):
    user_agent = event.get("user_agent")

    if not user_agent:
        return {"ok": False, "error": "user_agent is required"}

    if not isinstance(user_agent, str):
        return {"ok": False, "error": "user_agent must be a string"}

    try:
        if HAS_USER_AGENTS:
            ua = parse(user_agent)

            result = {
                "browser": ua.browser.family,
                "browser_version": ua.browser.version_string,
                "os": ua.os.family,
                "os_version": ua.os.version_string,
                "device": ua.device.family,
                "device_brand": ua.device.brand,
                "device_model": ua.device.model,
                "is_mobile": ua.is_mobile,
                "is_tablet": ua.is_tablet,
                "is_touch_capable": ua.is_touch_capable,
                "is_pc": ua.is_pc,
                "is_bot": ua.is_bot
            }
        else:
            result = _fallback_parse(user_agent)

        # Add some additional computed fields
        result["user_agent_string"] = user_agent
        result["is_legacy_browser"] = False

        # Check for legacy browsers
        if result.get("browser") in ["Internet Explorer", "IE"]:
            result["is_legacy_browser"] = True
        elif result.get("browser") == "Chrome" and result.get("browser_version"):
            try:
                version = int(result["browser_version"].split('.')[0])
                if version < 70:  # Arbitrary cutoff for "legacy"
                    result["is_legacy_browser"] = True
            except:
                pass

        return {
            "ok": True,
            "result": result
        }

    except Exception as e:
        return {"ok": False, "error": f"failed to parse user agent: {str(e)}"}