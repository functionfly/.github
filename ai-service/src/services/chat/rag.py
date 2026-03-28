"""RAG (retrieval-augmented generation) for chat.

Lightweight docs retrieval over the repo's Markdown docs. We keep this intentionally
dependency-free and robust in local + production environments.
"""

from __future__ import annotations

import hashlib
import logging
import math
import re
import time
from dataclasses import dataclass
from pathlib import Path
from typing import Dict, List, Optional, Tuple

from ...config import settings
from ..embeddings import get_embeddings_service
from ...models.schemas import EmbeddingRequest, ProviderType

logger = logging.getLogger(__name__)


_WORD_RE = re.compile(r"[a-zA-Z0-9_]{2,}")


def _tokenize(text: str) -> List[str]:
    return _WORD_RE.findall(text.lower())


def _cosine(a: List[float], b: List[float]) -> float:
    if not a or not b or len(a) != len(b):
        return 0.0
    dot = 0.0
    na = 0.0
    nb = 0.0
    for i in range(len(a)):
        x = float(a[i])
        y = float(b[i])
        dot += x * y
        na += x * x
        nb += y * y
    if na <= 0.0 or nb <= 0.0:
        return 0.0
    return dot / (math.sqrt(na) * math.sqrt(nb))


def _rag_embedding_request(text: str) -> EmbeddingRequest:
    """Build an embedding request for RAG using configured local/cloud provider."""
    p = (settings.rag_embedding_provider or "ollama").strip().lower()
    if p in ("ollama", "local"):
        return EmbeddingRequest(
            text=text,
            provider=ProviderType.OLLAMA,
            model=settings.ollama_embedding_model,
        )
    if p == "openai":
        return EmbeddingRequest(
            text=text,
            provider=ProviderType.OPENAI,
            model=settings.openai_embedding_model,
            dimensions=settings.openai_embedding_dimensions,
        )
    # Default: local Ollama
    return EmbeddingRequest(
        text=text,
        provider=ProviderType.OLLAMA,
        model=settings.ollama_embedding_model,
    )


def _rag_embedding_precheck() -> None:
    """Fail fast when the chosen provider cannot run (avoids long timeouts)."""
    p = (settings.rag_embedding_provider or "ollama").strip().lower()
    if p in ("anthropic", "openrouter"):
        raise RuntimeError(
            "rag_embedding_provider must be ollama or openai (Anthropic/OpenRouter have no embeddings API here)"
        )
    if p == "openai":
        if not settings.openai_api_key:
            raise RuntimeError("OPENAI_API_KEY not set (rag_embedding_provider=openai)")


@dataclass(frozen=True)
class RagChunk:
    chunk_id: str
    source: str
    title: str
    text: str
    tokens: Tuple[str, ...]


