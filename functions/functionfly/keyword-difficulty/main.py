"""Keyword Difficulty Calculator - Calculate keyword difficulty score."""
import re
from urllib.parse import urlparse


def estimate_domain_authority(url):
    parsed = urlparse(url)
    domain = parsed.netloc.lower()

    high_authority_domains = [
        "wikipedia.org", "github.com", "stackoverflow.com", "medium.com",
        "amazon.com", "google.com", "microsoft.com", "apple.com", "facebook.com",
        "linkedin.com", "twitter.com", "instagram.com", "youtube.com", "reddit.com",
        "wikihow.com", "thoughtco.com", "thebalance.com", "investopedia.com"
    ]

    medium_authority_domains = [
        "blogspot.com", "wordpress.com", "wix.com", "squarespace.com", "shopify.com",
        "tumblr.com", " Blogger.com", "ghost.org", "hubspot.com", "forbes.com",
        "business.com", "inc.com", "entrepreneur.com", "fastcompany.com"
    ]

    for da_domain in high_authority_domains:
        if da_domain in domain:
            return random.randint(70, 95)

    for da_domain in medium_authority_domains:
        if da_domain in domain:
            return random.randint(40, 70)

    if domain.startswith("www."):
        domain = domain[4:]

    if any(ext in domain for ext in [".edu", ".gov", ".org"]):
        return random.randint(50, 80)

    return random.randint(15, 45)


def check_title_keyword_match(url, keyword):
    keyword_lower = keyword.lower()
    domain_lower = url.lower()

    keyword_words = keyword_lower.split()
    domain_words = re.sub(r'[^\w]', ' ', domain_lower).split()

    matches = sum(1 for kw in keyword_words if kw in domain_words)
    if matches > 0:
        return min(matches / len(keyword_words), 1.0)

    return 0.3


def estimate_backlinks(url):
    parsed = urlparse(url)
    domain = parsed.netloc

    high_bl_domains = ["wikipedia.org", "github.com", "stackoverflow.com", "amazon.com"]
    medium_bl_domains = ["medium.com", "blogspot.com", "wordpress.com", "forbes.com"]

    for hb_domain in high_bl_domains:
        if hb_domain in domain:
            return random.randint(10000, 100000)

    for mb_domain in medium_bl_domains:
        if mb_domain in domain:
            return random.randint(1000, 10000)

    return random.randint(10, 500)


def calculate_difficulty_score(domain_authority, title_match, backlinks, search_volume):
    da_weight = 0.35
    title_weight = 0.25
    bl_weight = 0.25
    sv_weight = 0.15

    da_score = min(domain_authority, 100) * da_weight

    title_score = title_match * 100 * title_weight

    bl_normalized = min(backlinks / 1000, 100) if backlinks > 0 else 0
    bl_score = bl_normalized * bl_weight

    sv_normalized = min(search_volume / 10000, 100) if search_volume else 30
    sv_score = sv_normalized * sv_weight

    total_score = da_score + title_score + bl_score + sv_score

    return min(max(total_score, 0), 100)


def get_difficulty_label(score):
    if score < 25:
        return "easy"
    elif score < 50:
        return "medium"
    elif score < 75:
        return "hard"
    else:
        return "very_hard"


def get_competition_level(score):
    if score < 25:
        return "low"
    elif score < 50:
        return "medium"
    elif score < 75:
        return "high"
    else:
        return "very_high"


def handler(event):
    try:
        keyword = event.get("keyword", "")
        search_volume = event.get("search_volume")
        competing_pages = event.get("competing_pages", [])

        if not keyword:
            return {"ok": False, "error": "keyword is required"}
        if not isinstance(competing_pages, list):
            return {"ok": False, "error": "competing_pages must be a list"}

        if search_volume is None:
            search_volume = random.randint(100, 10000)

        if not competing_pages:
            competing_pages = [
                "https://www.wikipedia.org/wiki/Keyword",
                "https://www.forbes.com/keyword",
                "https://blog.example.com/keyword-guide"
            ]

        domain_authorities = []
        title_matches = []
        backlinks_list = []

        for page_url in competing_pages:
            da = estimate_domain_authority(page_url)
            domain_authorities.append(da)

            tm = check_title_keyword_match(page_url, keyword)
            title_matches.append(tm)

            bl = estimate_backlinks(page_url)
            backlinks_list.append(bl)

        avg_da = sum(domain_authorities) / len(domain_authorities) if domain_authorities else 50
        avg_title_match = sum(title_matches) / len(title_matches) if title_matches else 0.3
        avg_backlinks = sum(backlinks_list) / len(backlinks_list) if backlinks_list else 100

        difficulty_score = calculate_difficulty_score(avg_da, avg_title_match, avg_backlinks, search_volume)
        difficulty_score = round(difficulty_score, 1)

        difficulty_label = get_difficulty_label(difficulty_score)
        competition_level = get_competition_level(difficulty_score)

        return {
            "ok": True,
            "difficulty_score": difficulty_score,
            "difficulty_label": difficulty_label,
            "competition_level": competition_level,
            "keyword": keyword,
            "search_volume": search_volume,
            "competing_pages_count": len(competing_pages),
            "estimated_metrics": {
                "avg_domain_authority": round(avg_da, 1),
                "avg_title_match": round(avg_title_match, 2),
                "avg_backlinks": round(avg_backlinks, 0)
            }
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
