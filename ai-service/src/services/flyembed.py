"""FlyEmbed: Triple-vector ColBERT-style embedding service for FunctionFly functions.

Each function gets three specialized 512-dim vectors:
  - Contract: I/O schemas, types, error codes
  - Semantic: Behavioral meaning, description, category
  - Code: Implementation patterns, AST structure

Queries produce three corresponding vectors. Relevance is scored via weighted
MaxSim across all three vectors.
"""

import asyncio
import logging
import time
from typing import Optional

from ..config import get_settings
from ..models.schemas import (
    EmbeddingRequest,
    TripleEmbeddingResult,
    TripleQueryVector,
)
from .embeddings import get_embeddings_service
from .flyembed_contract import ContractTextBuilder
from .flyembed_code import CodeTextBuilder

logger = logging.getLogger(__name__)

# Instruction-tuned prefixes following Qwen3's multi-task embedding paradigm
INSTRUCTION_PREFIXES = {
    "contract": "Represent this function contract for schema matching and I/O compatibility",
    "semantic": "Represent this function for semantic retrieval and behavioral similarity",
    "code": "Represent this code for implementation pattern similarity",
    "contract_query": "Represent this query for matching function contracts",
    "semantic_query": "Represent this query for finding semantically similar functions",
    "code_query": "Represent this query for finding implementation-similar code",
}


class FlyEmbedService:
    """Triple-vector embedding service for FunctionFly functions."""

    def __init__(self):
        self._embeddings_service = get_embeddings_service()
        self._contract_builder = ContractTextBuilder()
        self._code_builder = CodeTextBuilder()

    async def embed_function(self, function_data: dict) -> TripleEmbeddingResult:
        """Generate triple embeddings for a function.

        Args:
            function_data: Dict with keys: name, title, description, category, tags,
                           manifest (input/output schemas), source_code, runtime, capabilities

        Returns:
            TripleEmbeddingResult with contract, semantic, code vectors + source texts
        """
        start = time.monotonic()

        contract_text = self._contract_builder.build(function_data)
        semantic_text = self._build_semantic_text(function_data)
        code_text = self._code_builder.build(function_data)

        # Generate all three embeddings in parallel
        contract_task = self._embed_with_prefix(contract_text, "contract")
        semantic_task = self._embed_with_prefix(semantic_text, "semantic")
        code_task = self._embed_with_prefix(code_text, "code")

        contract_vec, semantic_vec, code_vec = await asyncio.gather(
            contract_task, semantic_task, code_task
        )

        elapsed_ms = (time.monotonic() - start) * 1000

        return TripleEmbeddingResult(
            function_id=function_data.get("function_id", ""),
            contract_embedding=contract_vec,
            semantic_embedding=semantic_vec,
            code_embedding=code_vec,
            contract_text=contract_text,
            semantic_text=semantic_text,
            code_text=code_text,
            latency_ms=elapsed_ms,
        )

    async def embed_query(self, query: str) -> TripleQueryVector:
        """Generate triple query vectors for search.

        Args:
            query: Natural language search query

        Returns:
            TripleQueryVector with three query vectors
        """
        start = time.monotonic()

        # Generate all three query embeddings in parallel
        contract_task = self._embed_with_prefix(query, "contract_query")
        semantic_task = self._embed_with_prefix(query, "semantic_query")
        code_task = self._embed_with_prefix(query, "code_query")

        contract_vec, semantic_vec, code_vec = await asyncio.gather(
            contract_task, semantic_task, code_task
        )

        elapsed_ms = (time.monotonic() - start) * 1000

        return TripleQueryVector(
            query=query,
            contract_vector=contract_vec,
            semantic_vector=semantic_vec,
            code_vector=code_vec,
            latency_ms=elapsed_ms,
        )

    async def embed_batch(self, functions: list[dict]) -> list[TripleEmbeddingResult]:
        """Batch embed multiple functions.

        Processes in batches of flyembed_batch_size to respect rate limits.

        Args:
            functions: List of function data dicts

        Returns:
            List of TripleEmbeddingResult
        """
        settings = get_settings()
        batch_size = settings.flyembed_batch_size
        results = []

        for i in range(0, len(functions), batch_size):
            batch = functions[i : i + batch_size]
            batch_tasks = [self.embed_function(fn) for fn in batch]
            batch_results = await asyncio.gather(*batch_tasks, return_exceptions=True)

            for j, result in enumerate(batch_results):
                if isinstance(result, Exception):
                    logger.error(
                        "Failed to embed function %s: %s",
                        batch[j].get("function_id", "unknown"),
                        result,
                    )
                else:
                    results.append(result)

        return results

    async def _embed_with_prefix(self, text: str, vector_type: str) -> list[float]:
        """Embed text with instruction-tuned prefix and Matryoshka truncation.

        Args:
            text: Text to embed
            vector_type: One of contract, semantic, code, contract_query, semantic_query, code_query

        Returns:
            Embedding vector (512 dims)
        """
        settings = get_settings()
        prefix = INSTRUCTION_PREFIXES.get(vector_type, "")
        full_text = f"{prefix}\n{text}" if prefix else text

        request = EmbeddingRequest(
            text=full_text,
            dimensions=settings.flyembed_dimensions,
        )
        response = await self._embeddings_service.generate_embedding(request)
        return response.embedding

    def _build_semantic_text(self, function_data: dict) -> str:
        """Build semantic text from function metadata.

        Args:
            function_data: Function metadata dict

        Returns:
            Structured semantic text for embedding
        """
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
            if isinstance(tags, list):
                parts.append(f"Tags: {', '.join(tags)}")
        return "\n".join(parts)


# Global singleton
_flyembed_service: Optional[FlyEmbedService] = None


def get_flyembed_service() -> FlyEmbedService:
    """Get the global FlyEmbed service instance."""
    global _flyembed_service
    if _flyembed_service is None:
        _flyembed_service = FlyEmbedService()
    return _flyembed_service
