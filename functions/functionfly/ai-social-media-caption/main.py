"""AI Social Media Caption Generator - Generate social media captions."""
import random


CAPTION_TEMPLATES = {
    "instagram": [
        "Here's what {topic} taught me about life 🌍✨\n\nSwipe to learn more!\n\n#motivation #{topic} #inspiration",
        "Life hack: {topic} edition 💡\n\nSave this for later!\n\n.{topic}\n\n#lifehacks #tips",
        "POV: You just discovered {topic} 🎯\n\n.{topic}\n\n#discover #viral #trending"
    ],
    "twitter": [
        "Hot take: {topic} is underrated.\n\nHere's why 👇\n\n#{topic}",
        "Quick tip on {topic}:\n\n1. Start\n2. Continue\n3. Profit\n\n#{topic} #tips",
        "We need to talk about {topic}.\n\n(Thread 🧵)"
    ],
    "linkedin": [
        "After years in the industry, here's what I've learned about {topic}:\n\n1. Foundation matters\n2. Consistency is key\n3. Continuous learning essential\n\nI'd love to hear your thoughts.",
        "The biggest misconception about {topic} that I see:\n\nIt's not about perfection—it's about progress.\n\nWhat's your experience with {topic}?",
        "Three things I wish I knew about {topic} when I started:\n\n• Start before you're ready\n• Learn from failures\n• Build relationships\n\nWhat's on your list?"
    ],
    "facebook": [
        "Hey everyone! I wanted to share some thoughts on {topic} today.\n\nWhat's your experience? Let me know in the comments! 👇",
        "Quick update: We've been exploring {topic} lately and wanted to share our findings with you all.",
        "Question for the group: What's your take on {topic}? Would love to hear different perspectives!"
    ]
}

HASHTAG_POOLS = {
    "instagram": ["motivation", "inspiration", "success", "goals", "mindset", "growth", "learning", "development", "lifestyle", "daily", "positive", "vibes", "entrepreneur", "business", "startup"],
    "twitter": ["tech", "news", "update", "trending", "viral", "hot", "take", "opinion", "insights", "thoughts"],
    "linkedin": ["leadership", "management", "career", "professional", "development", "networking", "business", "strategy", "growth", "success"],
    "facebook": ["community", "discussion", "share", "thoughts", "opinion", "insights", "update", "news", "friendly", "engagement"]
}

POSTING_TIMES = {
    "instagram": "9:00 AM - 11:00 AM or 7:00 PM - 9:00 PM",
    "twitter": "8:00 AM - 10:00 AM or 7:00 PM - 9:00 PM",
    "linkedin": "8:00 AM - 10:00 AM (Tuesday-Thursday best)",
    "facebook": "1:00 PM - 3:00 PM or 7:00 PM - 9:00 PM"
}

EMOJI_POOLS = {
    "instagram": ["✨", "🔥", "💡", "🎯", "💪", "🙌", "💯", "🌟", "⭐", "👇", "👉", "💬", "❤️", "🔝", "📈"],
    "twitter": ["💡", "🔥", "👇", "⬇️", "📊", "✅", "⚠️", "❌", "➡️", "🔗"],
    "linkedin": ["💼", "📈", "✅", "👉", "⬇️", "💡", "🎯", "📊", "⚡", "👍"],
    "facebook": ["👋", "💬", "🙏", "❤️", "👍", "🔥", "😊", "🙌", "💯", "🎉"]
}


def handler(event):
    try:
        content_type = event.get("content_type", "post")
        topic = event.get("topic", "")
        platform = event.get("platform", "instagram")
        include_hashtags = event.get("include_hashtags", True)

        if not topic:
            return {"ok": False, "error": "topic is required"}
        if content_type not in ["post", "story", "reel"]:
            return {"ok": False, "error": "content_type must be post, story, or reel"}
        if platform not in ["instagram", "twitter", "linkedin", "facebook"]:
            return {"ok": False, "error": "platform must be instagram, twitter, linkedin, or facebook"}

        caption_template = random.choice(CAPTION_TEMPLATES.get(platform, CAPTION_TEMPLATES["instagram"]))
        caption = caption_template.format(topic=topic)

        if content_type == "story":
            caption = f"Story time: {topic} 📱\n\n{caption}"
        elif content_type == "reel":
            caption = f"Reel on {topic} 🎬\n\n{caption}"

        if len(caption) > 2200:
            caption = caption[:2197] + "..."

        hashtags = []
        if include_hashtags:
            hashtag_pool = HASHTAG_POOLS.get(platform, HASHTAG_POOLS["instagram"])
            num_hashtags = random.randint(10, 15)
            hashtags = ["#" + topic.lower().replace(" ", "")] + random.sample(hashtag_pool, min(num_hashtags - 1, len(hashtag_pool)))

        emoji_pool = EMOJI_POOLS.get(platform, EMOJI_POOLS["instagram"])
        emoji_suggestions = random.sample(emoji_pool, min(5, len(emoji_pool)))

        best_posting_time = POSTING_TIMES.get(platform, POSTING_TIMES["instagram"])

        return {
            "ok": True,
            "caption": caption,
            "hashtags": hashtags,
            "emoji_suggestions": emoji_suggestions,
            "best_posting_time": best_posting_time,
            "platform": platform,
            "content_type": content_type
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
