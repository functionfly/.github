"""AI Ad Copy Generator - Generate ad copy for various platforms."""
import random
import re


POSITIVE_WORDS = {
    "professional": ["trusted", "expert", "professional", "certified", "proven"],
    "friendly": ["friendly", "easy", "simple", "quick", "fast", "fun"],
    "urgent": ["now", "today", "limited", "hurry", "act fast", "don't miss"],
    "playful": ["awesome", "amazing", "love", "wow", "cool", "amazing"]
}

CTA_OPTIONS = {
    "google": ["Learn More", "Get Started", "Sign Up Now", "Get a Quote", "Contact Us"],
    "facebook": ["Learn More", "Shop Now", "Sign Up", "Book Now", "Get Offer"],
    "instagram": ["Shop Now", "Tap to Shop", "Link in Bio", "Get Yours", "Explore"]
}

HEADLINE_TEMPLATES = {
    "google": [
        "{product} - Trusted by Professionals",
        "Get {product} Today",
        "Best {product} Solutions",
        "{product} for {audience}",
        "Transform Your {product}"
    ],
    "facebook": [
        "Discover {product}",
        "Amazing {product} Deals",
        "Get {product} Now",
        "You Need {product}",
        "Try {product} Free"
    ],
    "instagram": [
        "{product} ✨",
        "Glow with {product}",
        "Love {product}",
        "{product} Goals",
        "Must Have {product}"
    ]
}

DESCRIPTION_TEMPLATES = {
    "google": [
        "Join thousands who trust {product}. {audience} love it.",
        "Get started in minutes. {audience} preferred choice.",
        "Professional {product} solutions. Free trial available.",
        "Save time and money with {product}. {audience} recommend."
    ],
    "facebook": [
        "Discover why {audience} are switching to {product}.",
        "Join our community of happy {audience} today!",
        "Limited time offer on {product}. Don't miss out!",
        "See what {audience} are saying about {product}."
    ],
    "instagram": [
        "Your daily dose of {product} ✨ {audience} approved!",
        "Living for this {product}! 🤩 {audience} needed this.",
        "{product} is everything {audience} dreamed of 💫",
        "Obsessed with {product} and so is {audience}!"
    ]
}


def handler(event):
    try:
        product_name = event.get("product_name", "")
        product_description = event.get("product_description", "")
        target_audience = event.get("target_audience", "everyone")
        platform = event.get("platform", "google")
        tone = event.get("tone", "professional")

        if not product_name:
            return {"ok": False, "error": "product_name is required"}
        if platform not in ["google", "facebook", "instagram"]:
            return {"ok": False, "error": "platform must be google, facebook, or instagram"}
        if tone not in ["professional", "friendly", "urgent", "playful"]:
            return {"ok": False, "error": "tone must be professional, friendly, urgent, or playful"}

        ads = []
        used_ctas = []
        positive_words = POSITIVE_WORDS.get(tone, POSITIVE_WORDS["professional"])
        headline_templates = HEADLINE_TEMPLATES.get(platform, HEADLINE_TEMPLATES["google"])
        desc_templates = DESCRIPTION_TEMPLATES.get(platform, DESCRIPTION_TEMPLATES["google"])
        cta_options = CTA_OPTIONS.get(platform, CTA_OPTIONS["google"])

        for i in range(3):
            headline = headline_templates[i % len(headline_templates)].format(
                product=product_name,
                audience=target_audience
            )
            if len(headline) > 30:
                headline = headline[:27] + "..."

            description = desc_templates[i % len(desc_templates)].format(
                product=product_name,
                audience=target_audience
            )
            if len(description) > 90:
                description = description[:87] + "..."

            available_ctas = [c for c in cta_options if c not in used_ctas]
            if not available_ctas:
                available_ctas = cta_options
            cta = random.choice(available_ctas)
            used_ctas.append(cta)

            ads.append({
                "headline": headline,
                "description": description,
                "cta": cta
            })

        return {
            "ok": True,
            "ads": ads,
            "platform": platform,
            "tone": tone,
            "product_name": product_name
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
