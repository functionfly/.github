# FlyEmbed: Triple-Vector ColBERT Architecture — Implementation Plan

## Implementation Status: COMPLETE

All four phases are implemented:

- **Phase 1 (Data Layer):** Migration, Go models, repository CRUD, weighted MaxSim search
- **Phase 2 (AI Service):** `FlyEmbedService`, `ContractTextBuilder`, `CodeTextBuilder`, API routes, Pydantic schemas, config
- **Phase 3 (Integration):** Indexer triple generation, search ranker, Go service methods, API handlers, route registration, publish pipeline wiring
- **Phase 4 (Migration):** Backfill CLI (`cmd/flyembed-backfill`), dual-read with single-vector fallback, publish pipeline dual-write

### Remaining operational tasks (not code):
1. Run `go run ./cmd/flyembed-backfill` to backfill existing functions
2. Verify triple embedding coverage after backfill
3. A/B comparison testing (100 queries, recall@10)
4. Monitor latency (<50ms p99 target for triple HNSW search)

---

## Executive Summary

Build **FlyEmbed** — a FunctionFly-native triple-vector embedding system using ColBERT-style late interaction. Each function gets three specialized 512-dim embedding vectors instead of one generic 1536-dim vector. Queries produce three corresponding vectors. Relevance is scored via weighted **MaxSim** across all three, enabling dramatically better multi-aspect search like *"auth function that returns JWT and handles rate limiting."*

This replaces the current single-vector `text-embedding-3-small` approach with a purpose-built system that understands function contracts, semantics, and code patterns independently.

---

## Current Architecture (as of 2026-03-28)

### Embedding Generation (Python AI Service)
- **Provider:** OpenAI `text-embedding-3-small` → 1536-dim vector (or Ollama `nomic-embed-text` → 768-dim)
- **Cache:** Redis, SHA-256 key of text+model+dimensions+provider
- **Files:** `ai-service/src/services/embeddings.py`, `ai-service/src/providers/openai.py`

### Embedding Storage (dual-path)
1. **Redis (Python search):** `search:function:{id}` → `{data: JSON, embedding: JSON vector}` — brute-force cosine similarity in Python
2. **PostgreSQL + pgvector (Go recommendations):** `function_embeddings` table with `vector(1536)` column, HNSW index

### Search Flow
- **Python path:** `POST /api/search/functions` → `QueryProcessor` → `SearchIndexer.search()` (Redis brute-force) → `ResultRanker`
- **Go path:** `GET /v1/recommendations` → `Service.SearchSimilarByEmbedding()` → `Repository.SearchFunctionEmbeddingsByVector()` (pgvector `<=>` cosine distance)

### Key Tables
- `function_embeddings` — one row per function, `vector(1536)`, HNSW indexed
- `registry_functions` — function metadata (name, description, category, tags, etc.)
- `registry_function_versions` — manifest JSONB (input/output schemas), source_code, runtime

### Data Available for Text Builders
- **Contract:** `manifest.input.{type,properties,required,schema}`, `manifest.output.{type,properties,schema}`, runtime, timeout, capabilities
- **Semantic:** `name`, `title`, `description`, `category`, `tags`
- **Code:** `source_code` (TEXT), manifest `name`, `runtime`

---

## Architecture: Triple-Vector

Each function gets 3 vectors (512 dims each):

| Vector | Encodes | Source Text | Weight (default) |
|--------|---------|-------------|-------------------|
| **Contract** | I/O schemas, types, error codes | `input_schema + output_schema + param names/types + error codes + runtime + capabilities` | α = 0.35 |
| **Semantic** | Behavioral meaning, description, category | `name + title + description + category + tags` | β = 0.40 |
| **Code** | Implementation patterns, AST structure | `source_code + function name + runtime` | γ = 0.25 |

### Scoring Formula
```
score(Q, F) = α·MaxSim(Q_contract, F_contract) + β·MaxSim(Q_semantic, F_semantic) + γ·MaxSim(Q_code, F_code)
```

Each MaxSim = cosine similarity between query vector and function vector (sentence-level, not per-token ColBERT). The "ColBERT-style" aspect is the multi-vector late interaction pattern — three independent similarity computations merged at query time.

### Why 512 dims
- `text-embedding-3-small` supports Matryoshka (truncate to 512 with minimal quality loss)
- 3×512 = 1536 total dims — same storage as current single vector
- With binary quantization: 3×64 bytes = 192 bytes per function (97% reduction)