class RagIndex:
    def __init__(self) -> None:
        self._loaded = False
        self._chunks: List[RagChunk] = []
        self._embedding_cache: Dict[str, List[float]] = {}

    def _make_chunk_id(self, source: str, idx: int, text: str) -> str:
        h = hashlib.sha256()
        h.update(source.encode("utf-8"))
        h.update(b"\n")
        h.update(str(idx).encode("utf-8"))
        h.update(b"\n")
        h.update(text.encode("utf-8"))
        return h.hexdigest()

    def _guess_title(self, markdown: str, fallback: str) -> str:
        for line in markdown.splitlines():
            s = line.strip()
            if s.startswith("#"):
                return s.lstrip("#").strip()[:80] or fallback
        return fallback

    def _chunk_markdown(self, markdown: str, max_chars: int) -> List[str]:
        # Simple chunking by paragraphs with a char budget.
        paras = [p.strip() for p in markdown.split("\n\n") if p.strip()]
        chunks: List[str] = []
        buf: List[str] = []
        size = 0
        for p in paras:
            if len(p) > max_chars:
                # Hard-split very large paragraphs
                for i in range(0, len(p), max_chars):
                    part = p[i:i + max_chars].strip()
                    if part:
                        chunks.append(part)
                continue

            if size + len(p) + 2 > max_chars and buf:
                chunks.append("\n\n".join(buf).strip())
                buf = [p]
                size = len(p)
            else:
                buf.append(p)
                size += len(p) + 2

        if buf:
            chunks.append("\n\n".join(buf).strip())
        return chunks

    def load(self) -> None:
        if self._loaded:
            return
        self._loaded = True

        if not settings.enable_rag:
            logger.info("RAG disabled by config")
            return

        docs_dir = Path(settings.rag_docs_dir)
        if not docs_dir.exists() or not docs_dir.is_dir():
            logger.warning(f"RAG docs dir not found: {docs_dir}")
            return

        md_files = sorted(docs_dir.rglob("*.md"))
        if not md_files:
            logger.warning(f"No markdown docs found in: {docs_dir}")
            return

        max_chars = max(400, int(settings.rag_chunk_max_chars))

        chunks: List[RagChunk] = []
        for md_path in md_files:
            if len(chunks) >= int(settings.rag_max_chunks):
                break
            try:
                raw = md_path.read_text(encoding="utf-8", errors="ignore")
            except Exception as e:
                logger.debug(f"Failed to read doc {md_path}: {e}")
                continue

            source = str(md_path.relative_to(docs_dir))
            title = self._guess_title(raw, fallback=md_path.stem)

            for idx, chunk in enumerate(self._chunk_markdown(raw, max_chars=max_chars)):
                if len(chunk) < int(settings.rag_chunk_min_chars):
                    continue
                toks = tuple(_tokenize(chunk))
                if not toks:
                    continue
                chunk_id = self._make_chunk_id(source, idx, chunk)
                chunks.append(
                    RagChunk(
                        chunk_id=chunk_id,
                        source=source,
                        title=title,
                        text=chunk,
                        tokens=toks,
                    )
                )
                if len(chunks) >= int(settings.rag_max_chunks):
                    break

        self._chunks = chunks
        logger.info(f"RAG loaded {len(self._chunks)} chunks from {docs_dir}")

    def _lexical_score(self, query_tokens: List[str], chunk_tokens: Tuple[str, ...]) -> float:
        if not query_tokens or not chunk_tokens:
            return 0.0
        q = set(query_tokens)
        if not q:
            return 0.0
        c = set(chunk_tokens)
        inter = q.intersection(c)
        if not inter:
            return 0.0
        # Slight preference for chunks that match more distinct query tokens.
        return float(len(inter)) / float(max(1, len(q)))

    async def retrieve(self, query: str, k: Optional[int] = None) -> List[Tuple[RagChunk, float]]:
        self.load()
        if not self._chunks:
            return []

        k = int(k or settings.rag_top_k)
        if k <= 0:
            return []

        q_tokens = _tokenize(query)
        if not q_tokens:
            return []

        # Candidate selection by lexical overlap (cheap).
        scored = []
        for ch in self._chunks:
            s = self._lexical_score(q_tokens, ch.tokens)
            if s > 0:
                scored.append((ch, s))
        if not scored:
            return []

        scored.sort(key=lambda x: x[1], reverse=True)
        candidates = [ch for ch, _ in scored[: int(settings.rag_candidate_chunks)]]

        # Prefer embeddings similarity, but cap work/time to keep chat responsive.
        try:
            _rag_embedding_precheck()

            embeddings_service = get_embeddings_service()
            query_emb_resp = await embeddings_service.generate_embedding(_rag_embedding_request(query))
            query_emb = query_emb_resp.embedding

            results: List[Tuple[RagChunk, float]] = []
            rerank_limit = max(
                int(settings.rag_top_k),
                int(settings.rag_embedding_rerank_chunks),
            )
            embedding_candidates = candidates[:rerank_limit]
            deadline = time.monotonic() + float(settings.rag_embedding_max_seconds)
            for ch in embedding_candidates:
                if time.monotonic() > deadline:
                    logger.info(
                        "RAG rerank time budget exceeded; returning partial rerank results",
                    )
                    break
                emb = self._embedding_cache.get(ch.chunk_id)
                if emb is None:
                    emb_resp = await embeddings_service.generate_embedding(_rag_embedding_request(ch.text))
                    emb = emb_resp.embedding
                    self._embedding_cache[ch.chunk_id] = emb
                sim = _cosine(query_emb, emb)
                results.append((ch, sim))

            if not results:
                return [(ch, float(self._lexical_score(q_tokens, ch.tokens))) for ch in candidates[:k]]
            results.sort(key=lambda x: x[1], reverse=True)
            return results[:k]
        except Exception as e:
            logger.warning(f"RAG embeddings unavailable; using lexical fallback: {e}")
            # Use the already lexically-scored candidates.
            return [(ch, float(self._lexical_score(q_tokens, ch.tokens))) for ch in candidates[:k]]

    async def build_context_block(self, query: str) -> str:
        hits = await self.retrieve(query)
        if not hits:
            return ""

        lines: List[str] = []
        lines.append("Relevant documentation excerpts (use these as ground truth when applicable):")
        for i, (ch, score) in enumerate(hits, start=1):
            excerpt = ch.text.strip()
            lines.append(f"\n[{i}] {ch.title} — {ch.source} (score={score:.3f})")
            lines.append(excerpt)
        return "\n".join(lines).strip()


_rag_index: Optional[RagIndex] = None


def get_rag_index() -> RagIndex:
    global _rag_index
    if _rag_index is None:
        _rag_index = RagIndex()
    return _rag_index

