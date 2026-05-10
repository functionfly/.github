-- Revert github_imports FK back to original (no explicit ON DELETE CASCADE on connection and repo)
ALTER TABLE github_imports
  DROP CONSTRAINT IF EXISTS github_imports_connection_id_fkey,
  ADD CONSTRAINT github_imports_connection_id_fkey
    FOREIGN KEY (connection_id) REFERENCES github_connections(id);

ALTER TABLE github_imports
  DROP CONSTRAINT IF EXISTS github_imports_repo_id_fkey,
  ADD CONSTRAINT github_imports_repo_id_fkey
    FOREIGN KEY (repo_id) REFERENCES github_repos(id);
