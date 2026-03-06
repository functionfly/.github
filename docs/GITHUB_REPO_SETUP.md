# GitHub Repository Setup (FunctionFly)

Recommended GitHub setup for the FunctionFly project using the **functionfly** organization and account **functionflycom**.

## Recommended layout

| Item | Value | Notes |
|------|--------|--------|
| **Organization** | `functionfly` | Created under account functionflycom |
| **Primary repo** | `functionfly/functionfly` | Monorepo (Go module, dashboard, edge, docs) |
| **Default branch** | `master` | CI and workflows use `master`; `main` and `develop` also trigger CI |
| **Go module** | `github.com/functionfly/functionfly` | Already set in `go.mod`; do not change |

## Repository URL

- **Clone (HTTPS):** `https://github.com/functionfly/functionfly.git`
- **Clone (SSH):** `git@github.com:functionfly/functionfly.git`
- **Web:** https://github.com/functionfly/functionfly

## GitHub configuration already in this repo

- **CI/CD:** `.github/workflows/ci-cd.yml` — format, lint, tests, Docker build, staging/production deploy gates
- **Dependency updates:** `.github/workflows/dependency-updates.yml` — weekly Go dependency PRs
- **Issue templates:** `.github/ISSUE_TEMPLATE/` — Bug, Feature, Chore
- **PR template:** `.github/PULL_REQUEST_TEMPLATE.md` — type checklist and GIT_WORKFLOW link
- **Git workflow:** `docs/GIT_WORKFLOW.md` — branching (`feature/`, `bugfix/`, `hotfix/`, etc.) and PR flow

## Required secrets (org or repo)

Configure in **Settings → Secrets and variables → Actions** (org-level or repo-level):

| Secret | Used by | Purpose |
|--------|---------|---------|
| `DOCKER_USERNAME` | ci-cd.yml | Docker Hub login for pushing images |
| `DOCKER_PASSWORD` | ci-cd.yml | Docker Hub token/password |
| `CODECOV_TOKEN` | ci-cd.yml (optional) | Codecov upload; omit if not using codecov.io |

`GITHUB_TOKEN` is provided automatically for dependency-updates and other workflows.

## Docker images (CI)

The pipeline pushes to Docker Hub under the **functionfly** namespace:

- `functionfly/orchestrator-api`
- `functionfly/health-monitor`

Ensure the Docker Hub account/org **functionfly** exists and `DOCKER_USERNAME`/`DOCKER_PASSWORD` have push access.

## Environments

**staging** and **production** are already created. The CI workflow uses them for deploy-staging (on `develop`) and deploy-production (on `master`). Add environment secrets (e.g. `STAGING_ENV`, deploy keys) in **Settings → Environments** when you wire up real deployments.

## App / OAuth (GitHub as IdP)

For “Sign in with GitHub” in the app, create a GitHub OAuth App:

- **Owner:** organization `functionfly` (or user functionflycom)
- **URL:** https://github.com/settings/developers (or org settings → Developer settings)
- Set **Authorization callback URL** to your app (e.g. `https://api.staging.functionfly.com/v1/auth/callback/github` for staging)

Use `GITHUB_CLIENT_ID` and `GITHUB_CLIENT_SECRET` in env (see `.env.staging.example`).

## Staging / env vars

Staging and deployment docs already assume:

- `GITHUB_OWNER=functionfly`
- `GITHUB_REPO=functionfly`
- `GITHUB_TOKEN` — PAT for release/changelog (e.g. in Infisical or `.env.staging`)

## Branch protection (optional)

To require pull requests and status checks before merging into `master`:

1. Go to **Settings → Branches → Add branch protection rule**.
2. Branch name pattern: `master`.
3. Enable **Require a pull request before merging** (e.g. 0 or 1 approvals).
4. Enable **Require status checks to pass** and add the check **CI/CD Pipeline** (or the job names: `quality`, `test`, `integration-test`) after the first workflow run.
5. Optionally enable **Do not allow bypassing the above settings** for admins.

**Note:** On free plans, branch protection is only available for **public** repositories. For a private repo you need [GitHub Pro](https://docs.github.com/en/repositories/configuring-branches-and-merges-in-your-repository/managing-protected-branches/about-protected-branches) (or make the repo public).

## Environments (created)

The repo has two environments for deployments:

- **staging** — used by the workflow when pushing to `develop` (deploy-staging job).
- **production** — used when pushing to `master` (deploy-production job).

Add environment-specific secrets in **Settings → Environments → [staging|production] → Environment secrets** when you implement real deployment steps (e.g. `STAGING_ENV`, Fly tokens, or deploy keys).

## Summary

- **Best repo setup:** One org **functionfly**, one primary repo **functionfly/functionfly**, default branch **master**, with the existing `.github` workflows and docs.
- **Account:** functionflycom owns the **functionfly** org and can create the repo there (or transfer an existing repo into the org).
- **Branch strategy:** `master` = production, `develop` = staging; use `feature/`, `bugfix/`, `hotfix/`, `chore/`, `docs/` as in `docs/GIT_WORKFLOW.md`.
