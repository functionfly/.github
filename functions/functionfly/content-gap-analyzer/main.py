"""Content Gap Analyzer - Analyze content gaps between your content and competitors."""
import re
from collections import Counter


def extract_keywords(text, top_n=20):
    clean_text = re.sub(r'[^\w\s]', ' ', text.lower())
    words = clean_text.split()

    stop_words = {
        "a", "an", "the", "and", "or", "but", "in", "on", "at", "to", "for", "of", "with",
        "by", "from", "as", "is", "was", "are", "were", "been", "be", "have", "has", "had",
        "do", "does", "did", "will", "would", "could", "should", "may", "might", "must",
        "i", "you", "he", "she", "we", "they", "this", "that", "these", "those"
    }

    words = [w for w in words if w not in stop_words and len(w) > 2]
    word_freq = Counter(words)
    return set([word for word, count in word_freq.most_common(top_n)])


def calculate_keyword_density(text, keyword):
    clean_text = re.sub(r'[^\w\s]', ' ', text.lower())
    words = clean_text.split()
    if not words:
        return 0
    keyword_lower = keyword.lower()
    keyword_words = keyword_lower.split()
    count = sum(1 for w in words if keyword_lower in w or w in keyword_words)
    return round(count / len(words), 4)


def handler(event):
    try:
        your_content = event.get("your_content", "")
        competitor_contents = event.get("competitor_contents", [])
        target_keyword = event.get("target_keyword", "")

        if not your_content:
            return {"ok": False, "error": "your_content is required"}
        if not isinstance(competitor_contents, list):
            return {"ok": False, "error": "competitor_contents must be a list"}

        your_keywords = extract_keywords(your_content)

        all_competitor_keywords = set()
        competitor_keyword_lists = []
        for competitor_content in competitor_contents:
            keywords = extract_keywords(competitor_content)
            all_competitor_keywords.update(keywords)
            competitor_keyword_lists.append(keywords)

        your_gaps = all_competitor_keywords - your_keywords
        gaps = sorted(list(your_gaps))[:10]

        overlaps = sorted(list(your_keywords & all_competitor_keywords))

        if competitor_keyword_lists:
            common_in_competitors = set.intersection(*competitor_keyword_lists) if competitor_keyword_lists else set()
            opportunities = sorted(list(common_in_competitors - your_keywords))[:5]
        else:
            opportunities = []

        keyword_density_yours = calculate_keyword_density(your_content, target_keyword)

        competitor_densities = []
        for competitor_content in competitor_contents:
            density = calculate_keyword_density(competitor_content, target_keyword)
            competitor_densities.append(density)

        competitor_avg = sum(competitor_densities) / len(competitor_densities) if competitor_densities else 0

        return {
            "ok": True,
            "gaps": gaps,
            "overlaps": overlaps,
            "opportunities": opportunities,
            "keyword_density_yours": keyword_density_yours,
            "competitor_avg": round(competitor_avg, 4),
            "your_content_length": len(your_content.split()),
            "competitor_count": len(competitor_contents)
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
