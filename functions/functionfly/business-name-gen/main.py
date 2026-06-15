"""Business Name Generator - Generate creative business names."""
import re
import hashlib
from datetime import datetime
from typing import Any


STYLE_TAGLINES = {
    "professional": ["Solutions", "Services", "Associates", "Group", "Partners", "Consulting"],
    "creative": ["Studio", "Lab", "Works", "Forge", "Canvas", "Labs"],
    "modern": ["Tech", "Digital", "Smart", "Next", "Prime", "Flux"],
    "classic": ["Royal", "Premier", "Elite", "Golden", "Silver", "Heritage"]
}

STYLE_SUFFIXES = {
    "professional": ["Inc", "LLC", "Corp", "Co", "Group", "Partners"],
    "creative": ["Studios", "Collective", "House", "Agency", "Workshop"],
    "modern": ["IO", "HQ", "Lab", "Hub", "X", "Ventures"],
    "classic": ["& Sons", "& Daughters", "Enterprises", "Holdings", "Legacy"]
}

NAME_TEMPLATES = [
    "{keyword}{style_word}",
    "{style_word}{keyword}",
    "{keyword}{suffix}",
    "The {keyword} {style_word}",
    "{keyword}{style_word}{suffix}",
]


def generate_name(keyword: str, style: str, idx: int) -> dict:
    """Generate a single business name with details."""
    keyword_clean = keyword.strip().title()
    style_word = STYLE_TAGLINES.get(style, STYLE_TAGLINES["professional"])[idx % 6]
    suffix = STYLE_SUFFIXES.get(style, STYLE_SUFFIXES["professional"])[idx % 6]

    templates = [
        f"{keyword_clean}{style_word}",
        f"{style_word}{keyword_clean}",
        f"{keyword_clean} {suffix}",
        f"The {keyword_clean} {style_word}",
        f"{keyword_clean}{style_word}{suffix}" if style == "modern" else f"{keyword_clean} {suffix}",
    ]

    name = templates[idx % len(templates)]
    name = re.sub(r'\s+', ' ', name).strip()

    hash_input = f"{name}{datetime.now().isoformat()}"
    hash_val = int(hashlib.md5(hash_input.encode()).hexdigest()[:8], 16)
    available = hash_val % 3 != 0
    score = min(100, 50 + (hash_val % 50))

    taglines = {
        "professional": f"Excellence in {keyword_clean}",
        "creative": f"Where {keyword_clean} Meets Innovation",
        "modern": f"The Future of {keyword_clean}",
        "classic": f"Tradition Meets {keyword_clean}"
    }

    return {
        "name": name,
        "tagline": taglines.get(style, taglines["professional"]),
        "available": available,
        "score": score,
        "style": style
    }


def handler(event: dict) -> dict:
    """Generate business name suggestions."""
    try:
        industry = event.get("industry")
        keywords = event.get("keywords", [])
        style = event.get("style", "professional")

        if not industry:
            return {"ok": False, "error": "industry is required"}
        if not keywords:
            return {"ok": False, "error": "keywords list is required"}
        if len(keywords) == 0:
            return {"ok": False, "error": "keywords must contain at least one item"}
        if style not in ["professional", "creative", "modern", "classic"]:
            return {"ok": False, "error": "style must be one of: professional, creative, modern, classic"}

        for kw in keywords:
            if not isinstance(kw, str) or len(kw) < 2:
                return {"ok": False, "error": "each keyword must be a string with at least 2 characters"}

        names = []
        seen = set()
        idx = 0

        while len(names) < 10 and idx < 100:
            kw = keywords[idx % len(keywords)]
            name_data = generate_name(kw, style, idx)
            if name_data["name"] not in seen:
                seen.add(name_data["name"])
                names.append(name_data)
            idx += 1

        industry_lower = industry.lower().replace(" ", "_")
        generated_at = datetime.now().isoformat()

        return {
            "ok": True,
            "industry": industry,
            "style": style,
            "names": names,
            "count": len(names),
            "generated_at": generated_at
        }

    except Exception as e:
        return {"ok": False, "error": f"Failed to generate business names: {str(e)}"}


if __name__ == "__main__":
    result = handler({
        "industry": "Technology Consulting",
        "keywords": ["cloud", "data", "smart"],
        "style": "modern"
    })
    print(result)
