"""AI Product Description Generator - Generate product descriptions."""
import random


PERSUASIVE_TEMPLATES = [
    "Transform your {target_audience} experience with {product_name}. This isn't just another product—it's a game-changer designed specifically for those who demand excellence.",
    "Stop searching and start winning. {product_name} delivers exactly what {target_audience} need: reliability, performance, and results that speak for themselves.",
    "Join thousands of satisfied {target_audience} who chose {product_name} and never looked back. Your transformation starts today."
]

DESCRIPTIVE_TEMPLATES = [
    "{product_name} is a product designed for {target_audience}. It offers key features that address common needs in its category.",
    "This product provides essential functionality for {target_audience}. It is built with standard industry practices and quality materials.",
    "{product_name} serves the needs of {target_audience} with a focus on practical, everyday use cases."
]

SHORT_DESC_TEMPLATES = [
    "The essential {product_name} for {target_audience}. Quality meets practicality.",
    "{product_name}: Built for {target_audience} who demand the best.",
    "Discover {product_name}—the smart choice for {target_audience}."
]

FEATURE_PREFIXES = [
    "Experience", "Enjoy", "Get", "Discover", "Unlock"
]


def handler(event):
    try:
        product_name = event.get("product_name", "")
        features = event.get("features", [])
        target_audience = event.get("target_audience", "everyday users")
        tone = event.get("tone", "persuasive")

        if not product_name:
            return {"ok": False, "error": "product_name is required"}
        if tone not in ["persuasive", "descriptive"]:
            return {"ok": False, "error": "tone must be persuasive or descriptive"}

        if not features:
            features = [
                "Easy to use interface",
                "High-quality materials",
                "Affordable pricing",
                "Reliable performance",
                "Excellent customer support"
            ]

        short_template = random.choice(SHORT_DESC_TEMPLATES)
        short_description = short_template.format(
            product_name=product_name,
            target_audience=target_audience
        )

        if tone == "persuasive":
            full_template = random.choice(PERSUASIVE_TEMPLATES)
        else:
            full_template = random.choice(DESCRIPTIVE_TEMPLATES)

        full_description = full_template.format(
            product_name=product_name,
            target_audience=target_audience
        )

        bullets = []
        for i, feature in enumerate(features[:5]):
            prefix = FEATURE_PREFIXES[i] if i < len(FEATURE_PREFIXES) else "Features"
            bullets.append(f"{prefix} {feature.lower()}")

        keywords = [
            product_name.lower(),
            target_audience.lower(),
            product_name.lower().replace(" ", "_"),
            "quality",
            "reliable",
            "premium",
            "essential",
            "innovative"
        ]

        return {
            "ok": True,
            "short_description": short_description,
            "full_description": full_description,
            "bullets": bullets,
            "keywords": keywords,
            "product_name": product_name,
            "tone": tone
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
