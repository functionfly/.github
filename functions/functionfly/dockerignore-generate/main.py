PATTERNS = {
    "common": [
        "# Common",
        ".git/",
        ".gitignore",
        ".DS_Store",
        "*.log",
        "README.md",
        "LICENSE",
        ".env",
        ".env.*",
        "!.env.example",
    ],
    "python": [
        "# Python",
        "__pycache__/",
        "*.pyc",
        "*.pyo",
        "*.pyd",
        ".Python",
        ".venv/",
        "venv/",
        "env/",
        "*.egg-info/",
        "dist/",
        "build/",
        ".pytest_cache/",
        ".mypy_cache/",
        ".coverage",
        "htmlcov/",
    ],
    "node": [
        "# Node.js",
        "node_modules/",
        "npm-debug.log*",
        "yarn-debug.log*",
        "yarn-error.log*",
        ".npm",
        ".yarn",
        "dist/",
        "build/",
        ".next/",
        ".nuxt/",
        "coverage/",
    ],
    "go": [
        "# Go",
        "*.exe",
        "*.exe~",
        "*.dll",
        "*.so",
        "*.dylib",
        "*.test",
        "*.out",
        "vendor/",
    ],
    "java": [
        "# Java",
        "*.class",
        "*.jar",
        "*.war",
        "*.ear",
        "target/",
        ".gradle/",
        "build/",
        ".mvn/",
    ],
    "rust": [
        "# Rust",
        "target/",
        "Cargo.lock",
        "*.rs.bk",
    ],
}


def handler(event):
    """Generate a .dockerignore file."""
    try:
        language = event.get("language", "").lower()
        extras = event.get("extras", [])

        lines = PATTERNS.get("common", [])
        lang_patterns = PATTERNS.get(language, [])
        if lang_patterns:
            lines = lines + [""] + lang_patterns
        else:
            lines = lines + [f"# {language.title()} (no specific patterns)"]

        if extras:
            lines = lines + ["", "# Custom"] + extras

        dockerignore = "\n".join(lines) + "\n"
        return {"ok": True, "result": ".dockerignore generated", "dockerignore": dockerignore}
    except Exception as e:
        return {"ok": False, "error": str(e)}
