"""Logo Concept Generator - Generate logo design concepts."""
import hashlib
from datetime import datetime
from typing import Any


STYLE_CONFIGS = {
    "modern": {
        "description": "Clean, contemporary designs with bold shapes and minimalist elements",
        "typography": ["Sans-serif with geometric construction", "Custom geometric letterforms", "Clean uppercase treatments"],
        "iconography": ["Abstract geometric shapes", "Negative space designs", "Dynamic angles and lines"],
        "color_suggestions": ["Bold primaries", "Neon accents on dark", "Gradient blues and purples"]
    },
    "classic": {
        "description": "Timeless, established aesthetics that convey trust and heritage",
        "typography": ["Elegant serif typography", "Traditional letterforms", "Refined script treatments"],
        "iconography": ["Classic emblems and shields", "Traditional crests", "Established symbols"],
        "color_suggestions": ["Navy and gold", "Deep burgundy and cream", "Classic black and white"]
    },
    "playful": {
        "description": "Fun, approachable designs with personality and charm",
        "typography": ["Rounded sans-serif", "Hand-drawn styles", "Bubbly letterforms"],
        "iconography": ["Whimsical illustrations", "Character-based icons", "Colorful abstract shapes"],
        "color_suggestions": ["Bright pastels", "Rainbow palettes", "Vibrant single colors"]
    },
    "minimalist": {
        "description": "Simple, elegant designs with focus on essential elements",
        "typography": ["Ultra-thin sans-serif", "Single-weight letterforms", "Generous spacing"],
        "iconography": ["Simple line icons", "Single-color marks", "Geometric minimalism"],
        "color_suggestions": ["Black on white", "Single accent color", "Monochromatic schemes"]
    }
}


def generate_concept(brand_name: str, industry: str, style: str, color_pref: str, idx: int) -> dict:
    """Generate a single logo concept."""
    config = STYLE_CONFIGS.get(style, STYLE_CONFIGS["modern"])

    hash_val = int(hashlib.md5((brand_name + industry + str(idx)).encode()).hexdigest()[:8], 16)

    industry_keywords = {
        "technology": ["pixel", "circuit", "cloud", "data", "network"],
        "food": ["leaf", "flame", "drop", "grain", "harvest"],
        "healthcare": ["cross", "heart", "pulse", "shield", "care"],
        "finance": ["chart", "column", "shield", "key", "globe"],
        "retail": ["tag", "bag", "star", "cart", "badge"],
        "education": ["book", "light", "path", "star", "growth"],
        "creative": ["brush", "palette", "light", "star", "spark"],
        "default": ["shield", "star", "light", "key", "spark"]
    }

    ind_lower = industry.lower()
    icon_keywords = industry_keywords.get(ind_lower, industry_keywords["default"])

    icon_keyword = icon_keywords[idx % len(icon_keywords)]

    concept_templates = [
        {
            "name": f"The {icon_keyword.title()} Mark",
            "description": f"A stylized {icon_keyword} combined with the brand initial, creating a memorable monogram that works across all applications.",
            "iconography": f"Abstract {icon_keyword} formed by intersecting geometric shapes",
            "colors": config["color_suggestions"][idx % len(config["color_suggestions"])].split(", ")
        },
        {
            "name": "Wordmark + Symbol",
            "description": f"A clean wordmark paired with a minimalist {icon_keyword} symbol that can stand alone in smaller applications.",
            "iconography": f"Simple line-art {icon_keyword} with consistent stroke weight",
            "colors": config["color_suggestions"][(idx + 1) % len(config["color_suggestions"])].split(", ")
        },
        {
            "name": "The Negative Space",
            "description": f"A clever design where the {icon_keyword} is formed through negative space, creating an instantly memorable mark.",
            "iconography": f"Shape revealing {icon_keyword} through strategic negative space",
            "colors": config["color_suggestions"][(idx + 2) % len(config["color_suggestions"])].split(", ")
        },
        {
            "name": "Modern Monogram",
            "description": f"An interlocking letterform combining the first letters of the brand name into a contemporary {icon_keyword}-inspired shape.",
            "iconography": f"Interlocking geometric forms suggesting {icon_keyword}",
            "colors": config["color_suggestions"][(idx + 3) % len(config["color_suggestions"])].split(", ")
        },
        {
            "name": "The Abstract Mark",
            "description": f"A completely abstract mark that evokes {icon_keyword} through pure form, ensuring broad applicability and timeless appeal.",
            "iconography": f"Pure geometric abstraction representing {icon_keyword}",
            "colors": config["color_suggestions"][(idx + 4) % len(config["color_suggestions"])].split(", ")
        }
    ]

    concept = concept_templates[idx % len(concept_templates)]

    if color_pref:
        concept["colors"] = [color_pref] + concept["colors"][:2]

    return {
        "name": concept["name"],
        "description": concept["description"],
        "color_palette": concept["colors"],
        "typography_suggestion": config["typography"][idx % len(config["typography"])],
        "iconography": concept["iconography"],
        "versatility_score": 90 - (idx * 5)
    }


def handler(event: dict) -> dict:
    """Generate logo concepts."""
    try:
        brand_name = event.get("brand_name")
        industry = event.get("industry")
        style_preference = event.get("style_preference", "modern")
        color_preference = event.get("color_preference")

        if not brand_name:
            return {"ok": False, "error": "brand_name is required"}
        if not industry:
            return {"ok": False, "error": "industry is required"}
        if style_preference not in ["modern", "classic", "playful", "minimalist"]:
            return {"ok": False, "error": "style_preference must be one of: modern, classic, playful, minimalist"}

        if len(brand_name) < 2:
            return {"ok": False, "error": "brand_name must be at least 2 characters"}

        config = STYLE_CONFIGS[style_preference]

        concepts = []
        for i in range(5):
            concepts.append(generate_concept(brand_name, industry, style_preference, color_preference, i))

        return {
            "ok": True,
            "brand_name": brand_name,
            "industry": industry,
            "style_preference": style_preference,
            "style_description": config["description"],
            "concepts": concepts,
            "generated_at": datetime.now().isoformat()
        }

    except Exception as e:
        return {"ok": False, "error": f"Failed to generate logo concepts: {str(e)}"}


if __name__ == "__main__":
    result = handler({
        "brand_name": "TechNova",
        "industry": "Technology",
        "style_preference": "modern",
        "color_preference": "#0066FF"
    })
    print(result)
