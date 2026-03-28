PATTERNS = {
    "python": [
        "# Python",
        "__pycache__/",
        "*.py[cod]",
        "*$py.class",
        "*.so",
        ".Python",
        "build/",
        "develop-eggs/",
        "dist/",
        "downloads/",
        "eggs/",
        ".eggs/",
        "lib/",
        "lib64/",
        "parts/",
        "sdist/",
        "var/",
        "wheels/",
        "*.egg-info/",
        ".installed.cfg",
        "*.egg",
        ".venv",
        "venv/",
        "ENV/",
        ".pytest_cache/",
        ".mypy_cache/",
        ".coverage",
        "htmlcov/",
        ".tox/",
    ],
    "node": [
        "# Node.js",
        "node_modules/",
        "npm-debug.log*",
        "yarn-debug.log*",
        "yarn-error.log*",
        ".pnpm-debug.log*",
        ".npm",
        ".yarn/cache",
        ".yarn/unplugged",
        ".yarn/build-state.yml",
        ".yarn/install-state.gz",
        ".pnp.*",
        "dist/",
        "build/",
        ".next/",
        ".nuxt/",
        ".output/",
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
        "go.work",
    ],
    "java": [
        "# Java",
        "*.class",
        "*.log",
        "*.ctxt",
        ".mtj.tmp/",
        "*.jar",
        "*.war",
        "*.nar",
        "*.ear",
        "*.zip",
        "*.tar.gz",
        "*.rar",
        "hs_err_pid*",
        "replay_pid*",
        "target/",
        ".gradle/",
        "build/",
    ],
    "rust": [
        "# Rust",
        "/target",
        "**/*.rs.bk",
        "Cargo.lock",
    ],
    "vscode": [
        "# VSCode",
        ".vscode/*",
        "!.vscode/settings.json",
        "!.vscode/tasks.json",
        "!.vscode/launch.json",
        "!.vscode/extensions.json",
        "*.code-workspace",
        ".history/",
    ],
    "jetbrains": [
        "# JetBrains IDEs",
        ".idea/",
        "*.iws",
        "*.iml",
        "*.ipr",
        "out/",
    ],
    "vim": [
        "# Vim",
        "[._]*.s[a-v][a-z]",
        "[._]*.sw[a-p]",
        "[._]s[a-rt-v][a-z]",
        "[._]ss[a-gi-z]",
        "[._]sw[a-p]",
        "Session.vim",
        "Sessionx.vim",
        ".netrwhist",
        "*~",
        "tags",
    ],
    "macos": [
        "# macOS",
        ".DS_Store",
        ".AppleDouble",
        ".LSOverride",
        "Icon",
        "._*",
        ".DocumentRevisions-V100",
        ".fseventsd",
        ".Spotlight-V100",
        ".TemporaryItems",
        ".Trashes",
        ".VolumeIcon.icns",
        ".com.apple.timemachine.donotpresent",
    ],
    "windows": [
        "# Windows",
        "Thumbs.db",
        "Thumbs.db:encryptable",
        "ehthumbs.db",
        "ehthumbs_vista.db",
        "*.stackdump",
        "[Dd]esktop.ini",
        "$RECYCLE.BIN/",
        "*.cab",
        "*.msi",
        "*.msix",
        "*.msm",
        "*.msp",
        "*.lnk",
    ],
    "linux": [
        "# Linux",
        "*~",
        ".fuse_hidden*",
        ".directory",
        ".Trash-*",
        ".nfs*",
    ],
}


def handler(event):
    """Generate a .gitignore file."""
    try:
        language = event.get("language", "").lower()
        framework = event.get("framework", "").lower()
        ide = event.get("ide", "").lower()
        os_type = event.get("os", "").lower()
        extras = event.get("extras", [])

        sections = []

        for key in [language, framework, ide, os_type]:
            if key and key in PATTERNS:
                sections.append("\n".join(PATTERNS[key]))

        if extras:
            sections.append("# Custom\n" + "\n".join(extras))

        if not sections:
            return {"ok": False, "error": f"No patterns found for language: {language}"}

        gitignore = "\n\n".join(sections) + "\n"
        return {"ok": True, "result": ".gitignore generated", "gitignore": gitignore}
    except Exception as e:
        return {"ok": False, "error": str(e)}
