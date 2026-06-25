# Post-Launch TODO

Features below are implemented but gated, disabled, or need finishing touches before full availability.

---

## 🔜 City Rankings™

- [ ] **Company Rankings** — Top Agent Companies leaderboard (referenced as "🚧 future" in CITY_RANKINGS.md)
- [ ] **More countries** — Extend `data/cities_seed.csv` from launch ~70 metros to full top 500 worldwide
- [ ] **Dashboard polish** — "My city" widget UI improvements, opt-in/out flow UX
- [ ] **Email digest** — Notify users when their city moves in rankings (optional)
- [ ] **City Wars™** — Quarterly bracket UI polish, bracket visualization

### References
- Spec: `docs/CITY_RANKINGS.md`
- Scoring: `internal/storage/cityranking/scorer.go`
- Jobs: `internal/jobs/cityranking/`
- Plan: `.kilo/plans/1782018734195-city-rankings-plan.md`

---

## 🔜 University Rankings™

- [ ] **University ambassador program** — Mirror City Ambassadors for top universities
- [ ] **International expansion** — Seed more non-US universities (currently ~170 US-focused)
- [ ] **Department/college-level breakdown** — Sub-division within universities (optional)
- [ ] **University Wars** — Quarterly bracket mirroring City Wars (optional)

### References
- Spec: `docs/UNIVERSITY_RANKINGS.md`
- Scoring: `internal/storage/universityranking/scorer.go`
- Jobs: `internal/jobs/universityranking/`

---

## 🔜 Beta Features (Post-Launch Enablement)

These features are implemented but disabled at launch. They require additional work before general availability.

### Studio (`STUDIO_ENABLED`)
- [ ] **Route middleware** — Gate all `/studio/*` routes behind `STUDIO_ENABLED=true` env var
- [ ] **Frontend nav badge** — Add "Beta" or "Coming Soon" badge to Studio sidebar nav
- [ ] **Per-tenant enablement** — Move from global env var to tenant-specific feature flag system

### Ghost Mode (`GHOST_MODE_ENABLED`)
- [ ] **Route middleware** — Gate all `/v1/ghost/*` routes behind `GHOST_MODE_ENABLED=true` env var
- [ ] **Frontend nav badge** — Add "Beta" or "Coming Soon" badge to Ghost Mode sidebar nav
- [ ] **Per-tenant enablement** — Move from global env var to tenant-specific feature flag system
- [ ] **Reference:** `internal/api/handlers/ghost/handler.go` (~1600 lines, mostly stubbed)

### DRE Blockchain Anchoring (`DRE_BLOCKCHAIN_ANCHORING_ENABLED`)
- [ ] **Verify gate implementation** — Confirm `internal/api/handlers/registry/dre/handlers.go:298` properly returns 503 when disabled
- [ ] **HSM integration** — Implement actual blockchain anchoring when HSM is configured
- [ ] **Per-tenant enablement** — Move from global env var to tenant-specific feature flag system

### SAML SSO (`GBA_SAML_ENABLED`)
- [ ] **Production enablement** — Set `GBA_SAML_ENABLED=true` when enterprise customer requests SAML
- [ ] **Dashboard UI** — Build IdP configuration UI in admin dashboard
- [ ] **SSO audit logs** — Add SSO login/event logging in admin dashboard
- [ ] **Reference:** `internal/auth/gba/plugins/saml/plugin.go`

### SCIM Provisioning (`GBA_SCIM_ENABLED`)
- [ ] **Production enablement** — Set `GBA_SCIM_ENABLED=true` when enterprise customer requests SCIM
- [ ] **IdP Integration Guides** — Create Okta, Azure AD, OneLogin guides
- [ ] **Dashboard UI** — Display SCIM endpoint URL, last sync status, test connection
- [ ] **Reference:** `internal/auth/gba/plugins/scim/plugin.go`

---

## 🔜 Enterprise SSO & Identity

### SCIM Provisioning

**Status:** ✅ Fully implemented (2025-11). Routes registered at `/v1/scim/*`, gated behind Enterprise plan.

| Component | Status | Location |
|-----------|--------|----------|
| SCIM 2.0 routes | ✅ Registered | `internal/api/routes_scim.go` |
| Handler | ✅ Implemented | `internal/api/handlers/auth/scim.go` |
| Service | ✅ Implemented | `internal/auth/scim.go` |
| Storage | ✅ Implemented | `internal/storage/scim_repository.go` |
| Feature flag | ✅ Enterprise-only | `internal/plans/features.go` (`FeatureSCIM`) |

**Remaining SCIM Items:**

- [ ] **IdP Integration Guides**
  - [ ] Okta SCIM guide (`docs/guides/sso-okta.md`)
  - [ ] Azure AD/Entra ID SCIM guide (`docs/guides/sso-azure-ad.md`)
  - [ ] OneLogin SCIM guide (`docs/guides/sso-onelogin.md`)

- [ ] **Dashboard UI** — Settings > SSO
  - [ ] Display SCIM endpoint URL for IdP config
  - [ ] Last sync status and logs
  - [ ] Test connection button

- [ ] **Token Security**
  - [ ] Token expiration (currently no expiry)
  - [ ] Token rotation UI/API
  - [ ] Audit logging for token events

- [ ] **Testing**
  - [ ] E2E with Okta sandbox
  - [ ] E2E with Azure AD dev tenant
  - [ ] E2E with OneLogin dev account
  - [ ] Load test bulk provisioning (1000+ users)

### SAML SSO

- [ ] SAML dashboard UI for IdP configuration (code exists in `internal/auth/gba/plugins/saml/`)
- [ ] SSO audit logs in admin dashboard

---

## 🔜 Additional Post-Launch Items

- [ ] **Magic link / passwordless auth** for enterprise (optional)
- [ ] **Webhook notifications** for user/group changes
- [ ] **Outbound SCIM notifications** to IdP (optional)
- [ ] Just-in-Time (JIT) provisioning from SAML assertions

### Phone Support

- [ ] **Phone support UI** — EnterpriseSupportPage (Low priority, Post-launch)

---

## Reference

- SCIM RFC 7644: https://tools.ietf.org/html/rfc7644
- SCIM Schema RFC 7643: https://tools.ietf.org/html/rfc7643
- SCIM impl: `internal/auth/scim.go`
- SCIM handler: `internal/api/handlers/auth/scim.go`
- SCIM routes: `internal/api/routes_scim.go`
- SCIM storage: `internal/storage/scim_repository.go`
- Feature flag: `internal/plans/features.go` (`FeatureSCIM`)