---

## Phase 1: Data Layer (Go + SQL)

### 1.1 Migration: `function_embedding_triples`

**File:** `migrations/20260328000000_create_function_embedding_triples.up.sql`

```sql
-- FlyEmbed triple-vector embeddings
CREATE TABLE IF NOT EXISTS function_embedding_triples (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    function_id UUID NOT NULL REFERENCES registry_functions(id) ON DELETE CASCADE,

    -- Three specialized vectors (512 dims each)
    contract_embedding vector(512),
    semantic_embedding vector(512),
    code_embedding vector(512),

    -- Source texts for debugging/re-embedding
    contract_text TEXT,
    semantic_text TEXT,
    code_text TEXT,

    -- Metadata
    embedding_model VARCHAR(100) NOT NULL DEFAULT 'flyembed-v1',
    embedding_version INTEGER NOT NULL DEFAULT 1,
    computed_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    CONSTRAINT uq_triple_embedding UNIQUE (function_id)
);

-- HNSW indexes for each vector column (cosine distance)
CREATE INDEX idx_triple_contract_hnsw ON function_embedding_triples
    USING hnsw (contract_embedding vector_cosine_ops) WITH (m = 16, ef_construction = 64);
CREATE INDEX idx_triple_semantic_hnsw ON function_embedding_triples
    USING hnsw (semantic_embedding vector_cosine_ops) WITH (m = 16, ef_construction = 64);
CREATE INDEX idx_triple_code_hnsw ON function_embedding_triples
    USING hnsw (code_embedding vector_cosine_ops) WITH (m = 16, ef_construction = 64);
```

**File:** `migrations/20260328000000_create_function_embedding_triples.down.sql`

```sql
DROP TABLE IF EXISTS function_embedding_triples;
```

### 1.2 Go Models

**File:** `internal/recommendations/models.go` — add:

```go
// FunctionEmbeddingTriple stores three specialized vectors per function (FlyEmbed).
type FunctionEmbeddingTriple struct {
    ID               uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
    FunctionID       uuid.UUID `json:"function_id" gorm:"type:uuid;not null;uniqueIndex"`
    ContractEmbedding []float32 `json:"contract_embedding" gorm:"type:vector(512)"`
    SemanticEmbedding []float32 `json:"semantic_embedding" gorm:"type:vector(512)"`
    CodeEmbedding     []float32 `json:"code_embedding" gorm:"type:vector(512)"`
    ContractText     *string   `json:"contract_text,omitempty" gorm:"type:text"`
    SemanticText     *string   `json:"semantic_text,omitempty" gorm:"type:text"`
    CodeText         *string   `json:"code_text,omitempty" gorm:"type:text"`
    EmbeddingModel   string    `json:"embedding_model" gorm:"type:varchar(100);default:flyembed-v1"`
    EmbeddingVersion int       `json:"embedding_version" gorm:"default:1"`
    ComputedAt       time.Time `json:"computed_at"`
}

func (FunctionEmbeddingTriple) TableName() string {
    return "function_embedding_triples"
}

// TripleSearchWeights controls the contribution of each vector to the final score.
type TripleSearchWeights struct {
    Contract float64 `json:"contract"` // α — schema/contract match weight
    Semantic float64 `json:"semantic"` // β — behavioral/semantic match weight
    Code     float64 `json:"code"`     // γ — code pattern match weight
}

// DefaultTripleSearchWeights returns the default weight configuration.
func DefaultTripleSearchWeights() TripleSearchWeights {
    return TripleSearchWeights{Contract: 0.35, Semantic: 0.40, Code: 0.25}
}

// TripleSearchResult holds per-vector scores for a single function match.
type TripleSearchResult struct {
    FunctionID     uuid.UUID `json:"function_id"`
    ContractScore  float64   `json:"contract_score"`
    SemanticScore  float64   `json:"semantic_score"`
    CodeScore      float64   `json:"code_score"`
    CombinedScore  float64   `json:"combined_score"`
}
```

### 1.3 Go Repository

**File:** `internal/recommendations/repository.go` — add methods:

