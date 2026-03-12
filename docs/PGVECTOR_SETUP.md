# PostgreSQL pgvector setup for FunctionFly

This project uses **pgvector** for:

- **Agent memories** (`agent_memories.embedding`) — vector similarity search for AI memory (cosine distance, 1536 dimensions).
- **Function embeddings** (`function_embeddings.embedding`) — optional semantic similarity for recommendations (same dimensions).

Below is the recommended setup for local development and production.

---

## 1. Install the extension

Postgres must have the pgvector extension installed.

**Ubuntu/Debian (Postgres 16):**
```bash
sudo apt install postgresql-16-pgvector
```

**macOS:**
```bash
brew install pgvector
```

**Docker (production):** Use the official image that already includes pgvector:
```yaml
# docker-compose or similar
postgres:
  image: pgvector/pgvector:pg16
  # ...
```

Then ensure the extension is created in the DB (your app does this in `migrations/000000_postgres_extensions.up.sql`):

```sql
CREATE EXTENSION IF NOT EXISTS vector;
```

---

## 2. Schema: use `vector(1536)` (not BYTEA)

Migrations create `embedding` as **BYTEA** so the schema applies even when pgvector is missing. For real vector search you need the column as type **vector(1536)** (OpenAI-style embeddings).

Apply the migration `20260317000000_pgvector_agent_memories.up.sql` (see `migrations/`). It:

- Converts `agent_memories.embedding` from BYTEA → `vector(1536)` when the vector extension is present.
- Adds an **HNSW** index for fast cosine similarity (`<=>`).

Existing BYTEA values are not converted (they become NULL); the app can re-embed if needed.

---

## 3. Index: HNSW for cosine similarity

Your code uses **cosine distance** (`ORDER BY embedding <=> ?`). The index must use the matching operator class:

```sql
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_agent_memories_embedding_hnsw
ON agent_memories
USING hnsw (embedding vector_cosine_ops)
WITH (m = 16, ef_construction = 64);
```

- **vector_cosine_ops** matches the `<=>` operator.
- **HNSW** is usually better than IVFFlat for typical workloads (good recall/speed tradeoff).
- `m` and `ef_construction` are tuning knobs; defaults above are a solid start.

If you add vector search on `function_embeddings` later, use the same pattern (vector column + HNSW with `vector_cosine_ops`).

---

## 4. Dimension and operators

| Use case        | Dimension | Operator | Index opclass       |
|-----------------|-----------|----------|----------------------|
| Agent memories | 1536      | `<=>`    | vector_cosine_ops   |
| Function embed  | 1536      | `<=>`    | vector_cosine_ops   |

Keep dimensions aligned with your embedding model (e.g. OpenAI `text-embedding-3-small` 1536, or 3072 for `text-embedding-3-large`). If you change model, add a migration to alter the column and rebuild the index.

---

## 5. Checklist

- [ ] Postgres has pgvector installed (`apt`, `brew`, or `pgvector/pgvector:pg16`).
- [ ] Extension created: `CREATE EXTENSION IF NOT EXISTS vector;`
- [ ] `agent_memories.embedding` is type `vector(1536)` (migration applied).
- [ ] HNSW index on `agent_memories (embedding vector_cosine_ops)` exists.
- [ ] Queries use `ORDER BY embedding <=> $query_vector` (already the case in your code).

---

## 6. Optional: function_embeddings

`function_embeddings` is still BYTEA in migrations. If you add semantic search there:

1. Alter `embedding` to `vector(1536)` (same way as agent_memories).
2. Create HNSW with `vector_cosine_ops` and use `<=>` in queries.

Until then, BYTEA is fine if you only store embeddings without similarity search.
