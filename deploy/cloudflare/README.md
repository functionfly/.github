# Cloudflare deploy helpers

- **cloudflare-tunnel.example.sh** – Example script to run a Cloudflare Tunnel (cloudflared). Use it to expose the orchestrator API (and optionally dashboard) through Cloudflare without opening ports. See [docs/CLOUDFLARE.md](../../docs/CLOUDFLARE.md#cloudflare-tunnel-optional).
- **Pages** – To deploy the dashboard or docs to Cloudflare Pages, connect the repo in the Cloudflare dashboard, set build output (e.g. `web/dashboard` → build → `dist`), and configure `VITE_API_URL`. Point DNS (e.g. `app`, `dashboard`, `docs`) to your Pages project. See [docs/CLOUDFLARE.md](../../docs/CLOUDFLARE.md#cloudflare-pages-optional).

All Cloudflare-related env vars, DNS, CDN, R2, and Workers are documented in **[docs/CLOUDFLARE.md](../../docs/CLOUDFLARE.md)**.
