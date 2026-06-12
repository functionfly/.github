# FlyMind AI Service Dockerfile
# Security-hardened for production deployment

FROM python:3.12-slim-bookworm AS builder

# Set working directory
WORKDIR /app

# Install build dependencies
RUN apt-get update && apt-get install -y --no-install-recommends \
    build-essential \
    curl \
    && rm -rf /var/lib/apt/lists/*

# Copy project files
COPY pyproject.toml uv.lock ./

# Install Python dependencies with frozen lock file for reproducible builds
RUN pip install --no-cache-dir --upgrade pip && \
    pip install --no-cache-dir uv && \
    uv sync --frozen --no-dev

# Copy source code
COPY src/ ./src/

# Create non-root user with restricted permissions
RUN useradd --create-home --uid=1000 appuser && \
    mkdir -p /home/appuser/.cache && \
    chown -R appuser:appuser /app

# Final stage - runtime image
FROM python:3.12-slim-bookworm

# Set environment variables for security
ENV PYTHONDONTWRITEBYTECODE=1 \
    PYTHONUNBUFFERED=1 \
    PYTHONFAULTHANDLER=1 \
    PIP_NO_CACHE_DIR=1 \
    PIP_DISABLE_PIP_VERSION_CHECK=1 \
    # Security: prevent Python from writing .pyc files
    PYTHONDONTWRITEBYTECODE=1 \
    # Security: disable dangerous features
    PYTHONUNBUFFERED=1

# Create non-root user
RUN groupadd --gid=1000 appgroup && \
    useradd --uid=1000 --gid=appgroup --shell=/bin/false --create-home appuser

# Set working directory
WORKDIR /app

# Copy installed packages from builder
COPY --from=builder /usr/local/lib/python3.12/site-packages /usr/local/lib/python3.12/site-packages
COPY --from=builder /usr/local/bin /usr/local/bin

# Copy application code
COPY --chown=appuser:appgroup src/ ./src/

# Create tmp directory with restricted permissions
RUN mkdir -p /tmp && chmod 1777 /tmp

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 8081

# Health check with proper security
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD python -c "import urllib.request; urllib.request.urlopen('http://localhost:8081/health')"

# Run the application with security defaults
# Use --host 127.0.0.1 if accessible externally, 0.0.0.0 for Fly.io
CMD ["python", "-m", "uvicorn", "src.main:app", "--host", "0.0.0.0", "--port", "8081"]
