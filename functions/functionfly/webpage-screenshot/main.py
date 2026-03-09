def handler(event):
    url = event.get("url", "") if isinstance(event, dict) else ""
    if not url:
        return {"ok": False, "error": "url is required"}
    try:
        from playwright.sync_api import sync_playwright
    except ImportError:
        return {"ok": False, "error": "Playwright required; pip install playwright && playwright install chromium"}
    try:
        import base64
        with sync_playwright() as p:
            browser = p.chromium.launch()
            page = browser.new_page(viewport={"width": int(event.get("width", 1280)), "height": int(event.get("height", 720))})
            page.goto(url, wait_until="networkidle", timeout=15000)
            buf = page.screenshot(type="png")
            browser.close()
        return {"ok": True, "image_base64": base64.b64encode(buf).decode("ascii")}
    except Exception as e:
        return {"ok": False, "error": str(e)}