```go
// UpsertFunctionEmbeddingTriple creates or updates triple embeddings for a function.
func (r *Repository) UpsertFunctionEmbeddingTriple(ctx context.Context, t *FunctionEmbeddingTriple) error

// GetFunctionEmbeddingTriple returns the triple embedding for a function, or nil.
func (r *Repository) GetFunctionEmbeddingTriple(ctx context.Context, functionID uuid.UUID) (*FunctionEmbeddingTriple, error)

// SearchByTripleVector performs weighted MaxSim across all three vector columns.
// Returns top `limit` functions ordered by combined weighted score.
func (r *Repository) SearchByTripleVector(ctx context.Context, contractVec, semanticVec, codeVec []float32, weights TripleSearchWeights, limit int, excludeID *uuid.UUID) ([]TripleSearchResult, error)

// SearchByContractVector searches by contract vector only (for schema matching).
func (r *Repository) SearchByContractVector(ctx context.Context, vector []float32, limit int) ([]TripleSearchResult, error)

// SearchBySemanticVector searches by semantic vector only.
func (r *Repository) SearchBySemanticVector(ctx context.Context, vector []float32, limit int) ([]TripleSearchResult, error)

// SearchByCodeVector searches by code vector only.
func (r *Repository) SearchByCodeVector(ctx context.Context, vector []float32, limit int) ([]TripleSearchResult, error)

// CountTripleEmbeddings returns the number of functions with triple embeddings.
func (r *Repository) CountTripleEmbeddings(ctx context.Context) (int64, error)

// ListFunctionsWithoutTriples returns function IDs that lack triple embeddings.
func (r *Repository) ListFunctionsWithoutTriples(ctx context.Context, limit int) ([]uuid.UUID, error)
```

**SQL pattern for `SearchByTripleVector`:**
```sql
SELECT function_id,
    1 - (contract_embedding <=> $1) AS contract_score,
    1 - (semantic_embedding <=> $2) AS semantic_score,
    1 - (code_embedding <=> $3) AS code_score,
    ($4 * (1 - (contract_embedding <=> $1))) +
    ($5 * (1 - (semantic_embedding <=> $2))) +
    ($6 * (1 - (code_embedding <=> $3))) AS combined_score
FROM function_embedding_triples
WHERE contract_embedding IS NOT NULL
    AND semantic_embedding IS NOT NULL
    AND code_embedding IS NOT NULL
    AND ($7::uuid IS NULL OR function_id != $7)
ORDER BY combined_score DESC
LIMIT $8
```

### 1.4 Go Service

**File:** `internal/recommendations/service.go` — add methods:

```go
// SearchByTripleEmbedding generates 3 query embeddings via AI service, runs triple MaxSim.
func (s *Service) SearchByTripleEmbedding(ctx context.Context, query string, weights *TripleSearchWeights, limit int) ([]RecommendationResult, error)

// UpsertTripleEmbedding stores a triple embedding for a function.
func (s *Service) UpsertTripleEmbedding(ctx context.Context, functionID uuid.UUID, contract, semantic, code []float32, contractText, semanticText, codeText *string, model string) error

// SearchSimilarByTripleEmbedding uses triple vectors when available, falls back to single vector.
func (s *Service) SearchSimilarByTripleEmbedding(ctx context.Context, query string, limit int, excludeFunctionID *uuid.UUID) ([]RecommendationResult, error)

// FindComposableFunctions finds functions whose contract inputs match the target function's outputs.
func (s *Service) FindComposableFunctions(ctx context.Context, functionID uuid.UUID, limit int) ([]RecommendationResult, error)
```

The `SearchByTripleEmbedding` flow:
1. Call AI service `POST /api/flyembed/query` with the query text → returns 3 query vectors
2. Call `repo.SearchByTripleVector()` with the 3 vectors + weights
3. Resolve function details from registry
4. Return `[]RecommendationResult`

---

## Phase 2: AI Service (Python)

### 2.1 FlyEmbed Core Service

**File:** `ai-service/src/services/flyembed.py`

