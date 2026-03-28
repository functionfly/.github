from datetime import date


def handler(event):
    """Generate human-readable release notes."""
    try:
        version = event.get("version")
        if not version:
            return {"ok": False, "error": "version is required"}

        features = event.get("features", [])
        fixes = event.get("fixes", [])
        breaking = event.get("breaking_changes", [])
        deprecations = event.get("deprecations", [])
        release_date = event.get("date", str(date.today()))
        fmt = event.get("format", "markdown")

        if fmt == "markdown":
            lines = [f"# Release {version}", f"*Released: {release_date}*", ""]

            if breaking:
                lines.extend(["## ⚠️ Breaking Changes", ""])
                for item in breaking:
                    lines.append(f"- {item}")
                lines.append("")

            if features:
                lines.extend(["## What's New", ""])
                for item in features:
                    lines.append(f"- {item}")
                lines.append("")

            if fixes:
                lines.extend(["## Bug Fixes", ""])
                for item in fixes:
                    lines.append(f"- {item}")
                lines.append("")

            if deprecations:
                lines.extend(["## Deprecations", ""])
                for item in deprecations:
                    lines.append(f"- {item}")
                lines.append("")

            notes = "\n".join(lines)

        elif fmt == "html":
            parts = [f"<h1>Release {version}</h1>", f"<p><em>Released: {release_date}</em></p>"]
            if breaking:
                parts.append("<h2>⚠️ Breaking Changes</h2><ul>")
                parts.extend([f"<li>{i}</li>" for i in breaking])
                parts.append("</ul>")
            if features:
                parts.append("<h2>What's New</h2><ul>")
                parts.extend([f"<li>{i}</li>" for i in features])
                parts.append("</ul>")
            if fixes:
                parts.append("<h2>Bug Fixes</h2><ul>")
                parts.extend([f"<li>{i}</li>" for i in fixes])
                parts.append("</ul>")
            notes = "\n".join(parts)

        else:  # text
            lines = [f"Release {version} - {release_date}", "=" * 40, ""]
            if breaking:
                lines.extend(["BREAKING CHANGES:", *[f"  * {i}" for i in breaking], ""])
            if features:
                lines.extend(["New Features:", *[f"  * {i}" for i in features], ""])
            if fixes:
                lines.extend(["Bug Fixes:", *[f"  * {i}" for i in fixes], ""])
            notes = "\n".join(lines)

        return {"ok": True, "result": f"Release notes for {version} generated", "notes": notes}
    except Exception as e:
        return {"ok": False, "error": str(e)}
