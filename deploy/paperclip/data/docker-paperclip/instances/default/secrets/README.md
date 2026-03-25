# Secrets (local only)

Do **not** commit `master.key` or other secret material.

Generate a key for local Paperclip instances using your deployment process, and store it outside git (e.g. environment, secret manager, or a local file ignored by `.gitignore`).
