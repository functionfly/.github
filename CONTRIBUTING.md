# Contributing

## Git workflow

We use a simple branch-and-PR workflow:

- **Default branch:** `master`
- **Branch naming:** `feature/...`, `bugfix/...`, `hotfix/...`, `chore/...`, `docs/...`
- **Changes:** Open a **Pull Request** into `master`; CI must pass before merge.

Full details (when to use each type, how to merge, session branches): **[docs/GIT_WORKFLOW.md](docs/GIT_WORKFLOW.md)**.

## Quick start

```bash
git checkout master && git pull origin master
git checkout -b feature/my-change   # or bugfix/..., chore/...
# make changes, commit
git push -u origin feature/my-change
# Open a PR at GitHub → master
```

## Issues and PRs

- **New issue:** Use a template (Bug report, Feature request, or Chore) from the [Issues](https://github.com/olyntar/functionfly-web/issues/new/choose) page.
- **Pull requests:** Use the PR template; pick the type (Feature, Bug fix, etc.) and ensure the checklist is done.

## CI

The repo runs format checks, lint, and tests on push and on PRs to `master`. Keep the branch up to date with `master` and fix any failing steps before merge.
