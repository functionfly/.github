import re
import math
from typing import Any


def count_words(text: str) -> int:
    words = re.findall(r'\b\w+\b', text.lower())
    return len(words)


def count_sentences(text: str) -> int:
    sentences = re.split(r'[.!?]+', text)
    return len([s for s in sentences if len(s.strip()) > 0])


def calculate_readability_score(text: str) -> float:
    words = count_words(text)
    sentences = count_sentences(text)
    
    if sentences == 0:
        sentences = 1
    
    avg_sentence_length = words / sentences
    
    if avg_sentence_length <= 10:
        score = 90
    elif avg_sentence_length <= 15:
        score = 80
    elif avg_sentence_length <= 20:
        score = 70
    elif avg_sentence_length <= 25:
        score = 60
    else:
        score = max(30, 70 - (avg_sentence_length - 25) * 2)
    
    return round(score, 1)


def calculate_keyword_density(text: str, keyword: str) -> float:
    text_lower = text.lower()
    keyword_lower = keyword.lower()
    
    text_words = re.findall(r'\b\w+\b', text_lower)
    total_words = len(text_words)
    
    if total_words == 0:
        return 0.0
    
    keyword_words = keyword_lower.split()
    
    if len(keyword_words) == 1:
        keyword_count = sum(1 for w in text_words if w == keyword_words[0])
    else:
        keyword_count = text_lower.count(keyword_lower)
    
    density = (keyword_count / total_words) * 100
    
    return round(density, 2)


def analyze_on_page_factors(text: str, keyword: str) -> dict:
    text_lower = text.lower()
    keyword_lower = keyword.lower()
    keyword_words = keyword_lower.split()
    
    factors = {}
    
    factors["title_tag"] = 1 if len(text) > 0 else 0
    factors["meta_description"] = 1 if len(text) > 50 else 0
    factors["headings"] = 1 if any(h in text_lower for h in ['<h1', '<h2', '<h3']) else 0
    factors["keyword_in_first_100"] = 1 if keyword_lower in text_lower[:500] else 0
    factors["keyword_density_ok"] = 1 if 0.5 <= calculate_keyword_density(text, keyword) <= 3.0 else 0
    factors["content_length"] = 1 if count_words(text) >= 300 else 0
    factors["internal_links"] = 0
    factors["external_links"] = 0
    factors["images"] = 0
    
    link_pattern = r'href=["\'](https?://[^"\']+)["\']'
    links = re.findall(link_pattern, text_lower)
    for link in links:
        if "yourdomain.com" in link or "example.com" not in link:
            factors["external_links"] += 1
        else:
            factors["internal_links"] += 1
    
    img_pattern = r'<img[^>]+>'
    factors["images"] = len(re.findall(img_pattern, text_lower))
    
    return factors


def generate_recommendations(text: str, keyword: str, on_page_factors: dict, keyword_density: float) -> list:
    recommendations = []
    
    words = count_words(text)
    if words < 300:
        recommendations.append(f"Content is too thin ({words} words). Aim for at least 300-500 words for better SEO.")
    
    if keyword_density < 0.5:
        recommendations.append(f"Keyword '{keyword}' density is too low ({keyword_density}%). Aim for 1-2% density.")
    elif keyword_density > 3.0:
        recommendations.append(f"Keyword '{keyword}' density is too high ({keyword_density}%). Reduce to avoid keyword stuffing.")
    
    if not on_page_factors.get("keyword_in_first_100"):
        recommendations.append(f"Include the keyword '{keyword}' early in your content (within first 100 characters).")
    
    if not on_page_factors.get("headings"):
        recommendations.append("Add heading tags (H1, H2, H3) to structure your content.")
    
    if not on_page_factors.get("meta_description"):
        recommendations.append("Add a meta description between 50-160 characters.")
    
    if on_page_factors.get("images", 0) == 0:
        recommendations.append("Add relevant images to improve engagement and SEO.")
    
    if not recommendations:
        recommendations.append("Your content looks well-optimized for the target keyword.")
    
    return recommendations[:5]


def calculate_seo_score(on_page_factors: dict, readability_score: float, keyword_density: float) -> int:
    score = 0
    
    score += 30 if on_page_factors.get("content_length", 0) else 0
    score += 15 if on_page_factors.get("keyword_density_ok", 0) else 0
    score += 10 if on_page_factors.get("title_tag", 0) else 0
    score += 10 if on_page_factors.get("meta_description", 0) else 0
    score += 10 if on_page_factors.get("headings", 0) else 0
    score += 10 if on_page_factors.get("keyword_in_first_100", 0) else 0
    score += 5 if on_page_factors.get("internal_links", 0) else 0
    score += 5 if on_page_factors.get("external_links", 0) else 0
    score += 5 if on_page_factors.get("images", 0) else 0
    
    readability_factor = readability_score / 100 * 10
    score += readability_factor
    
    return min(100, max(0, int(score)))


def handler(event: dict[str, Any]) -> dict[str, Any]:
    try:
        page_content = event.get("page_content", event.get("url", ""))
        target_keyword = event.get("target_keyword", "")
        
        if not page_content:
            return {"ok": False, "error": "page_content or url is required"}
        
        if not target_keyword:
            return {"ok": False, "error": "target_keyword is required"}
        
        if not isinstance(page_content, str):
            page_content = str(page_content)
        
        if not isinstance(target_keyword, str):
            return {"ok": False, "error": "target_keyword must be a string"}
        
        if len(target_keyword.strip()) == 0:
            return {"ok": False, "error": "target_keyword cannot be empty"}
        
        text_for_analysis = re.sub(r'<[^>]+>', ' ', page_content)
        text_for_analysis = re.sub(r'\s+', ' ', text_for_analysis).strip()
        
        keyword_density = calculate_keyword_density(text_for_analysis, target_keyword)
        on_page_factors = analyze_on_page_factors(page_content, target_keyword)
        readability_score = calculate_readability_score(text_for_analysis)
        seo_score = calculate_seo_score(on_page_factors, readability_score, keyword_density)
        recommendations = generate_recommendations(text_for_analysis, target_keyword, on_page_factors, keyword_density)
        
        on_page_factors_scores = {}
        for k, v in on_page_factors.items():
            on_page_factors_scores[k] = int(v) * 20
        
        return {
            "ok": True,
            "seo_score": seo_score,
            "on_page_factors": on_page_factors_scores,
            "recommendations": recommendations,
            "keyword_density": keyword_density,
            "readability_score": readability_score,
            "word_count": count_words(text_for_analysis)
        }
        
    except Exception as e:
        return {"ok": False, "error": str(e)}
