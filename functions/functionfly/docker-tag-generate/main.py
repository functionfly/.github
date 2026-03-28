import re


def handler(event):
    """Generate Docker image tags from a version string."""
    try:
        image = event.get("image")
        version = event.get("version")
        if not image or not version:
            return {"ok": False, "error": "image and version are required"}

        registry = event.get("registry", "").rstrip("/")
        branch = event.get("branch")
        commit_sha = event.get("commit_sha")
        include_latest = event.get("latest", True)

        version = str(version).strip().lstrip("v")
        base = f"{registry}/{image}" if registry else image

        tags = []

        # Full semver tag
        tags.append(f"{base}:{version}")

        # Parse semver for partial tags
        match = re.match(r'^(\d+)\.(\d+)\.(\d+)', version)
        if match:
            major, minor, patch = match.group(1), match.group(2), match.group(3)
            # Only add partial tags for stable releases (no prerelease)
            if not re.search(r'-', version):
                tags.append(f"{base}:{major}.{minor}")
                tags.append(f"{base}:{major}")

        if branch:
            # Sanitize branch name for Docker tag
            safe_branch = re.sub(r'[^a-zA-Z0-9._-]', '-', branch)
            tags.append(f"{base}:{safe_branch}")

        if commit_sha:
            tags.append(f"{base}:{commit_sha[:7]}")

        if include_latest and not re.search(r'-', version):
            tags.append(f"{base}:latest")

        return {"ok": True, "result": tags, "tags": tags}
    except Exception as e:
        return {"ok": False, "error": str(e)}
