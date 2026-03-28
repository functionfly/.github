import re


def handler(event):
    text = event.get("text") if isinstance(event, dict) else None
    if not text or not isinstance(text, str):
        return {"ok": False, "error": "text (string) is required"}
    try:
        chunk_size = int(event.get("chunk_size", 512))
        overlap = int(event.get("overlap", 50))
        strategy = event.get("strategy", "characters")  # "characters", "words", "sentences"
        if chunk_size < 1:
            return {"ok": False, "error": "chunk_size must be >= 1"}
        if overlap < 0 or overlap >= chunk_size:
            return {"ok": False, "error": "overlap must be >= 0 and < chunk_size"}

        chunks = []
        if strategy == "sentences":
            sentences = re.split(r'(?<=[.!?])\s+', text.strip())
            current_chunk = []
            current_len = 0
            for sent in sentences:
                sent_len = len(sent)
                if current_len + sent_len > chunk_size and current_chunk:
                    chunk_text = " ".join(current_chunk)
                    chunks.append({"text": chunk_text, "start": 0, "end": len(chunk_text), "index": len(chunks)})
                    # Keep overlap sentences
                    overlap_sents = max(0, len(current_chunk) - 1)
                    current_chunk = current_chunk[-overlap_sents:] if overlap_sents > 0 else []
                    current_len = sum(len(s) for s in current_chunk)
                current_chunk.append(sent)
                current_len += sent_len
            if current_chunk:
                chunk_text = " ".join(current_chunk)
                chunks.append({"text": chunk_text, "start": 0, "end": len(chunk_text), "index": len(chunks)})
        elif strategy == "words":
            words = text.split()
            step = chunk_size - overlap
            for i in range(0, len(words), step):
                chunk_words = words[i:i + chunk_size]
                chunk_text = " ".join(chunk_words)
                chunks.append({"text": chunk_text, "word_start": i, "word_end": i + len(chunk_words), "index": len(chunks)})
                if i + chunk_size >= len(words):
                    break
        else:  # characters
            step = chunk_size - overlap
            for i in range(0, len(text), step):
                chunk_text = text[i:i + chunk_size]
                chunks.append({"text": chunk_text, "start": i, "end": i + len(chunk_text), "index": len(chunks)})
                if i + chunk_size >= len(text):
                    break

        return {
            "ok": True,
            "result": chunks,
            "chunks": chunks,
            "count": len(chunks),
            "chunk_size": chunk_size,
            "overlap": overlap,
            "strategy": strategy
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
