# Admin authentication — production setup

Checklist for items called out in auth reviews: OAuth2, MFA, SAML, Cloudflare Access, IP allowlist. **Code paths exist**; most work is **configuration** and (for strict MFA-at-login) a **small backend change**.

---

## 1. OAuth2 (Google / GitHub)

**Backend** (`internal/auth/oauth.go`) registers providers when env vars are set:

| Variable | Purpose |
|----------|---------|
| `GOOGLE_CLIENT_ID` | Google OAuth client ID |
| `GOOGLE_CLIENT_SECRET` | Google OAuth secret |
| `GITHUB_CLIENT_ID` | GitHub OAuth app client ID |
| `GITHUB_CLIENT_SECRET` | GitHub OAuth secret |
| `BASE_URL` | Must be the public API origin, e.g. `https://api.functionfly.com` (used to build redirect URLs) |

**Redirect URLs to register in each provider:**

- Google Cloud Console / GitHub OAuth App → Authorized redirect URIs:
  - `https://api.functionfly.com/v1/auth/oauth/google/callback`
  - `https://api.functionfly.com/v1/auth/oauth/github/callback`

**API discovery (for the admin UI):**

- `GET https://api.functionfly.com/v1/auth/oauth/providers` — lists configured providers.
- `GET https://api.functionfly.com/v1/auth/oauth/url?provider=google` (and optional `redirect_uri` for CLI) — returns the authorization URL.

**Deploy:** Set secrets on Fly (or your host), e.g.:

```bash
fly secrets set GOOGLE_CLIENT_ID=... GOOGLE_CLIENT_SECRET=... GITHUB_CLIENT_ID=... GITHUB_CLIENT_SECRET=... BASE_URL=https://api.functionfly.com --app functionfly-control
```

**Admin dashboard:** Add buttons that call `GET /v1/auth/oauth/url?provider=...` and redirect the browser to the returned URL; after callback, the user should land with a session/token per your existing OAuth callback flow. (Wire to the same patterns as the main dashboard if it already has OAuth.)

**Microsoft (Entra / Azure AD):** Not wired in `initOAuthProviders` today. Options: **SAML** (see §3), add an Azure AD OAuth2 config in code, or use **Cloudflare Access** with Azure AD as the IdP (§4).

---

## 2. MFA “required” vs enforced at login

**Today:**

- `MFAService.IsMFARequired` treats **`admin`** and **`super_admin`** as MFA-required roles (`internal/auth/mfa.go`).
- `AuthService.Login` (`internal/auth/user_auth.go`) **does not** stop after password success to demand MFA — it issues JWT immediately if the password is valid.

So **“MFA required” is reflected in status APIs**, not as a hard gate on `POST /v1/auth/login`.

**Options:**

| Approach | Effort | Notes |
|----------|--------|--------|
| **Cloudflare Access** (recommended for fast production) | Config only | IdP + MFA at the edge **before** the SPA loads; no change to `HandleLogin`. |
| **Enforce in API** | Code change | After password OK, if `IsMFARequired(user.ID)` and `!user.MFAEnabled`, return **403** with a clear code (e.g. `mfa_setup_required`); if MFA enabled, return **401** `mfa_verification_required` until TOTP step completes (extend login flow or add `POST /v1/auth/mfa/verify-login`). |
| **Policy** | Ops | Require MFA enrollment in runbooks; use admin UI to verify `mfa_enabled` before granting access. |

**Quick win (API):** Add the check **after** successful password verification and **before** JWT issuance — align with your product’s desired UX (block vs step-up).

---

## 3. SAML SSO

**Handler code** exists in `internal/api/handlers/auth/saml.go` (metadata, login redirect, ACS, config). **Verify** that your deployment **registers** these handlers on `Server` / `routes.go` — a repo-wide search shows they may not be wired yet. If SAML routes are absent, use **Cloudflare Access** with your IdP (§4) until API SAML is fully integrated.

When routes are live, IdP configuration typically needs:

- **SP metadata URL** (entity ID, ACS URL) from FunctionFly for each `tenant_id`.
- **ACS** posting SAML responses to the URL your routes expose.

Store SAML config per tenant via the handler’s config APIs / DB as implemented in `auth.SAMLService`.

---

## 4. Cloudflare Access (Zero Trust) — recommended

Adds SSO + MFA + device posture **in front of** `https://admin.functionfly.com` without changing Go login code.

1. **Cloudflare Zero Trust** dashboard → **Access** → **Applications** → **Add an application**.
2. **Type:** Self-hosted (or SaaS if applicable).
3. **Application domain:** `admin.functionfly.com` (same hostname as Pages).
4. **Identity providers:** Add **Google Workspace**, **Azure AD**, **Okta**, or **One-time PIN** as needed.
5. **Policy:** e.g. **Include** → Emails ending in `@yourcompany.com`, or **SAML groups** / **Azure groups** → **Allow**; optional **Require** MFA in the Access policy.
6. **Bypass for automation:** Use **Service Auth** or narrow IP allowlists only if required; avoid wide bypass.

**DNS:** `admin.functionfly.com` must be proxied through Cloudflare so Access runs at the edge.

**CORS / cookies:** Admin SPA still calls `api.functionfly.com`; keep `CORS_ALLOWED_ORIGINS` including `https://admin.functionfly.com`.

---

## 5. IP allowlist (checklist item 37)

**Runtime:** `internal/api/middleware` IP allowlist runs **before** admin routes (`routes_admin.go`).

**Populate:**

1. **Admin UI:** `Admin IP Allowlist` page — add CIDRs for office egress (and optional labels).
2. **API:** `POST /v1/admin/ip-allowlist` (requires admin auth + HMAC where configured) — see `internal/api/handlers/admin/ip_allowlist_admin.go`.

**Order of operations:** First allowlist **your** IP via DB seed or a bootstrap path, or temporarily disable strict allowlist in env if your deployment supports it — otherwise you can lock yourself out.

---

## 6. RBAC (checklist item 38)

Roles such as `super_admin`, `support`, `billing_admin`, `developer_admin`, `read_only` are enforced via `authMiddleware.RequirePermission(...)` on routes. **Verify** in staging that each role can only reach the intended pages (automated tests or manual matrix).

---

## 7. Verification summary

| Item | Verify |
|------|--------|
| OAuth | `GET /v1/auth/oauth/providers` returns `google`/`github` after secrets; OAuth callback completes and user exists in DB. |
| MFA policy | Decide Access-only vs API enforcement; if API, add tests for admin without MFA. |
| SAML | IdP login completes for a test tenant; metadata XML loads. |
| Cloudflare Access | Unauthenticated browser gets Access login; allowed identity reaches SPA. |
| IP allowlist | Request from non-allowlisted IP gets **403** when policy is deny-by-default. |

---

## References

- `docs/ADMIN_CLOUDFLARE_SECURITY.md` — WAF, rate limits, headers.
- `docs/ADMIN_SETUP_README.md` — Creating admin users.
- `docs/FLY_DEPLOYMENT.md` — Secrets and `BASE_URL`.
- `internal/auth/oauth.go` — OAuth redirect URLs.
- `internal/auth/mfa.go` — `IsMFARequired`, `GetMFAStatus`.
