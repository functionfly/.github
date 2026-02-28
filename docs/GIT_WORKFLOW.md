# Git workflow

This doc describes how we use branches, PRs, and merges for bug fixes, features, and other work.

## Default branch

- **`master`** is the main integration branch. All changes land here via pull requests (or direct push for small/urgent fixes when appropriate).

## Branch naming

Create a branch from an up-to-date `master` using one of these prefixes:

| Type | Prefix | Example | Use for |
|------|--------|---------|--------|
| **Feature** | `feature/` | `feature/state-fabric-ui` | New functionality, larger changes |
| **Bug fix** | `bugfix/` | `bugfix/signup-validation-422` | Fixing incorrect behavior |
| **Hotfix** | `hotfix/` | `hotfix/auth-token-expiry` | Urgent production fixes from `master` |
| **Chore** | `chore/` | `chore/upgrade-deps` | Tooling, deps, refactors, no user-facing change |
| **Docs** | `docs/` | `docs/api-readme` | Documentation only |
| **Session/scratch** | `session/` | `session/agent_xxx` | Ephemeral agent or exploratory work (optional) |

Use short, kebab-case names. Optionally include an issue id: `feature/FF-123-add-export`.

## Workflow by type

### Feature

1. Update from `master`:  
   `git checkout master && git pull origin master`
2. Create branch:  
   `git checkout -b feature/my-feature`
3. Implement, commit with clear messages (e.g. “Add export API”, “Wire export in UI”).
4. Push:  
   `git push -u origin feature/my-feature`
5. Open a **Pull Request** into `master`. Fill the PR template (type = Feature).
6. After review and CI green, merge (squash or merge commit per team preference).
7. Delete the branch after merge; update local:  
   `git checkout master && git pull && git branch -d feature/my-feature`

### Bug fix

1. Update from `master`:  
   `git checkout master && git pull origin master`
2. Create branch:  
   `git checkout -b bugfix/short-description`
3. Fix, add or adjust tests, commit.
4. Push and open a **PR** into `master` (type = Bug fix). Reference any issue (e.g. “Fixes #123”).
5. Merge after review and CI; delete branch when done.

### Hotfix (production)

1. Branch from **`master`**:  
   `git checkout master && git pull && git checkout -b hotfix/issue-description`
2. Make minimal change, test, commit.
3. Push, open PR to `master`, get review and merge.
4. Deploy from `master` (or your release process). Optionally backport to a release branch if you use one.

### Chore / docs

1. Branch from `master`:  
   `git checkout -b chore/thing` or `docs/thing`
2. Commit, push, open PR to `master`. Use PR type “Chore” or “Docs”.
3. Merge when CI passes; no need for heavy review unless it’s risky (e.g. major refactor).

## Pull requests

- **Target**: PRs go into **`master`** unless you use a different integration branch (e.g. `develop`); then document it here.
- **Scope**: One logical change per PR (one feature, one bug, one chore). Split large work into smaller PRs where possible.
- **CI**: Fix any failing checks before merge. The repo’s CI runs on push and on PRs to `master` (and `main`/`develop` if configured).
- **Titles**: Use a short, descriptive title. Optionally prefix: `[Feature] Export API`, `[Bugfix] Signup 422`, `[Chore] Bump Go 1.24`.

## Merging

- Prefer **squash merge** for feature/bugfix branches so `master` gets one commit per PR (easier history and revert).
- Use **merge commit** when you want to preserve full branch history (e.g. for big features).
- Avoid merging with failing CI or unresolved review comments.

## Quick reference

```bash
# Start a new feature
git checkout master && git pull origin master
git checkout -b feature/my-feature
# ... work, commit ...
git push -u origin feature/my-feature
# Open PR at GitHub → master

# Start a bugfix
git checkout master && git pull origin master
git checkout -b bugfix/short-name
# ... fix, commit ...
git push -u origin bugfix/short-name
# Open PR → master

# After PR is merged, tidy local
git checkout master && git pull origin master
git branch -d feature/my-feature   # or bugfix/...
```

## Session / agent branches

Branches like `session/agent_*` are for exploratory or agent-generated work. When something is ready:

1. Merge the session branch into a proper branch (`feature/…` or `bugfix/…`), or  
2. Open a PR from the session branch into `master` and then merge or squash.  
Prefer merging into a `feature/` or `bugfix/` branch and then one PR to `master` so the main branch stays consistent with the workflow above.
