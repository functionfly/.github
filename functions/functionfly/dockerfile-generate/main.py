def handler(event):
    """
    Generate a Dockerfile based on project configuration and requirements.
    """
    try:
        language = event.get("language")
        framework = event.get("framework", "")
        ports = event.get("ports", [])
        env_vars = event.get("env_vars", {})

        if not language:
            return {"ok": False, "error": "language is required"}

        # Basic Dockerfile generation logic
        dockerfile_lines = []

        # Base image selection
        base_images = {
            "python": "python:3.12-slim",
            "node": "node:18-alpine",
            "go": "golang:1.21-alpine",
            "rust": "rust:1.70-slim",
            "java": "openjdk:17-slim",
            "php": "php:8.2-fpm-alpine",
            "ruby": "ruby:3.2-alpine"
        }

        base_image = base_images.get(language.lower(), f"{language}:latest")
        dockerfile_lines.append(f"FROM {base_image}")

        # Working directory
        dockerfile_lines.append("WORKDIR /app")

        # Copy requirements/package files first for better caching
        if language == "python":
            dockerfile_lines.append("COPY requirements.txt .")
            dockerfile_lines.append("RUN pip install --no-cache-dir -r requirements.txt")
        elif language == "node":
            dockerfile_lines.append("COPY package*.json .")
            dockerfile_lines.append("RUN npm ci --only=production")

        # Copy source code
        dockerfile_lines.append("COPY . .")

        # Environment variables
        for key, value in env_vars.items():
            dockerfile_lines.append(f"ENV {key}={value}")

        # Expose ports
        for port in ports:
            dockerfile_lines.append(f"EXPOSE {port}")

        # Default command
        if language == "python" and framework == "flask":
            dockerfile_lines.append("CMD [\"python\", \"app.py\"]")
        elif language == "node":
            dockerfile_lines.append("CMD [\"npm\", \"start\"]")
        elif language == "go":
            dockerfile_lines.append("CMD [\"./main\"]")
        else:
            dockerfile_lines.append("CMD [\"echo\", \"Container started\"]")

        dockerfile_content = "\n".join(dockerfile_lines)

        return {
            "ok": True,
            "result": "Dockerfile generated successfully",
            "dockerfile": dockerfile_content
        }
    except Exception as e:
        return {
            "ok": False,
            "error": str(e)
        }
