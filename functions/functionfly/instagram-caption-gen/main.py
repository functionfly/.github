"""Instagram Caption Generator - Generate engaging Instagram captions."""
import random
from datetime import datetime
from typing import Any


EMOJI_LAYOUTS = {
    "top": [
        "✨ 📸 ✨",
        "━━━━━✦━━━━━",
        "━━━━━━━━━━━━━━",
        "▔▔▔▔▔▔▔▔▔▔▔",
    ],
    "inline": [
        "•",
        "○",
        "◉",
        "✧",
        "→",
    ],
    "bottom": [
        "━━━━━━━━━━━━━━",
        "━━━━━✦━━━━━",
        "♡ · ♡ · ♡",
        "✦ · ✦ · ✦",
    ]
}


def generate_caption_hook(image_description: str, goal: str) -> str:
    """Generate attention-grabbing hook."""
    hooks = {
        "engagement": [
            "This changed everything...",
            "Nobody talks about this but...",
            "I have thoughts...",
            "Lettle truth bomb...",
            "Controversial take ahead...",
        ],
        "sales": [
            "If you struggle with {topic}, this is for you.",
            "The secret nobody tells you about {topic}...",
            "Stop scrolling - this solves your {topic} problem.",
            "Here's why {topic} matters more than you think.",
        ],
        "awareness": [
            "Today I learned something fascinating about...",
            "Let's talk about something important...",
            "This is what {topic} looks like in real life.",
            "Understanding {topic} changed my perspective.",
        ]
    }

    topic = image_description.split()[0] if image_description else "this"
    hook_list = hooks.get(goal, hooks["engagement"])
    hook = random.choice(hook_list)

    if "{topic}" in hook:
        hook = hook.format(topic=topic.lower())

    return hook


def generate_main_caption(hook: str, image_description: str, goal: str) -> str:
    """Generate main caption body."""
    topic = image_description.split()[0] if image_description else "this"

    bodies = {
        "engagement": [
            f"What are your thoughts? Drop them below 👇\n\nSave this for later 📌",
            f"Double-tap if you relate ❤️\n\nTag someone who needs to see this!",
            f"Let me know in the comments what you think! 💬\n\nFollow for more.",
        ],
        "sales": [
            f"Swipe left to learn more ⬅️\n\nLink in bio for full details 🔗",
            f"Ready to get started? DM me or click the link in bio! 📩",
            f"Send me a message to learn more about how we can help! 💬",
        ],
        "awareness": [
            f"Knowledge is power 📚\n\nFollow for more content like this!",
            f"Share your experience in the comments below 👇",
            f"Have you ever experienced this? Let me know! 💬",
        ]
    }

    body_list = bodies.get(goal, bodies["engagement"])
    body = random.choice(body_list)

    return f"{hook}\n\n{body}"


def generate_hashtags(image_description: str, goal: str) -> list:
    """Generate relevant hashtags."""
    words = image_description.lower().split()[:3]

    goal_tags = {
        "engagement": ["fyp", "explore", "viral", "trending", "community"],
        "sales": ["smallbusiness", "shoplocal", "support", "entrepreneur", "business"],
        "awareness": ["education", "learn", "knowledge", "information", "facts"]
    }

    hashtags = [f"#{word}" for word in words if len(word) > 2]
    hashtags += goal_tags.get(goal, goal_tags["engagement"])
    hashtags += ["lifestyle", "photooftheday", "instagood"]

    return list(set(hashtags))[:20]


def handler(event: dict) -> dict:
    """Generate an Instagram caption."""
    try:
        image_description = event.get("image_description")
        goal = event.get("goal", "engagement")
        include_emoji = event.get("include_emoji", True)

        if not image_description:
            return {"ok": False, "error": "image_description is required"}
        if goal not in ["engagement", "sales", "awareness"]:
            return {"ok": False, "error": "goal must be one of: engagement, sales, awareness"}

        if len(image_description) < 3:
            return {"ok": False, "error": "image_description must be at least 3 characters"}

        topic = image_description.split()[0]

        hook = generate_caption_hook(image_description, goal)
        main_caption = generate_main_caption(hook, image_description, goal)

        top_emoji = random.choice(EMOJI_LAYOUTS["top"]) if include_emoji else ""
        bottom_emoji = random.choice(EMOJI_LAYOUTS["bottom"]) if include_emoji else ""

        caption_hook = f"{top_emoji}\n{hook}" if include_emoji else hook

        full_caption = f"{caption_hook}\n\n{main_caption}\n\n{bottom_emoji}"

        if len(full_caption) > 2200:
            full_caption = full_caption[:2197] + "..."

        hashtags = generate_hashtags(image_description, goal)

        emoji_layout_suggestion = {
            "top": "Place decorative emojis or dividers at the top to draw attention",
            "inline": "Use single emojis or symbols between sentences for visual breaks",
            "bottom": "Add closing emojis or dividers to signal the end and encourage engagement"
        }

        return {
            "ok": True,
            "caption_hook": hook[:150],
            "caption": full_caption,
            "caption_length": len(full_caption),
            "hashtags": hashtags,
            "emoji_layout": emoji_layout_suggestion,
            "goal": goal,
            "generated_at": datetime.now().isoformat()
        }

    except Exception as e:
        return {"ok": False, "error": f"Failed to generate Instagram caption: {str(e)}"}


if __name__ == "__main__":
    result = handler({
        "image_description": "Sunset beach photography",
        "goal": "engagement",
        "include_emoji": True
    })
    print(result)
