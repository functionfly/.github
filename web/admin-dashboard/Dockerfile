# syntax=docker/dockerfile:1.7
# ---------------------------------------------------------------------------
# Admin Dashboard — multi-stage build
# Stage 1: build the SPA (needs Node + Vite)
# Stage 2: serve via nginx (alpine, non-root, no source code)
# ---------------------------------------------------------------------------

# ---- Stage 1: build ------------------------------------------------------
FROM node:20-alpine AS builder

WORKDIR /app

# Install only the manifests first so dependency install is cached unless
# they actually change.
COPY package.json package-lock.json* ./
# `npm ci` fails if the lockfile is missing; fall back to `npm install`
# so contributors without a committed lockfile can still build locally.
RUN if [ -f package-lock.json ]; then npm ci; else npm install; fi

# Now copy the rest of the source. Vite's `define` resolves env at build,
# so build-time secrets are not present in the final image.
COPY . .

# Build-time env. `VITE_*` are baked into the bundle; the API URL is
# the only one that's safe to commit publicly.
ARG VITE_API_BASE_URL=https://api.functionfly.com
ENV VITE_API_BASE_URL=${VITE_API_BASE_URL}

RUN npm run build:vite

# Prune dev dependencies so they're not in the final stage's tarball.
# (Not strictly needed since the builder is discarded, but keeps layers small
#  for anyone who `docker cp`s out of an intermediate stage.)

# ---- Stage 2: runtime ----------------------------------------------------
FROM nginx:1.27-alpine AS runtime

# Strip the default nginx config.
RUN rm -f /etc/nginx/conf.d/default.conf

# Production nginx config (HSTS, CSP, etc).
COPY nginx.conf /etc/nginx/conf.d/default.conf

# Build artifact.
COPY --from=builder /app/dist /usr/share/nginx/html

# Run as the unprivileged `nginx` user that the upstream image ships with.
# Drop privileges before the listen socket is bound.
RUN chown -R nginx:nginx /usr/share/nginx/html /var/cache/nginx /var/log/nginx \
    && touch /var/run/nginx.pid && chown nginx:nginx /var/run/nginx.pid \
    && chmod -R u+rwX /usr/share/nginx/html
USER nginx

EXPOSE 80

# Healthcheck against the in-app /health endpoint served by nginx.
HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
  CMD wget --quiet --tries=1 --spider http://127.0.0.1/health || exit 1

CMD ["nginx", "-g", "daemon off;"]
