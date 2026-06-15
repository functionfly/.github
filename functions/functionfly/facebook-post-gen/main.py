"""Facebook Post Generator - Generate engaging Facebook posts."""
import re
import hashlib
from datetime import datetime
from typing import Any


TONE_TEMPLATES = {
    "funny": {
        "openers": ["You'll never guess what happened...", "Plot twist: ", "Breaking news: ", "My face when ", "Reality check: "],
        "style": "humorous, light-hearted, shareable"
    },
    "inspirational": {
        "openers": ["Remember: ", "Here's the truth about ", "The secret to ", "Today I learned that ", "Success isn't about "],
        "style": "uplifting, motivational, thoughtful"
    },
    "educational": {
        "openers": ["Did you know? ", "Here's how to ", "The science behind ", "5 things you should know about ", "Understanding "],
        "style": "informative, valuable, professional"
    },
    "promotional": {
        "openers": ["Excited to announce ", "Introducing ", "Big news: ", "Limited time: ", "Special offer: "],
        "style": "compelling, action-oriented, clear"
    }
}


def generate_hashtags(topic: str, count: int = 5) -> list:
    """Generate relevant hashtags for a topic."""
    topic_clean = topic.lower().replace(" ", "")
    base_tags = [
        f"#{topic_clean}",
        f"#{topic.replace(' ', '')}",
        "#business",
        "#success",
        "#growth"
    ]

    modifiers = ["tips", "strategy", "goals", "motivation", "mindset", "learning", "development"]
    extra_tags = [f"#{topic_clean}{mod}" for mod in modifiers[:count]]

    all_tags = base_tags + extra_tags
    hash_input = f"{topic}{datetime.now().isoformat()}"
    hash_val = int(hashlib.md5(hash_input.encode()).hexdigest()[:8], 16)

    selected = []
    for i in range(count):
        idx = (hash_val + i) % len(all_tags)
        tag = all_tags[idx]
        if tag not in selected:
            selected.append(tag)

    return selected[:count]


def generate_image_suggestion(topic: str) -> str:
    """Generate image suggestion based on topic."""
    suggestions = [
        f"High-quality photo of {topic} in professional setting",
        f"Infographic explaining {topic} with clean design",
        f"Team collaboration image representing {topic}",
        f"Abstract concept visualization of {topic}",
        f"Before/after comparison related to {topic}",
        f"Step-by-step visual guide for {topic}",
        f"Customer testimonial graphic about {topic}",
        f"Product/service demo screenshot for {topic}"
    ]

    hash_val = int(hashlib.md5(topic.encode()).hexdigest()[:8], 16)
    return suggestions[hash_val % len(suggestions)]


def handler(event: dict) -> dict:
    """Generate a Facebook post."""
    try:
        topic = event.get("topic")
        include_image_suggestion = event.get("include_image_suggestion", False)
        include_cta = event.get("include_cta", False)
        tone = event.get("tone", "promotional")

        if not topic:
            return {"ok": False, "error": "topic is required"}
        if tone not in ["funny", "inspirational", "educational", "promotional"]:
            return {"ok": False, "error": "tone must be one of: funny, inspirational, educational, promotional"}

        tone_config = TONE_TEMPLATES[tone]

        hash_val = int(hashlib.md5((topic + datetime.now().isoformat()).encode()).hexdigest()[:8], 16)
        opener_idx = hash_val % len(tone_config["openers"])
        opener = tone_config["openers"][opener_idx]

        body_templates = {
            "funny": [
                "Because life is too short not to laugh at ourselves. 😄\n\nWhat's your take? Drop it in the comments! 👇",
                "Sometimes the best lessons come from the most unexpected places. 🙃\n\nTag someone who needs to see this!",
                "The struggle is real, am I right? 😅\n\nDouble-tap if you agree! ❤️"
            ],
            "inspirational": [
                "That feeling when it all finally clicks. ✨\n\nKeep pushing forward. Your breakthrough is coming.\n\nSave this for when you need a reminder. 🔖",
                "The journey of 1000 miles begins with a single step. 🌟\n\nWhat's your first step today?",
                "Your only limit is your mind. 🚀\n\nShare this with someone who needs motivation today. 💪"
            ],
            "educational": [
                "Here's something that changed my perspective on this topic:\n\n1. Start with why\n2. Focus on value\n3. Measure what matters\n\nSave this post for later! 🔖\n\nWhat would you add? Comment below! 👇",
                "Quick tip that'll save you hours:\n\nRead the full guide in our bio link. 🔗\n\nFollow for more insights like this!",
                "This is worth bookmarking:\n\nShare with your team and let's discuss in the comments! 💬"
            ],
            "promotional": [
                "🎉 Available NOW!\n\nCheck out [Product/Service] - link in bio!\n\nLimited time offer. Don't miss out! ⏰",
                "📢 NEW: [Announcement]\n\nReady to [benefit]? Click the link below to get started.\n\nWhat are you waiting for? 🌟",
                "Special offer just for you!\n\n[Offer details]\n\nDM us or click the link in bio to learn more! 💬"
            ]
        }

        body_idx = (hash_val + 1) % len(body_templates[tone])
        body = body_templates[tone][body_idx]

        full_text = f"{opener} {topic}\n\n{body}"

        if len(full_text) > 630:
            full_text = full_text[:627] + "..."

        hashtags = generate_hashtags(topic, 5)

        cta_text = None
        if include_cta:
            cta_options = [
                "Learn More",
                "Shop Now",
                "Sign Up",
                "Get Started",
                "Book Now",
                "Contact Us"
            ]
            cta_text = cta_options[hash_val % len(cta_options)]
            full_text += f"\n\n🔗 {cta_text} (link in bio)"

        image_suggestion = None
        if include_image_suggestion:
            image_suggestion = generate_image_suggestion(topic)

        return {
            "ok": True,
            "topic": topic,
            "tone": tone,
            "post_text": full_text,
            "character_count": len(full_text),
            "image_suggestion": image_suggestion,
            "hashtag_suggestions": hashtags,
            "cta_text": cta_text,
            "generated_at": datetime.now().isoformat()
        }

    except Exception as e:
        return {"ok": False, "error": f"Failed to generate Facebook post: {str(e)}"}


if __name__ == "__main__":
    result = handler({
        "topic": "Digital Marketing Success",
        "include_image_suggestion": True,
        "include_cta": True,
        "tone": "inspirational"
    })
    print(result)
