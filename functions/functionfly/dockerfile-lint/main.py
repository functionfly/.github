import re


def handler(event):
    """Lint a Dockerfile for common issues."""
    try:
        content = event.get("content")
        if not content:
            return {"ok": False, "error": "content is required"}

        issues = []
        warnings = []
        lines = content.splitlines()

        has_from = False
        run_count = 0
        last_run_line = -1
        has_user = False
        has_healthcheck = False

        for i, line in enumerate(lines, 1):
            stripped = line.strip()
            if not stripped or stripped.startswith("#"):
                continue

            parts = stripped.split(None, 1)
            if not parts:
                continue
            instruction = parts[0].upper()
            args = parts[1] if len(parts) > 1 else ""

            if instruction == "FROM":
                has_from = True
                # Check for untagged or latest
                if ":" not in args and "@" not in args:
                    warnings.append(f"Line {i}: Use specific image tag (e.g. ubuntu:22.04)")
                elif ":latest" in args:
                    warnings.append(f"Line {i}: Avoid using ':latest' tag for reproducibility")
                # Check for root user
                if "root" in args.lower():
                    warnings.append(f"Line {i}: Consider using a non-root base image")

            elif instruction == "RUN":
                run_count += 1
                if last_run_line > 0 and i == last_run_line + 1:
                    warnings.append(f"Line {i}: Consider combining consecutive RUN commands with && to reduce layers")
                last_run_line = i
                # Check for apt-get without cleanup
                if "apt-get install" in args and "rm -rf /var/lib/apt/lists" not in args:
                    warnings.append(f"Line {i}: Clean apt cache after install: && rm -rf /var/lib/apt/lists/*")
                # Check for sudo
                if "sudo" in args:
                    warnings.append(f"Line {i}: Avoid using sudo in Dockerfile")

            elif instruction == "ADD":
                if not args.startswith("http"):
                    warnings.append(f"Line {i}: Prefer COPY over ADD for local files")

            elif instruction == "USER":
                has_user = True
                if args.lower() == "root":
                    issues.append(f"Line {i}: Running as root is a security risk")

            elif instruction == "HEALTHCHECK":
                has_healthcheck = True

            elif instruction == "EXPOSE":
                try:
                    port = int(args.split("/")[0])
                    if port < 1 or port > 65535:
                        issues.append(f"Line {i}: Invalid port number: {port}")
                except ValueError:
                    pass

        if not has_from:
            issues.append("Missing FROM instruction")
        if not has_user:
            warnings.append("Consider adding USER instruction to run as non-root")
        if not has_healthcheck:
            warnings.append("Consider adding HEALTHCHECK instruction")

        return {
            "ok": True,
            "valid": len(issues) == 0,
            "issues": issues,
            "warnings": warnings,
            "issue_count": len(issues),
            "warning_count": len(warnings),
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
