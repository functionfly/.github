import re


SEMVER_RE = re.compile(
    r'^(?P<major>0|[1-9]\d*)\.(?P<minor>0|[1-9]\d*)\.(?P<patch>0|[1-9]\d*)'
    r'(?:-(?P<prerelease>[a-zA-Z0-9\-\.]+))?'
    r'(?:\+(?P<build>[a-zA-Z0-9\-\.]+))?$'
)


def handler(event):
    """Parse a semantic version string into its components."""
    try:
        version = event.get("version")
        if not version:
            return {"ok": False, "error": "version is required"}

        version = str(version).strip().lstrip("v")
        match = SEMVER_RE.match(version)
        if not match:
            return {"ok": True, "valid": False, "error": f"Invalid semver: {version}"}

        return {
            "ok": True,
            "valid": True,
            "result": {
                "major": int(match.group("major")),
                "minor": int(match.group("minor")),
                "patch": int(match.group("patch")),
                "prerelease": match.group("prerelease"),
                "build": match.group("build"),
                "version": version,
            },
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
