REQUIRED_FIELDS = ["name", "short_name", "start_url", "display", "icons"]
RECOMMENDED_FIELDS = ["theme_color", "background_color", "description", "lang", "scope"]
VALID_DISPLAY = ["fullscreen", "standalone", "minimal-ui", "browser"]


def handler(event):
    """Parse and validate a Web App Manifest."""
    try:
        manifest = event.get("manifest")
        if not manifest:
            return {"ok": False, "error": "manifest is required"}

        missing_fields = [f for f in REQUIRED_FIELDS if f not in manifest]
        warnings = []

        # Validate display
        display = manifest.get("display")
        if display and display not in VALID_DISPLAY:
            warnings.append(f"Invalid display value: {display}")

        # Check icons
        icons = manifest.get("icons", [])
        has_192 = any("192" in str(i.get("sizes", "")) for i in icons)
        has_512 = any("512" in str(i.get("sizes", "")) for i in icons)
        if not has_192:
            warnings.append("Missing 192x192 icon (required for Android)")
        if not has_512:
            warnings.append("Missing 512x512 icon (required for splash screen)")

        # Check recommended fields
        missing_recommended = [f for f in RECOMMENDED_FIELDS if f not in manifest]
        if missing_recommended:
            warnings.append(f"Missing recommended fields: {', '.join(missing_recommended)}")

        return {
            "ok": True,
            "valid": len(missing_fields) == 0,
            "name": manifest.get("name"),
            "short_name": manifest.get("short_name"),
            "start_url": manifest.get("start_url"),
            "display": display,
            "icons": icons,
            "missing_fields": missing_fields,
            "warnings": warnings,
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