```python
class FlyEmbedService:
    """Triple-vector embedding service for FunctionFly functions."""

    def __init__(self):
        self._embeddings_service = get_embeddings_service()
        self._contract_builder = ContractTextBuilder()
        self._code_builder = CodeTextBuilder()

    async def embed_function(self, function_data: dict) -> TripleEmbedding:
        """Generate triple embeddings for a function.

        Args:
            function_data: Dict with keys: name, title, description, category, tags,
                           manifest (input/output schemas), source_code, runtime, capabilities

        Returns:
            TripleEmbedding with contract, semantic, code vectors + source texts
        """
        contract_text = self._contract_builder.build(function_data)
        semantic_text = self._build_semantic_text(function_data)
        code_text = self._code_builder.build(function_data)

        # Generate all three embeddings in parallel
        contract_vec = await self._embed_with_prefix(contract_text, "contract")
        semantic_vec = await self._embed_with_prefix(semantic_text, "semantic")
        code_vec = await self._embed_with_prefix(code_text, "code")

        return TripleEmbedding(
            contract_embedding=contract_vec,
            semantic_embedding=semantic_vec,
            code_embedding=code_vec,
            contract_text=contract_text,
            semantic_text=semantic_text,
            code_text=code_text,
        )

    async def embed_query(self, query: str) -> TripleQueryVector:
        """Generate triple query vectors for search."""
        contract_vec = await self._embed_with_prefix(query, "contract_query")
        semantic_vec = await self._embed_with_prefix(query, "semantic_query")
        code_vec = await self._embed_with_prefix(query, "code_query")
        return TripleQueryVector(
            contract=contract_vec, semantic=semantic_vec, code=code_vec
        )

    async def embed_batch(self, functions: list[dict]) -> list[TripleEmbedding]:
        """Batch embed multiple functions."""
        # Process in batches of 10 to respect rate limits
        ...

    async def _embed_with_prefix(self, text: str, vector_type: str) -> list[float]:
        """Embed text with instruction-tuned prefix."""
        prefix = INSTRUCTION_PREFIXES[vector_type]
        full_text = f"{prefix}\n{text}"
        request = EmbeddingRequest(text=full_text, dimensions=512)
        response = await self._embeddings_service.generate_embedding(request)
        return response.embedding

    def _build_semantic_text(self, function_data: dict) -> str:
        """Build semantic text from function metadata."""
        parts = []
        if name := function_data.get("name"):
            parts.append(f"Function: {name}")
        if title := function_data.get("title"):
            parts.append(f"Title: {title}")
        if desc := function_data.get("description"):
            parts.append(f"Description: {desc}")
        if category := function_data.get("category"):
            parts.append(f"Category: {category}")
        if tags := function_data.get("tags"):
            parts.append(f"Tags: {', '.join(tags)}")
        return "\n".join(parts)
```

### 2.2 Instruction-Tuned Prefixes

Following Qwen3's paradigm — each vector type gets a task-specific prefix:

```python
INSTRUCTION_PREFIXES = {
    "contract": "Represent this function contract for schema matching and I/O compatibility",
    "semantic": "Represent this function for semantic retrieval and behavioral similarity",
    "code": "Represent this code for implementation pattern similarity",
    "contract_query": "Represent this query for matching function contracts",
    "semantic_query": "Represent this query for finding semantically similar functions",
    "code_query": "Represent this query for finding implementation-similar code",
}
```

### 2.3 Contract Text Builder

**File:** `ai-service/src/services/flyembed_contract.py`

```python
class ContractTextBuilder:
    """Builds structured contract representation from function manifest data."""

    def build(self, function_data: dict) -> str:
        """Build contract text from function data.

        Returns structured representation like:
            Function: jwt-verify
            Accepts: token: string (required)
            Input Schema: {"type":"object","properties":{"token":{"type":"string"}},"required":["token"]}
            Returns: {"type":"object","properties":{"valid":{"type":"boolean"},"payload":{"type":"object"}}}
            Runtime: node18
            Timeout: 30000ms
            Deterministic: true
            Capabilities: network
        """
        manifest = function_data.get("manifest", {})
        parts = []

        parts.append(f"Function: {function_data.get('name', 'unknown')}")

        # Input schema
        input_schema = manifest.get("input", {})
        if input_schema:
            if props := input_schema.get("properties"):
                params = []
                required = set(input_schema.get("required", []))
                for param_name, param_def in props.items():
                    param_type = param_def.get("type", "any")
                    desc = param_def.get("description", "")
                    req_marker = " (required)" if param_name in required else ""
                    params.append(f"{param_name}: {param_type}{req_marker}")
                    if desc:
                        params.append(f"  {desc}")
                parts.append(f"Accepts: {', '.join(p for p in params if not p.startswith('  '))}")

            # Full JSON schema (compact)
            parts.append(f"Input Schema: {json.dumps(input_schema, separators=(',', ':'))}")

        # Output schema
        output_schema = manifest.get("output", {})
        if output_schema:
            parts.append(f"Returns: {json.dumps(output_schema, separators=(',', ':'))}")

        # Runtime constraints
        if runtime := function_data.get("runtime"):
            parts.append(f"Runtime: {runtime}")
        if timeout := manifest.get("timeout_ms"):
            parts.append(f"Timeout: {timeout}ms")
        if deterministic := manifest.get("deterministic"):
            parts.append(f"Deterministic: {deterministic}")
        if caps := function_data.get("capabilities"):
            parts.append(f"Capabilities: {', '.join(caps)}")
        if side_effects := manifest.get("side_effects"):
            parts.append(f"Side Effects: {side_effects}")

        return "\n".join(parts)
```

