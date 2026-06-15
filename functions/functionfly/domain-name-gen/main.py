"""Domain Name Generator - Generate domain name suggestions."""
import hashlib
from datetime import datetime
from typing import Any


SUFFIXES = {
    "com": {"popularity": "highest", "memorability": "high"},
    "net": {"popularity": "high", "memorability": "medium"},
    "org": {"popularity": "medium", "memorability": "medium"},
    "io": {"popularity": "growing", "memorability": "high"}
}


def generate_domain_name(keywords: list, style: str, suffix: str, idx: int) -> dict:
    """Generate a single domain name with scoring."""
    keywords = [k.lower().strip() for k in keywords if k]

    hash_val = int(hashlib.md5((",".join(keywords) + str(idx) + datetime.now().isoformat()).encode()).hexdigest()[:8], 16)

    if style == "brandable":
        patterns = [
            lambda k: f"{k[0] if k else 'x'}{''.join(k[1:]) if len(k) > 1 else 'app'}",
            lambda k: f"{''.join(k)}{suffix if idx % 2 == 0 else 'ly'}",
            lambda k: f"{k[0] if k else 'my'}{k[-1] if len(k) > 1 else 'app'}",
            lambda k: f"get{k[0].upper() if k else 'X'}{''.join(k[1:]) if len(k) > 1 else ''}",
            lambda k: f"{''.join(k[:2])}{suffix if idx % 3 == 0 else 'co'}",
        ]
        pattern = patterns[idx % len(patterns)]
        name = pattern(keywords)
    elif style == "keyword":
        if idx % 3 == 0:
            name = f"{keywords[0] if keywords else 'search'}{keywords[-1] if len(keywords) > 1 else 'hub'}"
        elif idx % 3 == 1:
            name = f"{keywords[0] if keywords else 'find'}-{keywords[-1] if len(keywords) > 1 else 'all'}"
        else:
            name = f"{keywords[0] if keywords else 'get'}{keywords[-1] if len(keywords) > 1 else 'now'}"
    else:
        if len(keywords) >= 2:
            name = f"{keywords[0]}-{keywords[1]}"
        else:
            name = keywords[0] if keywords else "domain"

    name = name.lower().replace(" ", "-").replace("_", "-")
    while "--" in name:
        name = name.replace("--", "-")
    name = name.strip("-")

    domain = f"{name}.{suffix}"

    hash_score = int(hashlib.md5(domain.encode()).hexdigest()[:4], 16)
    base_score = (hash_score % 10000) / 100

    length_penalty = max(0, (len(name) - 10) * 3) if len(name) > 10 else 0
    score = max(0, min(100, base_score - length_penalty))

    available = hash_score % 5 != 0

    return {
        "name": domain,
        "base_name": name,
        "available": available,
        "score": round(score, 1),
        "suffix": suffix,
        "style": style
    }


def handler(event: dict) -> dict:
    """Generate domain name suggestions."""
    try:
        keywords = event.get("keywords", [])
        suffix = event.get("suffix", "com")
        style = event.get("style", "brandable")

        if not keywords or len(keywords) == 0:
            return {"ok": False, "error": "keywords list is required and must not be empty"}

        for kw in keywords:
            if not isinstance(kw, str) or len(kw) < 1:
                return {"ok": False, "error": "each keyword must be a non-empty string"}

        if suffix not in SUFFIXES:
            return {"ok": False, "error": f"suffix must be one of: {', '.join(SUFFIXES.keys())}"}

        if style not in ["brandable", "keyword", "tld-specific"]:
            return {"ok": False, "error": "style must be one of: brandable, keyword, tld-specific"}

        seen = set()
        domains = []

        for i in range(100):
            domain_data = generate_domain_name(keywords, style, suffix, i)
            if domain_data["name"] not in seen and len(domain_data["name"]) <= 30:
                seen.add(domain_data["name"])
                domains.append(domain_data)
                if len(domains) >= 20:
                    break

        domains.sort(key=lambda x: x["score"], reverse=True)

        premium_keywords = ["ai", "cloud", "tech", "data", "smart", "pro", "app", "hub"]
        premium_domains = []

        for kw in premium_keywords:
            for sfx in ["com", "io", "ai"]:
                name = f"{kw}.{sfx}"
                if name not in seen:
                    hash_val = int(hashlib.md5(name.encode()).hexdigest()[:6], 16)
                    premium_domains.append({
                        "name": name,
                        "premium_score": round(50 + (hash_val % 50), 1),
                        "available": hash_val % 3 == 0,
                        "reason": f"{kw.capitalize()} is a high-value keyword in tech"
                    })
                    seen.add(name)

        premium_domains.sort(key=lambda x: x["premium_score"], reverse=True)
        premium_domains = premium_domains[:3]

        return {
            "ok": True,
            "keywords": keywords,
            "style": style,
            "suffix": suffix,
            "domains": domains,
            "premium_domains": premium_domains,
            "generated_at": datetime.now().isoformat()
        }

    except Exception as e:
        return {"ok": False, "error": f"Failed to generate domain names: {str(e)}"}


if __name__ == "__main__":
    result = handler({
        "keywords": ["cloud", "data", "analytics"],
        "suffix": "com",
        "style": "brandable"
    })
    print(result)
