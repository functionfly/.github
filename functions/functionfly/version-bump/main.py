import re


def parse_version(v):
    v = str(v).strip().lstrip("v")
    match = re.match(r'^(\d+)\.(\d+)\.(\d+)(?:-([a-zA-Z0-9\-\.]+))?(?:\+([a-zA-Z0-9\-\.]+))?$', v)
    if not match:
        raise ValueError(f"Invalid semver: {v}")
    return int(match.group(1)), int(match.group(2)), int(match.group(3)), match.group(4), match.group(5)


def handler(event):
    """Increment a semantic version string."""
    try:
        version = event.get("version")
        bump = event.get("bump", "patch")
        preid = event.get("preid", "alpha")
        if not version:
            return {"ok": False, "error": "version is required"}

        major, minor, patch, pre, build = parse_version(version)
        previous = version.lstrip("v")

        if bump == "major":
            new_version = f"{major + 1}.0.0"
        elif bump == "minor":
            new_version = f"{major}.{minor + 1}.0"
        elif bump == "patch":
            new_version = f"{major}.{minor}.{patch + 1}"
        elif bump == "premajor":
            new_version = f"{major + 1}.0.0-{preid}.0"
        elif bump == "preminor":
            new_version = f"{major}.{minor + 1}.0-{preid}.0"
        elif bump == "prepatch":
            new_version = f"{major}.{minor}.{patch + 1}-{preid}.0"
        elif bump == "prerelease":
            if pre and pre.startswith(preid):
                # Increment pre-release number
                parts = pre.split(".")
                if parts[-1].isdigit():
                    parts[-1] = str(int(parts[-1]) + 1)
                    new_version = f"{major}.{minor}.{patch}-{'.'.join(parts)}"
                else:
                    new_version = f"{major}.{minor}.{patch}-{pre}.1"
            else:
                new_version = f"{major}.{minor}.{patch}-{preid}.0"
        else:
            return {"ok": False, "error": f"Unknown bump type: {bump}"}

        return {"ok": True, "result": new_version, "previous": previous}
    except Exception as e:
        return {"ok": False, "error": str(e)}