### 2.4 Code Text Builder

**File:** `ai-service/src/services/flyembed_code.py`

```python
class CodeTextBuilder:
    """Builds AST-aware code representation for embedding."""

    def build(self, function_data: dict) -> str:
        """Build code text from function data.

        Returns:
            function jwt-verify
            runtime: node18
            ---
            // source code content
        """
        parts = []
        name = function_data.get("name", "unknown")
        runtime = function_data.get("runtime", "unknown")

        parts.append(f"function {name}")
        parts.append(f"runtime: {runtime}")

        # Extract imports from source code (best-effort)
        source_code = function_data.get("source_code", "")
        if source_code:
            imports = self._extract_imports(source_code, runtime)
            if imports:
                parts.append(f"imports: {', '.join(imports)}")
            parts.append("---")
            # Truncate to first 2000 chars to stay within embedding context limits
            parts.append(source_code[:2000])

        return "\n".join(parts)

    def _extract_imports(self, source: str, runtime: str) -> list[str]:
        """Extract import statements from source code."""
        imports = []
        if "node" in runtime or "javascript" in runtime:
            # Match require() and import statements
            for match in re.finditer(r"(?:require\(['\"]([^'\"]+)['\"]\)|import\s+.*?from\s+['\"]([^'\"]+)['\"])", source):
                imports.append(match.group(1) or match.group(2))
        elif "python" in runtime:
            # Match import and from...import
            for match in re.finditer(r"(?:^from\s+(\S+)\s+import|^import\s+(\S+))", source, re.MULTILINE):
                imports.append(match.group(1) or match.group(2))
        return imports[:10]  # Limit to 10 imports
```

### 2.5 Config Updates

**File:** `ai-service/src/config.py` — add:

```python
# FlyEmbed configuration
flyembed_model: str = "text-embedding-3-small"
flyembed_dimensions: int = 512
flyembed_default_weight_contract: float = 0.35
flyembed_default_weight_semantic: float = 0.40
flyembed_default_weight_code: float = 0.25
flyembed_batch_size: int = 10
flyembed_max_source_code_chars: int = 2000
```

### 2.6 Pydantic Schemas

**File:** `ai-service/src/models/schemas.py` — add:

```python
class TripleEmbeddingRequest(BaseModel):
    function_id: str
    name: str
    title: Optional[str] = None
    description: Optional[str] = None
    category: Optional[str] = None
    tags: List[str] = Field(default_factory=list)
    manifest: Dict[str, Any] = Field(default_factory=dict)
    source_code: Optional[str] = None
    runtime: Optional[str] = None
    capabilities: List[str] = Field(default_factory=list)

class TripleEmbeddingResponse(BaseModel):
    function_id: str
    contract_embedding: List[float]
    semantic_embedding: List[float]
    code_embedding: List[float]
    contract_text: str
    semantic_text: str
    code_text: str
    embedding_model: str = "flyembed-v1"
    dimensions: int = 512
    latency_ms: float = 0.0

class TripleQueryRequest(BaseModel):
    query: str = Field(..., min_length=1, max_length=500)

class TripleQueryResponse(BaseModel):
    query: str
    contract_vector: List[float]
    semantic_vector: List[float]
    code_vector: List[float]
    dimensions: int = 512
    latency_ms: float = 0.0

class TripleEmbeddingBatchRequest(BaseModel):
    functions: List[TripleEmbeddingRequest] = Field(..., max_length=50)

class TripleEmbeddingBatchResponse(BaseModel):
    results: List[TripleEmbeddingResponse]
    total_count: int
    latency_ms: float = 0.0
```

