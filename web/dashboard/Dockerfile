# Build stage
FROM node:22-alpine AS builder

WORKDIR /app

# Copy package files
COPY package.json package-lock.json ./

# Install dependencies
RUN npm ci --only=production

# Copy source code
COPY . .

# Build the application
RUN npm run build

# Runtime stage
FROM nginx:1.25-alpine

WORKDIR /usr/share/nginx/html

# Remove default nginx configuration
RUN rm /etc/nginx/conf.d/default.conf

# Copy custom nginx configuration
COPY nginx.conf /etc/nginx/conf.d/

# Copy built application from builder
COPY --from=builder /app/dist .

# Create non-root user
RUN addgroup -g 1000 -S functionfly && \
    adduser -u 1000 -S functionfly -G functionfly

# Set ownership
RUN chown -R functionfly:functionfly /usr/share/nginx/html && \
    chown -R functionfly:functionfly /var/cache/nginx && \
    chown -R functionfly:functionfly /var/log/nginx && \
    chown -R functionfly:functionfly /etc/nginx/conf.d && \
    touch /var/run/nginx.pid && \
    chown -R functionfly:functionfly /var/run/nginx.pid

# Switch to non-root user
USER functionfly

# Expose port
EXPOSE 80

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=30s --retries=3 \
    CMD curl -f http://localhost:80/health || exit 1

# Run nginx
CMD ["nginx", "-g", "daemon off;"]
