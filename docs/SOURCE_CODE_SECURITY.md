# Source code confidentiality and secret safety

This document supports **defense in depth**: fewer people with access, automated detection of mistakes, and secrets kept out of git. It does not replace GitHub org settings or legal agreements.

**Related:** [FLY_SECRETS_FROM_ENV.md](FLY_SECRETS_FROM_ENV.md), [INFISICAL_SETUP.md](INFISICAL_SETUP.md), [AGENTS.md](../AGENTS.md) (vault model).

---

## 1. GitHub organization and repository (audit checklist)

Complete periodically (e.g. quarterly with the secrets review below). Owners/admins run this in GitHub UI.

### Access and identity

- [ ] Repository is **private** (or org-internal / Enterprise-only as appropriate).
- [ ] Organization requires **2FA** for all members.
- [ ] **SSO** (SAML/OIDC) enforced for the org if available on your plan.
- [ ] Collaborator list matches current headcount; no stale **outside collaborators**.

### Branch protection (default branch: `master` / `main`)

- [ ] **Require pull request** before merging.
- [ ] **Required approvals** (at least one) from code owners or team.
- [ ] **Dismiss stale reviews** when new commits are pushed (optional but recommended).
- [ ] **No force-push** and **no deletions** on the default branch.
- [ ] **Restrict who can push** to the default branch (admins only or release role).
- [ ] Optional: **require signed commits** for high-assurance workflows.

### Tokens and integrations

- [ ] Prefer **fine-grained personal access tokens** with minimum repositories and permissions.
- [ ] **Classic PATs** scoped narrowly; rotate after offboarding or yearly.
- [ ] **GitHub Actions secrets** scoped per environment (`staging`, `production`) with **required reviewers** where deploy keys are used.
- [ ] Third-party GitHub Apps (Vercel, etc.) reviewed: only necessary repos and permissions.

### Forks and visibility

- [ ] Confirm **fork policy** for the org (allow / disallow private forks per policy).
- [ ] No accidental **public** forks or template exposure of private code.

### Audit (GitHub Enterprise or Cloud audit log)

- [ ] Sample recent **audit log** entries: permission escalations, new tokens, OAuth apps.
- [ ] Confirm **secret scanning** and **push protection** are enabled if your GitHub plan includes them.

---

## 2. Developer endpoints (laptops and workstations)

Enforcement is primarily **policy + IT**; engineering documents expectations here.

### Baseline for anyone with clone access

- [ ] **Full-disk encryption** enabled (FileVault, BitLocker, LUKS).
- [ ] **OS and browser** on automatic security updates.
- [ ] **Screen lock** with short timeout; no shared accounts on dev machines.
- [ ] No copying this repository to **personal remotes**, public gists, or unapproved cloud drives.

### Offboarding (within 24 hours of access removal)

- [ ] Remove user from **GitHub org/repo** and any **Infisical** / **Fly** / **Vercel** projects.
- [ ] Revoke **PATs**, **SSH deploy keys**, and **machine users** tied to that person.
- [ ] Rotate **shared** credentials they may have seen (API keys, `JWT_SECRET`, webhook secrets) per risk assessment.

### AI and IDE tools

- [ ] Use only **company-approved** assistants and modes; prefer **privacy / enterprise** tiers that define data handling and retention.
- [ ] Do not paste **production secrets** or **customer PII** into prompts.
- [ ] If policy requires code to stay in-boundary, use **offline** or **no-training** options per vendor documentation (e.g. Cursor, Copilot business settings).

---

## 3. Quarterly secrets and privilege review

Schedule **at least quarterly** (calendar invite + responsible owner).

### Infisical and Fly.io

- [ ] List active **Infisical** environments and members; remove stale access.
- [ ] `fly secrets list` (and staging app) — names only; confirm each is still required.
- [ ] Re-run or verify sync path: [FLY_SECRETS_FROM_ENV.md](FLY_SECRETS_FROM_ENV.md).

### GitHub

- [ ] **Actions secrets** and **environments**: remove unused; split prod vs staging.
- [ ] **Dependabot** and **Security** tab: open alerts triaged (see CI: CodeQL, Trivy, govulncheck).

### Rotation

- [ ] Rotate **long-lived** tokens on a schedule (OAuth client secrets, webhook signing secrets, `JWT_SECRET` if compromise suspected).
- [ ] After rotation, update **Infisical → Fly** sync or `fly secrets set` as documented; avoid committing values.

### CI/CD least privilege

- [ ] Workflows use **minimal** `permissions:` (see `.github/workflows/*.yml`).
- [ ] **Self-hosted runners** (if any) treated as production assets: patching, network egress rules, and secret isolation.

---

## 4. Automation in this repository

| Control | Location |
|--------|----------|
| Secret scanning (commits) | `.github/workflows/secret-scan.yml` (Gitleaks) |
| Dependency review on PRs | `.github/workflows/dependency-review.yml` |
| CodeQL (Go + JS/TS) | `.github/workflows/codeql.yml` |
| Dependabot version bumps | `.github/dependabot.yml` |
| Go SAST / vuln scan / Trivy | `.github/workflows/ci-cd.yml` |

False positives from Gitleaks can be tuned with a repo-root `.gitleaks.toml` or `.gitleaksignore` if needed.
