import re


def parse_version(v):
    v = str(v).strip().lstrip("v")
    match = re.match(r'^(\d+)\.(\d+)\.(\d+)', v)
    if not match:
        return 0, 0, 0
    return int(match.group(1)), int(match.group(2)), int(match.group(3))


def handler(event):
    """Analyze commits to determine next semantic version."""
    try:
        commits = event.get("commits")
        if not commits:
            return {"ok": False, "error": "commits is required"}

        current_version = event.get("current_version", "0.0.0")
        major, minor, patch = parse_version(current_version)

        has_breaking = False
        has_feat = False
        has_fix = False

        for commit in commits:
            ctype = commit.get("type", "").lower()
            breaking = commit.get("breaking", False) or "BREAKING CHANGE" in commit.get("body", "")

            if breaking:
                has_breaking = True
            elif ctype == "feat":
                has_feat = True
            elif ctype in ("fix", "perf"):
                has_fix = True

        if has_breaking:
            bump_type = "major"
            next_version = f"{major + 1}.0.0"
        elif has_feat:
            bump_type = "minor"
            next_version = f"{major}.{minor + 1}.0"
        elif has_fix:
            bump_type = "patch"
            next_version = f"{major}.{minor}.{patch + 1}"
        else:
            bump_type = "none"
            next_version = current_version.lstrip("v")

        return {
            "ok": True,
            "bump_type": bump_type,
            "next_version": next_version,
            "current_version": current_version.lstrip("v"),
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
