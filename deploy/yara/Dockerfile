FROM python:3.11-slim

# Install curl for health checks
RUN apt-get update && apt-get install -y curl && rm -rf /var/lib/apt/lists/*

# Set working directory
WORKDIR /app

# Copy requirements and install Python dependencies
COPY ./deploy/yara/requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Copy application code
COPY ./deploy/yara/yara_service.py .

# Expose port
EXPOSE 8080

# Run the service
CMD ["python", "yara_service.py"]