### 2.7 API Routes

**File:** `ai-service/src/api/routes.py` — add:

```python
@router.post("/api/flyembed/embed", response_model=TripleEmbeddingResponse)
async def flyembed_embed(request: TripleEmbeddingRequest) -> TripleEmbeddingResponse:
    """Generate triple embeddings for a single function."""

@router.post("/api/flyembed/embed-batch", response_model=TripleEmbeddingBatchResponse)
async def flyembed_embed_batch(request: TripleEmbeddingBatchRequest) -> TripleEmbeddingBatchResponse:
    """Batch generate triple embeddings for multiple functions."""

@router.post("/api/flyembed/query", response_model=TripleQueryResponse)
async def flyembed_query(request: TripleQueryRequest) -> TripleQueryResponse:
    """Generate triple query vectors for search."""

@router.get("/api/flyembed/health")
async def flyembed_health():
    """Health check for FlyEmbed service."""
```

---

## Phase 3: Integration

### 3.1 Indexing Pipeline Update

**File:** `ai-service/src/services/search/indexer.py`

Modify `index_function()` to generate triple embeddings:

```python
async def index_function(self, function_id: str, function_data: dict) -> bool:
    # Existing: single-vector embedding for backward compat
    searchable_text = self._create_searchable_text(function_data)
    request = EmbeddingRequest(text=searchable_text)
    embedding_response = await self._embeddings_service.generate_embedding(request)

    # NEW: triple-vector embedding via FlyEmbed
    flyembed = get_flyembed_service()
    triple = await flyembed.embed_function(function_data)

    # Store single vector in Redis (existing)
    await redis.hset(function_key, mapping={
        "data": json.dumps(function_data),
        "embedding": json.dumps(embedding_response.embedding),
        # NEW: triple vectors
        "contract_embedding": json.dumps(triple.contract_embedding),
        "semantic_embedding": json.dumps(triple.semantic_embedding),
        "code_embedding": json.dumps(triple.code_embedding),
    })

    # Store triple in pgvector via Go API
    await self._store_triple_embedding(function_id, triple)

    return True
```

### 3.2 Search Pipeline Update

**File:** `ai-service/src/services/search/indexer.py` — add triple search:

```python
async def search_triple(self, tenant_id: str, query_embedding: TripleQueryVector, limit: int = 20, weights: dict = None) -> list[dict]:
    """Search using triple-vector MaxSim scoring."""
    if weights is None:
        weights = {"contract": 0.35, "semantic": 0.40, "code": 0.25}

    # Option A: pgvector search via Go API (recommended for scale)
    results = await self._search_pgvector_triple(query_embedding, weights, limit)

    # Option B: Redis brute-force (fallback, for small indices)
    # results = await self._search_redis_triple(tenant_id, query_embedding, weights, limit)

    return results
```

**File:** `ai-service/src/services/search/ranker.py` — add triple-aware ranking:

```python
def rank_results_triple(self, results, query, weights=None):
    """Rank results using triple-vector scores + keyword + recency."""
    if weights is None:
        weights = {"contract": 0.35, "semantic": 0.40, "code": 0.25}

    for result in results:
        contract_score = result.get("contract_score", 0)
        semantic_score = result.get("semantic_score", 0)
        code_score = result.get("code_score", 0)

        triple_score = (
            contract_score * weights["contract"] +
            semantic_score * weights["semantic"] +
            code_score * weights["code"]
        )

        keyword_score = self._calculate_keyword_score(result.get("data", {}), query.split())
        recency_score = self._calculate_recency_score(result.get("data", {}))

        result["final_score"] = (
            triple_score * 0.5 +
            keyword_score * 0.3 +
            recency_score * 0.2
        )
        result["triple_score"] = triple_score

    results.sort(key=lambda x: x["final_score"], reverse=True)
    for i, r in enumerate(results):
        r["rank"] = i + 1
    return results
```

### 3.3 Search Route Update

**File:** `ai-service/src/api/routes.py` — update `POST /api/search/functions`:

Add optional `use_triple` and `weights` parameters to `SearchQuery`:

```python
class SearchQuery(BaseModel):
    query: str = Field(..., min_length=1, max_length=500)
    limit: int = Field(default=20, ge=1, le=50)
    filters: Optional[Dict[str, Any]] = None
    use_triple: bool = True  # NEW: enable triple-vector search
    weights: Optional[Dict[str, float]] = None  # NEW: custom weights {contract, semantic, code}
```

