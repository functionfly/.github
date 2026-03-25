# Admin Dashboard Cloudflare Security Configuration

Step-by-step guide for hardening `admin.domain.com` (e.g., `admin.functionfly.com`) in Cloudflare Dashboard.

---

## Prerequisites

- Domain added to Cloudflare (Zone provisioned)
- DNS record for `admin.domain.com` created and proxied (orange cloud)
- Cloudflare Pro or higher plan (required for WAF, Rate Limiting, Bot Management)

---

## 1. SSL/TLS Configuration

**Goal:** Enforce TLS 1.3, disable older protocols.

### Steps

1. Navigate to **SSL/TLS** → **Edge Certificates**
2. Set **Minimum TLS Version** to **TLS 1.3**
3. Disable older versions:
   - Go to **SSL/TLS** → **Edge Certificates**
   - Ensure **TLS 1.2** is NOT enabled as minimum (keep at 1.3 minimum)
   - Under **Cipher Suites**, use Cloudflare default (-modern)
4. Set **Automatic HTTPS Rewrites** to **On**
5. Set **Opportunistic Encryption** to **On**
6. Set **Universal SSL** status: ensure a valid certificate is issued

**Verification:** Use [SSLLabs](https://www.ssllabs.com/ssltest/) to check `https://admin.domain.com`

---

## 2. WAF Settings for admin.domain.com

**Goal:** Enable and tune the Cloudflare WAF specifically for admin endpoints.

### Steps

1. Navigate to **Security** → **WAF**
2. Under **Managed Rules**, click **Deploy** → **Deploy a managed ruleset**
3. Select **OWASP ModSecurity Core Rule Set** (see Section 6 below)
4. Configure **Custom Filter** for admin subdomain:
   - Add a rule: `AND` → `Field: Hostname` → `Operator: equals` → `Value: admin.domain.com`
5. Set **Action** to `Simulate` initially, then `Block` after testing
6. Enable **Cloudflare Managed Rules**:
   - Go to **Security** → **WAF** → **Managed Rules**
   - Add a rule: `Field: Hostname` `equals` `admin.domain.com`
   - Enable all Cloudflare managed rulesets (PHP, WordPress, etc. if applicable)
   - Set paranoia level to **High** for admin

---

## 3. Bot Management Settings

**Goal:** Identify and block automated threats to admin endpoints.

### Steps

1. Navigate to **Security** → **Bots**
2. Enable **Bot Fight Mode** (if not already):
   - Toggle **Bot Fight Mode** to **On**
3. For admin.domain.com, enable **Bot Management** (Enterprise):
   - If on Pro/Team plan: enable **Bot Score** via **Security** → **Settings** → **Bot Score**
4. Create a **Custom Rule** for admin:
   - Go to **Security** → **WAF** → **Custom Rules**
   - Click **Create rule**:
     - **Name:** `admin-block-bots`
     - **Field:** `Host`
     - **Operator:** `equals`
     - **Value:** `admin.domain.com`
     - **AND**
     - **Field:** `Bot Score`
     - **Operator:** `less than`
     - **Value:** `30`
     - **Action:** `Block`
5. Verify legitimate admin users are not blocked (add Page Rules exceptions if needed)

---

## 4. Rate limiting (API hostname)

**Goal:** Cap abuse on the **orchestrator** (`api.functionfly.com`). The admin **static** app lives on `admin.functionfly.com`, but JSON calls go to the API.

**Navigation (names vary by plan):** **Security** → **Security rules** → **Rate limiting rules**, or **Security** → **WAF** → **Rate limiting rules**. (Older UI: **Security** → **Tools** → **Rate limiting**.)

### Rule 1 — `/v1/admin/*`

| Field | Value |
|--------|--------|
| **Name** | `api-admin-prefix-limit` |
| **When** | Hostname equals `api.functionfly.com` **and** URI Path starts with `/v1/admin` |
| **Counting** | Per source IP (or per IP + JA3 fingerprint if available on your plan) |
| **Threshold** | **20 requests / 10 seconds** |
| **Action** | Block or Managed Challenge (start with challenge if you see false positives) |
| **Duration** | e.g. 1 hour block (tune to taste) |

### Rule 2 — login endpoints (stricter)

Admin login uses **`POST /auth/login`** and **`POST /v1/auth/login`** (not under `/v1/admin`). There is no `/v1/admin/login` in the current API.

| Field | Value |
|--------|--------|
| **Name** | `api-auth-login-limit` |
| **When** | Hostname equals `api.functionfly.com` **and** (URI Path equals `/auth/login` **or** URI Path equals `/v1/auth/login`) **and** Method is `POST` |
| **Threshold** | **5 requests / 60 seconds** per IP |
| **Action** | Block (short duration) or Managed Challenge |

### Office / allowlisted IPs (exception)

1. Create an **IP list** (**Manage account** → **Configurations** → **Lists**, or **Security** → **WAF** → **Tools** → **IP lists**) with your office egress IPs (`/32` or `/29` as appropriate).
2. On **each** rate limit rule, add an **exception** (or a **skip** rule with higher priority): **if** Source IP is in list **Office allowlist** → **skip** rate limiting / **bypass**.
3. Alternatively, use a **WAF custom rule** evaluated first: `Source IP` in `Office allowlist` **and** `Hostname` equals `api.functionfly.com` → **Skip** → remaining rules.

**Note:** Cloudflare rate limiting is a **second layer** on top of application rate limits (`internal/api/middleware`). Keep thresholds aligned so legitimate admins are not throttled during bulk UI actions.

### When the dashboard shows “1 / 1 used” (rule quota)

Zones often have a **limited number of Rate limiting rules** (the counter shows **used / included**). If you already have a rule such as **FFly** matching `URI Path` `*/execute*` (or similar) with **Block**, that rule **consumes one slot**—typically to protect **function execution** abuse.

**If you cannot add Rule 1 and Rule 2 above:**

| Approach | When to use |
|----------|-------------|
| **Keep** `/execute` at the edge | High priority; execution endpoints are expensive. Leave **FFly** as-is. |
| **Rely on the API** for `/v1/admin` | The Go orchestrator already applies admin rate limits (`internal/api/middleware`). Cloudflare rules are optional reinforcement. |
| **Upgrade** Cloudflare plan | More rate-limit rule slots (and/or **Advanced Rate Limiting** on Enterprise). |
| **Combine** (advanced) | Only if the UI supports a **single** rule with multiple expressions or phases; **different thresholds** (20/10s vs 5/60s) usually need **separate** rules or app-side limits. |
| **WAF Custom rules** | Use for **challenge/block** by path without the same “rate limit rule” quota in some setups—check your plan; behavior differs from **Rate limiting rules**. |

**“Go to web application exploits settings”** in the dashboard links to related **WAF** / managed rules—useful for exploit signatures, but it does **not** increase the rate-limit rule count.

**Practical default:** keep **FFly** on `/execute`, and depend on **orchestrator** limits for `/v1/admin` and login until you add quota or Advanced Rate Limiting.

---

## 5. IP reputation & IP Access Rules

**Goal:** Reduce noise from known abusive networks, VPNs, and proxies hitting `api.functionfly.com` / `admin.functionfly.com`.

### IP Access Rules (allow/block)

**Navigation:** **Security** → **WAF** → **Tools** → **IP Access Rules** (or **Security** → **Overview** → **Tools**).

1. **Allowlist** office/static IPs if you use them for break-glass access (optional; coordinate with rate-limit exceptions).
2. **Block** known-bad ranges only when you have high confidence (avoid blocking mobile carrier NAT).

### WAF custom rules (threat intelligence)

1. **Security** → **Security rules** → **Custom rules** (or **WAF** → **Custom rules**).
2. Example — high threat to API:
   - **When:** Hostname equals `api.functionfly.com` **and** **Threat score** greater than `10` (tune after observing **Security** → **Events**).
   - **Action:** Managed Challenge or Block.
3. Use **Cloudflare threat intelligence** fields available on your plan (e.g. **Bot score**, **ASN reputation**, **Verified bots**).
4. **VPN / proxy:** If your plan exposes **proxy** or **VPN** signals, add a separate rule with **Log** first, then **Challenge** or **Block** after validation (many legitimate admins use VPNs—test carefully).

### Known spam sources

Prefer **managed rulesets** and **leaked credential checks** where available; block “spam” sources via **threat score** and **custom lists** rather than broad country blocks unless policy requires it.

---

## 6. OWASP Ruleset Configuration

**Goal:** Enable OWASP protection with tuned sensitivity.

### Steps

1. Navigate to **Security** → **WAF** → **Managed Rules**
2. Click **Deploy** → **Deploy a managed ruleset**
3. Select **OWASP ModSecurity Core Rule Set**
4. Configure:
   - **Paranoia Level:** `High` (recommended for admin)
   - **Sensitivity:** `Medium` (adjust after testing)
   - **Score Threshold:** `60` (blocks at 60+ score)
5. Apply to `admin.domain.com`:
   - Add a **Zone** filter: `Host` `equals` `admin.domain.com`
6. For SQLi and XSS detection:
   - Under **OWASP Rules**, ensure these are set to `Block`:
     - `SQL Injection` (rule ID 942100 series)
     - `Cross-Site Scripting` (rule ID 941100 series)
     - `Local File Inclusion` (rule ID 932100 series)
7. **Testing Phase:**
   - Set to `Simulate` initially
   - Monitor **Security Events** → **Overview**
   - After 24-48 hours with no false positives, switch to `Block`

---

## 7. DDoS protection (L7)

**Goal:** Keep HTTP DDoS mitigation on with sensible sensitivity; use overrides only when needed.

### Steps

1. Navigate to **Security** → **DDoS** (or **Security** → **Settings** → **DDoS** depending on UI).
2. **HTTP DDoS / L7 protection:** **On** (default on proxied zones).
3. **Sensitivity:** **Medium** for steady state; raise to **High** only during an incident (expect more challenges).
4. **Manual override:** Under **DDoS** → **Overrides** (or **HTTP DDoS rule** advanced settings), add a temporary override for `api.functionfly.com` / `admin.functionfly.com` if traffic spikes—remove after the event to avoid excess friction.
5. **Under Attack mode:** Use **Security** → **Settings** → **Under Attack Mode** only for active attacks (adds JS challenge to visitors).

**Note:** Static Pages on `admin.functionfly.com` benefit from edge DDoS; the API on Fly still relies on origin capacity—Cloudflare protects the **proxied** hostname only.

---

## 8. Security headers (response)

**Goal:** Send consistent security headers for HTML and APIs at the edge.

**Navigation:** **Rules** → **Overview** → **Transform Rules** → **Modify Response Header** (modern replacement for legacy Page Rules header overrides). Match hostname `admin.functionfly.com` (and optionally `api.functionfly.com` if the origin does not set headers).

Add or merge these (avoid duplicating conflicting headers):

| Header | Suggested value |
|--------|-----------------|
| `Strict-Transport-Security` | `max-age=63072000; includeSubDomains; preload` |
| `X-Frame-Options` | `DENY` |
| `X-Content-Type-Options` | `nosniff` |
| `Referrer-Policy` | `strict-origin-when-cross-origin` |

The admin SPA already ships `_headers` from Cloudflare Pages (`web/admin-dashboard/public/_headers`); use Transform Rules only to align or override what Pages sends. For `api.functionfly.com`, prefer headers from the Go orchestrator where possible to avoid conflicting `Content-Security-Policy` with JSON responses.

### 8.1 Restrict by country (if needed)

1. Go to **Security** → **WAF** → **Custom Rules**
2. Create rule:
   - **Name:** `admin-allow-countries`
   - **Field:** `Host`
   - **Value:** `admin.domain.com`
   - **Action:** `Allow`
   - Add condition: `Country` `is in` `US,GB,...` (whitelist allowed countries)

### 8.2 Firewall / IP Access Rule (break-glass allowlist)

1. Navigate to **Security** → **Tools** → **IP Access Rules**
2. Add your **office/static IP**:
   - **IP:** `YOUR_IP/32`
   - **Action:** `Allow`
   - **Scope:** `admin.domain.com`
3. Block all other traffic to admin subdomain:
   - Add rule: `Host` `equals` `admin.domain.com`
   - **Action:** `Block`
   - **Scope:** `All`

---

## 9. Quick verification checklist

After configuring, verify each setting:

| Area | Check |
|------|--------|
| SSL/TLS | [SSLLabs](https://www.ssllabs.com/ssltest/) on `admin.functionfly.com` / `api.functionfly.com` |
| WAF | Malformed query to admin static host; review **Security** → **Events** |
| Bot Management | **Security** → **Bots** shows scores for sampled requests |
| Rate limiting | Exceed Rule 1/2 thresholds from a non-allowlisted IP → `429` or block; allowlisted office IP → still succeeds |
| IP reputation | **Security** → **Events** shows matches for threat / challenge rules |
| OWASP | `Simulate` then `Block`; test payloads logged before enforcement |
| DDoS | L7 protection **On**, sensitivity **Medium**; overrides documented for incidents |
| Response headers | See commands below |

**Headers (admin Pages):**

```bash
curl -sI https://admin.functionfly.com/ | grep -iE 'strict-transport|x-frame-options|x-content-type|referrer-policy'
```

**API (CORS preflight sample):**

```bash
curl -sI -X OPTIONS 'https://api.functionfly.com/v1/admin/csrf' \
  -H 'Origin: https://admin.functionfly.com' \
  -H 'Access-Control-Request-Method: GET' \
  | grep -i access-control
```

Expect `Access-Control-Allow-Origin: https://admin.functionfly.com` when `CORS_ALLOWED_ORIGINS` is set on the orchestrator.

---

## 10. Monitoring & Tuning

1. **Security Events:** Review daily under **Security** → **Events**
2. **Analytics:** Monitor traffic under **Analytics & Logs**
3. **Alerts:** Set up alerts:
   - **Security** → **Alerts** → **New alert** → e.g., "High traffic volume", "WAF triggered"
4. **Logpush:** For advanced, configure Logpush to send to a SIEM

---

## Orchestrator API: CORS for `admin.functionfly.com`

The admin SPA calls `https://api.functionfly.com` from the browser. The orchestrator **must** allow that origin or preflight requests fail.

Set on the API (Fly secrets, Kubernetes env, or `.env`):

```bash
CORS_ALLOWED_ORIGINS=https://functionfly.com,https://www.functionfly.com,https://app.functionfly.com,https://admin.functionfly.com
```

Include every deployed frontend origin (comma-separated, no spaces). After changing secrets, redeploy or restart the orchestrator. See `docs/FLY_DEPLOYMENT.md` and `.fly/set-secrets-from-neon.sh`.

---

## References

- [Cloudflare WAF Documentation](https://developers.cloudflare.com/waf/)
- [Cloudflare Rate Limiting](https://developers.cloudflare.com/rate-limiting/)
- [Cloudflare Bot Management](https://developers.cloudflare.com/bot-management/)
- [OWASP Ruleset](https://developers.cloudflare.com/waf/reference/modify-core-ruleset/)
- [SSL/TLS Settings](https://developers.cloudflare.com/ssl/get-started/)
- [DDoS Protection](https://developers.cloudflare.com/ddos/)
