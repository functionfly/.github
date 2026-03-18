# Build and run Paperclip from official repo (for internal control-plane deployment).
# See deploy/paperclip/README.md.
FROM node:20-bookworm-slim AS base
RUN apt-get update \
  && apt-get install -y --no-install-recommends ca-certificates curl git \
  && rm -rf /var/lib/apt/lists/*
RUN corepack enable && corepack prepare pnpm@latest --activate

WORKDIR /app
ARG PAPERCLIP_REF=master
RUN git clone --depth 1 https://github.com/paperclipai/paperclip.git . \
  && git fetch --depth 1 origin "${PAPERCLIP_REF}" 2>/dev/null && git checkout "${PAPERCLIP_REF}" 2>/dev/null || true

# Install with frozen lockfile when present; otherwise install (clone includes lockfile on main)
RUN pnpm install --frozen-lockfile 2>/dev/null || pnpm install

# Build all workspace packages in dependency order (server depends on db, shared, adapters, etc.)
RUN pnpm build
RUN test -f server/dist/index.js || (echo "ERROR: server build missing" && exit 1)

RUN mkdir -p /paperclip && chown node:node /paperclip

# Entrypoint: ensure volume-mounted /paperclip is writable by node; create instance dirs so embedded Postgres can init
RUN echo '#!/bin/sh\nset -e\nmkdir -p /paperclip/instances/default/db /paperclip/instances/default/logs /paperclip/instances/default/data\nchown -R node:node /paperclip\nexec gosu node "$@"' > /entrypoint.sh \
  && chmod +x /entrypoint.sh
RUN apt-get update && apt-get install -y --no-install-recommends gosu && rm -rf /var/lib/apt/lists/*

ENV NODE_ENV=production \
  HOST=0.0.0.0 \
  PORT=3100 \
  SERVE_UI=true \
  PAPERCLIP_HOME=/paperclip \
  PAPERCLIP_INSTANCE_ID=default \
  PAPERCLIP_DEPLOYMENT_MODE=authenticated \
  PAPERCLIP_DEPLOYMENT_EXPOSURE=private

VOLUME ["/paperclip"]
EXPOSE 3100

ENTRYPOINT ["/entrypoint.sh"]
CMD ["node", "--import", "./server/node_modules/tsx/dist/loader.mjs", "server/dist/index.js"]
