import re
import math
from typing import Any


def tokenize_text(text: str) -> list[str]:
    text = text.lower()
    text = re.sub(r'[^a-z0-9\s]', ' ', text)
    tokens = text.split()
    return [t for t in tokens if len(t) > 2]


def get_ngrams(tokens: list[str], n: int) -> set:
    if len(tokens) < n:
        return set()
    
    ngrams = set()
    for i in range(len(tokens) - n + 1):
        ngram = tuple(tokens[i:i+n])
        ngrams.add(ngram)
    
    return ngrams


def compute_shingle_similarity(text1_tokens: list[str], text2_tokens: list[str], n: int = 3) -> float:
    if not text1_tokens or not text2_tokens:
        return 0.0
    
    ngrams1 = get_ngrams(text1_tokens, n)
    ngrams2 = get_ngrams(text2_tokens, n)
    
    if not ngrams1 or not ngrams2:
        return 0.0
    
    intersection = len(ngrams1 & ngrams2)
    union = len(ngrams1 | ngrams2)
    
    if union == 0:
        return 0.0
    
    jaccard = intersection / union
    
    return round(jaccard * 100, 2)


def find_flagged_segments(text: str, comparison_texts: list[str], threshold: float = 30.0) -> list[dict]:
    text_tokens = tokenize_text(text)
    
    if not text_tokens:
        return []
    
    flagged = []
    
    for i, comp_text in enumerate(comparison_texts):
        if not comp_text or not isinstance(comp_text, str):
            continue
        
        comp_tokens = tokenize_text(comp_text)
        
        if not comp_tokens:
            continue
        
        similarity = compute_shingle_similarity(text_tokens, comp_tokens)
        
        if similarity >= threshold:
            sentences = re.split(r'[.!?]+', comp_text)
            sentences = [s.strip() for s in sentences if len(s.strip()) > 20]
            
            flagged.append({
                "text": sentences[0][:200] + "..." if len(sentences[0]) > 200 else sentences[0],
                "similarity": similarity,
                "source_index": i
            })
    
    flagged.sort(key=lambda x: x["similarity"], reverse=True)
    
    return flagged[:10]


def handler(event: dict[str, Any]) -> dict[str, Any]:
    try:
        text = event.get("text", "")
        comparison_texts = event.get("comparison_texts", [])
        
        if not text:
            return {"ok": False, "error": "text is required"}
        
        if not isinstance(text, str):
            return {"ok": False, "error": "text must be a string"}
        
        if not isinstance(comparison_texts, list):
            return {"ok": False, "error": "comparison_texts must be a list"}
        
        text_tokens = tokenize_text(text)
        
        if len(text_tokens) < 3:
            return {"ok": False, "error": "text is too short for plagiarism analysis (minimum 10 characters with meaningful words)"}
        
        max_similarity = 0.0
        total_similarity = 0.0
        comparisons_made = 0
        
        for comp_text in comparison_texts:
            if not comp_text or not isinstance(comp_text, str):
                continue
            
            comp_tokens = tokenize_text(comp_text)
            
            if not comp_tokens:
                continue
            
            similarity = compute_shingle_similarity(text_tokens, comp_tokens)
            
            max_similarity = max(max_similarity, similarity)
            total_similarity += similarity
            comparisons_made += 1
        
        if comparisons_made == 0:
            similarity_score = 0.0
            originality_percent = 100.0
        else:
            avg_similarity = total_similarity / comparisons_made
            similarity_score = round(avg_similarity, 2)
            originality_percent = round(max(0.0, 100.0 - max_similarity), 2)
        
        flagged_segments = find_flagged_segments(text, comparison_texts)
        
        return {
            "ok": True,
            "similarity_score": similarity_score,
            "flagged_segments": flagged_segments,
            "originality_percent": originality_percent,
            "comparisons_made": comparisons_made
        }
        
    except Exception as e:
        return {"ok": False, "error": str(e)}