When `use_triple=True`, the handler calls `indexer.search_triple()` instead of `indexer.search()`.

### 3.4 Go Recommendations Integration

**File:** `internal/recommendations/service.go`

Update `SearchSimilarByEmbedding()` to try triple vectors first:

```go
func (s *Service) SearchSimilarByEmbedding(ctx context.Context, queryEmbedding []float32, limit int, excludeFunctionID *uuid.UUID) ([]RecommendationResult, error) {
    // Try triple-vector search first (if triples exist)
    tripleCount, _ := s.repo.CountTripleEmbeddings(ctx)
    if tripleCount > 0 {
        // Note: this requires the caller to provide a query string, not just a vector
        // For the single-vector path, we still support the existing flow
        logrus.Debug("Triple embeddings available, but single-vector query provided; using single-vector search")
    }
    // ... existing single-vector logic
}
```

Add new triple-native methods (as defined in Phase 1.4).

### 3.5 API Routes Registration

**File:** `internal/api/routes_registry.go` — add:

```go
// Triple-vector search
api.HandleFunc("/recommendations/triple-search", recommendationHandler.HandleTripleSearch).Methods("POST", "OPTIONS")
api.HandleFunc("/recommendations/composable/{function_id}", recommendationHandler.HandleFindComposable).Methods("GET", "OPTIONS")
```

**File:** `internal/api/handlers/recommendations/handler.go` — add:

```go
// HandleTripleSearch handles triple-vector search requests
// POST /v1/recommendations/triple-search
func (h *Handler) HandleTripleSearch(w http.ResponseWriter, r *http.Request)

// HandleFindComposable finds functions composable with the target
// GET /v1/recommendations/composable/{function_id}
func (h *Handler) HandleFindComposable(w http.ResponseWriter, r *http.Request)
```

### 3.6 Dashboard Integration (optional, future)

Update search UI to show per-vector score breakdown:
- "Schema match: 92%"
- "Semantic match: 87%"
- "Code similarity: 71%"

This is a Phase 3b item — not blocking for the core implementation.

---

## Phase 4: Migration & Backfill — **COMPLETE**

### 4.1 Backfill Script

**File:** `cmd/flyembed-backfill/main.go` — **CREATED**

Usage:
```bash
# Dry run — see what would be backfilled
go run ./cmd/flyembed-backfill --dry-run

# Backfill all public functions without triple embeddings
go run ./cmd/flyembed-backfill

# Backfill with custom batch size and limit
go run ./cmd/flyembed-backfill --batch=20 --limit=100
```

The tool iterates all public functions in `registry_functions` that lack triple embeddings, calls `POST /api/flyembed/embed` on the AI service, and stores results via `UpsertFunctionEmbeddingTriple()`. Idempotent — skips functions that already have triple embeddings.

### 4.2 Dual-Read Period — **IMPLEMENTED**

`SearchSimilarByTripleEmbedding()` in `service.go` now:
1. Checks `CountTripleEmbeddings()` — if 0, falls back to single-vector search
2. If FlyEmbed query fails, falls back to single-vector search
3. Uses triple-vector MaxSim when triples are available

The Python search route (`POST /api/search/functions`) tries triple vectors first via `use_triple=True` (default), falls back to single-vector on failure.

### 4.3 Publish Pipeline Dual-Write — **IMPLEMENTED**

`HandlePublish` in `publish.go` fires a goroutine after successful publish to generate triple embeddings via `EmbedFunctionViaAIService()`. This is fire-and-forget — failures are logged but don't block the publish response.

New functions published via `POST /v1/registry/publish` automatically get both:
- Single-vector embedding (existing `function_embeddings` table, via AI service indexer)
- Triple-vector embedding (new `function_embedding_triples` table, via publish goroutine)

---

## Files to Create

| File | Description | Status |
|------|-------------|--------|
| `migrations/20260328000000_create_function_embedding_triples.up.sql` | New table with 3 vector(512) columns + HNSW indexes | DONE |
| `migrations/20260328000000_create_function_embedding_triples.down.sql` | Rollback migration | DONE |
| `ai-service/src/services/flyembed.py` | Core triple embedding service | DONE |
| `ai-service/src/services/flyembed_contract.py` | Contract text builder | DONE |
| `ai-service/src/services/flyembed_code.py` | Code text builder with AST preprocessing | DONE |
| `cmd/flyembed-backfill/main.go` | Backfill script for existing functions | DONE |

