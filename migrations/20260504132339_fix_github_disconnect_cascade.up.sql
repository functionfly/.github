-- Fix ON DELETE behavior for github_imports so that Hard-delete of a GitHub connection succeeds
-- Without this change, Disconnect fails because github_imports.connection_id has no ON DELETE CASCADE,
-- which blocks the DELETE on github_connections (the repo handler code already tries to cascade-delete repos+webhooks+imports in a manual loop before calling DeleteConnection)

ALTER TABLE github_imports
  DROP CONSTRAINT IF EXISTS github_imports_connection_id_fkey,
  ADD CONSTRAINT github_imports_connection_id_fkey
    FOREIGN KEY (connection_id) REFERENCES github_connections(id) ON DELETE CASCADE;

ALTER TABLE github_imports
  DROP CONSTRAINT IF EXISTS github_imports_repo_id_fkey,
  ADD CONSTRAINT github_imports_repo_id_fkey
    FOREIGN KEY (repo_id) REFERENCES github_repos(id) ON DELETE CASCADE;
