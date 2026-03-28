from datetime import date


SECTION_TITLES = {
    "feat": "Features",
    "fix": "Bug Fixes",
    "perf": "Performance Improvements",
    "refactor": "Code Refactoring",
    "docs": "Documentation",
    "style": "Styles",
    "test": "Tests",
    "build": "Build System",
    "ci": "Continuous Integration",
    "chore": "Chores",
    "revert": "Reverts",
    "breaking": "BREAKING CHANGES",
}

DISPLAY_ORDER = ["feat", "fix", "perf", "refactor", "docs", "style", "test", "build", "ci", "chore"]


def handler(event):
    """Generate a CHANGELOG.md from conventional commits."""
    try:
        commits = event.get("commits")
        version = event.get("version")
        if not commits or not version:
            return {"ok": False, "error": "commits and version are required"}

        release_date = event.get("date", str(date.today()))
        repo_url = event.get("repo_url", "")

        # Group commits by type
        sections = {}
        breaking = []
        for commit in commits:
            ctype = commit.get("type", "chore").lower()
            scope = commit.get("scope", "")
            subject = commit.get("subject", "")
            breaking_change = commit.get("breaking", False)

            if breaking_change:
                breaking.append(f"- **{scope}:** {subject}" if scope else f"- {subject}")

            if ctype not in sections:
                sections[ctype] = []
            entry = f"- **{scope}:** {subject}" if scope else f"- {subject}"
            if commit.get("hash") and repo_url:
                entry += f" ([{commit['hash'][:7]}]({repo_url}/commit/{commit['hash']}))"
            elif commit.get("hash"):
                entry += f" ({commit['hash'][:7]})"
            sections[ctype].append(entry)

        lines = [f"## [{version}] - {release_date}", ""]

        if breaking:
            lines.append("### ⚠ BREAKING CHANGES")
            lines.extend(breaking)
            lines.append("")

        for ctype in DISPLAY_ORDER:
            if ctype in sections:
                title = SECTION_TITLES.get(ctype, ctype.title())
                lines.append(f"### {title}")
                lines.extend(sections[ctype])
                lines.append("")

        changelog = "\n".join(lines)
        return {"ok": True, "result": f"Changelog for {version} generated", "changelog": changelog}
    except Exception as e:
        return {"ok": False, "error": str(e)}