## Files to Modify

| File | Changes | Status |
|------|---------|--------|
| `internal/recommendations/models.go` | Add `FunctionEmbeddingTriple`, `TripleSearchResult`, `TripleSearchWeights` structs | DONE |
| `internal/recommendations/repository.go` | Add `AutoMigrate` update, triple vector CRUD + search methods | DONE |
| `internal/recommendations/service.go` | Add `SearchByTripleEmbedding()`, `UpsertTripleEmbedding()`, `FindComposableFunctions()`, `SearchSimilarByTripleEmbedding()` with single-vector fallback | DONE |
| `internal/api/handlers/recommendations/handler.go` | Add `HandleTripleSearch`, `HandleFindComposable` handlers | DONE |
| `internal/api/routes_registry.go` | Register new triple search routes | DONE |
| `ai-service/src/api/routes.py` | Add `/api/flyembed/*` endpoints, update search to support `use_triple` | DONE |
| `ai-service/src/services/search/indexer.py` | Add `search_triple()`, update `index_function()` to generate triples | DONE |
| `ai-service/src/services/search/ranker.py` | Add `rank_results_triple()` method | DONE |
| `ai-service/src/config.py` | Add FlyEmbed config fields | DONE |
| `ai-service/src/models/schemas.py` | Add `TripleEmbeddingRequest`, `TripleEmbeddingResponse`, `TripleQueryRequest`, `TripleQueryResponse`, `TripleEmbeddingBatchRequest`, `TripleEmbeddingBatchResponse`; add `use_triple`/`weights` to `SearchQuery` | DONE |
| `internal/api/handlers/registry/handlers.go` | Add `recommendationSvc` field to `Handler` struct | DONE |
| `internal/api/handlers/registry/publish.go` | Fire-and-forget triple embedding generation on publish | DONE |
| `internal/api/routes.go` | Pass `recommendationSvc` to registry `NewHandler` | DONE |

---

## Key Design Decisions

1. **512 dims per vector (not 1536)** — Matryoshka means 512 is sufficient with `text-embedding-3-small`; saves 66% storage vs 1536 per vector
2. **Separate table** — `function_embedding_triples` coexists with `function_embeddings` during migration; no breaking changes
3. **Reuse OpenAI embeddings API** — No custom model training in v1; use `text-embedding-3-small` with `dimensions=512` + Matryoshka. Custom fine-tuning is v2
4. **Instruction-tuned prefixes** — Each vector type gets a task-specific prefix, following Qwen3's proven approach for multi-task embeddings
5. **Weighted scoring** — Default α=0.35, β=0.40, γ=0.25; configurable per-query via API
6. **pgvector HNSW** — One index per vector column; queries run 3 parallel HNSW searches then merge with weighted scoring
7. **Dual storage** — Triple vectors in pgvector (Go path) for production scale; Redis (Python path) gets triple vectors cached alongside single vectors for backward compatibility
8. **Backward compatible** — Single-vector search continues to work; triple-vector is additive. `SearchSimilarByEmbedding()` falls back gracefully

---

## Verification

1. **Unit tests:** Test `ContractTextBuilder` and `CodeTextBuilder` produce correct structured representations
2. **Integration tests:** Generate triple embeddings for sample functions, verify pgvector search returns expected results
3. **A/B comparison:** Run 100 test queries through both old single-vector and new triple-vector search; compare recall@10
4. **Backfill verification:** After backfill, verify all public functions have triple embeddings with correct dimensions
5. **Latency check:** Triple HNSW search should be <50ms p99 for 10K functions
6. **Go lint:** `golangci-lint run`
7. **Go tests:** `go test ./internal/recommendations/...`
8. **Python tests:** `cd ai-service && python -m pytest`

---

## Implementation Order

1. **Phase 1** (Data Layer) — migration + Go models/repository/service — COMPLETE
2. **Phase 2** (AI Service) — FlyEmbed service + text builders + API routes — COMPLETE
3. **Phase 3** (Integration) — wires everything together — COMPLETE
4. **Phase 4** (Migration) — backfill + dual-read period + publish pipeline — COMPLETE
