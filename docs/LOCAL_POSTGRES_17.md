# Local PostgreSQL 17 with pgvector (single cluster)

This guide sets up **PostgreSQL 17** as the **only** local Postgres cluster on port **5432**, with **pgvector** and other extensions used by FunctionFly. If you have an existing PG 15/16 cluster, you can migrate and then replace it so only PG 17 runs (see section 7).

## Why PostgreSQL 17 + these extensions

| Extension / feature | Purpose |
|--------------------|--------|
| **PostgreSQL 17** | Newer features, better vacuum memory use, incremental backup, logical replication improvements. |
| **pgvector** | Vector embeddings for agent memories and function recommendations (cosine similarity, HNSW). |
| **pg_trgm** | Trigram similarity and search (recommendations, fuzzy matching). |
| **uuid-ossp** | UUID generation (some migrations expect it). |
| **pgcrypto** | Encryption helpers used by the app (e.g. `internal/security/database.go`). |
| **pg_stat_statements** | Query performance monitoring (optional). |
| **unaccent** | Accent-insensitive full-text search (optional; improves search on `registry_functions` etc.). |

## 1. Add PostgreSQL 17 and pgvector (PGDG repository)

Use the [official PGDG APT repository](https://apt.postgresql.org/) so you get PostgreSQL 17 and `postgresql-17-pgvector`.

**Option A – Quick (script):**

```bash
sudo apt install -y postgresql-common ca-certificates
sudo /usr/share/postgresql-common/pgdg/apt.postgresql.org.sh
```

**Option B – Manual (Debian/Ubuntu):**

```bash
. /etc/os-release
sudo apt install -y curl ca-certificates
sudo install -d /usr/share/postgresql-common/pgdg
sudo curl -o /usr/share/postgresql-common/pgdg/apt.postgresql.org.asc --fail \
  https://www.postgresql.org/media/keys/ACCC4CF8.asc
sudo sh -c "echo 'deb [signed-by=/usr/share/postgresql-common/pgdg/apt.postgresql.org.asc] https://apt.postgresql.org/pub/repos/apt $VERSION_CODENAME-pgdg main' > /etc/apt/sources.list.d/pgdg.list"
```

**If you get “Conflicting values set for option Signed-By”:** You have two PGDG source entries using different key paths (e.g. `.asc` vs `.gpg`). Use a single key path and a single source file:

```bash
# See what's defining PGDG
grep -r apt.postgresql.org /etc/apt/sources.list /etc/apt/sources.list.d/

# Remove duplicate/conflicting PGDG entries: keep only one file, one line, with .asc
sudo rm -f /etc/apt/sources.list.d/pgdg.list /etc/apt/sources.list.d/pgdg*.list
. /etc/os-release
sudo install -d /usr/share/postgresql-common/pgdg
sudo curl -o /usr/share/postgresql-common/pgdg/apt.postgresql.org.asc --fail \
  https://www.postgresql.org/media/keys/ACCC4CF8.asc
echo "deb [signed-by=/usr/share/postgresql-common/pgdg/apt.postgresql.org.asc] https://apt.postgresql.org/pub/repos/apt $VERSION_CODENAME-pgdg main" | sudo tee /etc/apt/sources.list.d/pgdg.list
sudo apt update
```

Then install PostgreSQL 17 and pgvector:

```bash
sudo apt update
sudo apt install -y postgresql-17 postgresql-17-pgvector
```

If you see **"File has unexpected size ... Mirror sync in progress?"**:

1. Clear the PGDG apt list cache and update again:  
   `sudo rm -f /var/lib/apt/lists/apt.postgresql.org*`  
   then `sudo apt update`, then retry the install.
2. If it still fails, the mirror can be temporarily out of sync—try again in 30–60 minutes, or use the Docker image for local dev: `docker run -d -p 5432:5432 -e POSTGRES_PASSWORD=postgres pgvector/pgvector:pg17` (see section 8).

Optional: contrib modules (pg_trgm, uuid-ossp, pgcrypto, unaccent). On PGDG the package is often `postgresql-17-contrib`; if it’s not found, the main `postgresql-17` package may already include what you need, or you can skip this:

```bash
sudo apt install -y postgresql-17-contrib
```

## 2. Create a new PostgreSQL 17 cluster (recommended for clean migration)

Using a new cluster avoids mixing old and new major versions.

```bash
# Create default cluster for PG 17 (if not created by package)
sudo pg_createcluster 17 main --start

# Or use a dedicated cluster name, e.g. 'functionfly'
sudo pg_createcluster 17 functionfly --start
```

If you already have PG 16 on 5432, the new PG 17 cluster will get another port (e.g. 5433). Check with `sudo pg_lsclusters`. **You must complete section 7** so that only PG 17 runs on **5432**; the app and docs use 5432 only.

## 3. Dump from old database (e.g. PG 16)

If you are migrating data from an existing local DB:

```bash
# Replace with your actual DB name and port for the OLD cluster (e.g. 5432 for PG 16)
export OLD_PORT=5432
export DB_NAME=functionfly
pg_dump -h localhost -p $OLD_PORT -U postgres -Fc -f functionfly_pre17.dump "$DB_NAME"
```

For a schema-only migration (no data), add `--schema-only`:

```bash
pg_dump -h localhost -p $OLD_PORT -U postgres -Fc --schema-only -f functionfly_pre17_schema.dump "$DB_NAME"
```

## 4. Create database and restore on PostgreSQL 17

Use the port where PG 17 is currently running (`sudo pg_lsclusters`). If you still have PG 16, PG 17 will be on 5433 until you complete section 7.

```bash
# Temporary port for PG 17 (from pg_lsclusters; often 5433 until section 7)
export NEW_PORT=5433
export DB_NAME=functionfly
export DB_USER=postgres

sudo -u postgres psql -p $NEW_PORT -c "CREATE DATABASE $DB_NAME;"
sudo -u postgres pg_restore -p $NEW_PORT -d "$DB_NAME" -U postgres functionfly_pre17.dump
```

If PG 17 is already your only cluster on 5432, use `NEW_PORT=5432`. After section 7, everything uses **5432** only.

## 5. Enable extensions on the new database

Connect to the new DB (same port as step 4, `$NEW_PORT`) and enable extensions. The app’s migration `000000_postgres_extensions.up.sql` does this when you run migrations; you can also run manually:

```bash
sudo -u postgres psql -p $NEW_PORT -d "$DB_NAME" -f - <<'SQL'
-- Required
CREATE EXTENSION IF NOT EXISTS "pg_trgm";
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- pgvector (required for agent_memories and function_embeddings)
CREATE EXTENSION IF NOT EXISTS vector;

-- Encryption (used by app)
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Optional: query performance
CREATE EXTENSION IF NOT EXISTS pg_stat_statements;

-- Optional: accent-insensitive search
CREATE EXTENSION IF NOT EXISTS unaccent;
SQL
```

If you **restore first and then run migrations**, the migration will create any missing extensions. If you use **schema-only restore**, run migrations after restore so schema and extensions stay in sync.

## 6. Point FunctionFly to PostgreSQL 17 (port 5432 only)

**Complete section 7 first** so that PG 17 is the only cluster and runs on **5432**. Then set:

```bash
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=postgres
export DB_PASSWORD=postgres
export DB_NAME=functionfly
export DB_SSLMODE=disable
```

Or use a single URL:

```bash
export DATABASE_URL="postgresql://postgres:postgres@localhost:5432/functionfly?sslmode=disable"
```

Then start the API (with `--skip-migrations` if you already applied migrations, or run migrations as per your workflow). See `AGENTS.md` for the full start sequence.

## 7. Run PostgreSQL 17 on 5432 only (required)

FunctionFly uses **port 5432 only**. If you have both PG 16 and PG 17, do this so only PG 17 runs on 5432:

1. **Stop the old cluster:**

   ```bash
   sudo pg_ctlcluster 16 main stop
   ```

2. **Configure PG 17 to use port 5432** (Debian/Ubuntu):

   ```bash
   bash scripts/pg17-port-5432.sh
   ```

   Or manually: `sudo sed -i "s/^#*port = .*/port = 5432/" /etc/postgresql/17/main/postgresql.conf`

3. **Start PostgreSQL 17:**

   ```bash
   sudo pg_ctlcluster 17 main start
   ```

4. **Optional:** Remove the old cluster so only 17 remains:

   ```bash
   sudo pg_dropcluster 16 main --stop
   ```

Verify: `sudo pg_lsclusters` — PG 17 `main` must show port **5432**. Then use `DB_PORT=5432` everywhere (see section 6 and `AGENTS.md`).

## 8. Docker (local dev when apt fails, or production / CI)

When the PGDG mirror keeps failing with "File has unexpected size", use the official Docker image so you’re not blocked:

```bash
# Stop any existing container on 5432 if you have one
docker stop functionfly-pg17 2>/dev/null; docker rm functionfly-pg17 2>/dev/null

# Run Postgres 17 + pgvector; create DB and user in one go
docker run -d --name functionfly-pg17 -p 5432:5432 \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_PASSWORD=postgres \
  -e POSTGRES_DB=functionfly \
  pgvector/pgvector:pg17
```

Wait a few seconds, then create extensions (once):

```bash
docker exec -i functionfly-pg17 psql -U postgres -d functionfly <<'SQL'
CREATE EXTENSION IF NOT EXISTS "pg_trgm";
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS vector;
CREATE EXTENSION IF NOT EXISTS "pgcrypto";
SQL
```

Use your normal env (port 5432):

```bash
export DB_HOST=localhost DB_PORT=5432 DB_USER=postgres DB_PASSWORD=postgres DB_NAME=functionfly DB_SSLMODE=disable
# Or: export DATABASE_URL="postgresql://postgres:postgres@localhost:5432/functionfly?sslmode=disable"
```

Then start the API and run migrations (or use `--skip-migrations` if you apply them yourself).

For production or CI with docker-compose:

```yaml
postgres:
  image: pgvector/pgvector:pg17
  # ...
```

Ensure extensions are created (migration `000000_postgres_extensions.up.sql` or the manual block in step 5).

## Summary checklist

- [ ] PGDG repo added; `postgresql-17` and `postgresql-17-pgvector` installed.
- [ ] PG 17 cluster created; **section 7 done** so PG 17 runs on **port 5432 only**.
- [ ] Old DB dumped (if migrating data).
- [ ] New DB created on PG 17; dump restored (if applicable).
- [ ] Extensions enabled: `pg_trgm`, `uuid-ossp`, `vector`, `pgcrypto`; optionally `pg_stat_statements`, `unaccent`.
- [ ] `DB_PORT=5432` (and other `DB_*` or `DATABASE_URL`) point to PG 17.
- [ ] Migrations run (or already applied via restore).
- [ ] API and dashboard connect and run correctly.

See also: [PGVECTOR_SETUP.md](PGVECTOR_SETUP.md), [NEON.md](NEON.md), [PRODUCTION_DEPLOYMENT.md](PRODUCTION_DEPLOYMENT.md).
