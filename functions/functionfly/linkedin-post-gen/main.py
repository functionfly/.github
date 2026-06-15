"""LinkedIn Post Generator - Generate professional LinkedIn posts."""
import hashlib
from datetime import datetime
from typing import Any


TONE_CONFIGS = {
    "professional": {
        "style": "Professional and business-focused",
        "opening_templates": [
            "I recently came across an insight that changed my perspective on {topic}.",
            "After years of working in this space, here's what I've learned about {topic}.",
            "Something I've been thinking about lately: {topic}.",
        ],
        "cta_options": ["Share your thoughts in the comments", "Let's connect to discuss further", "Drop your perspective below"]
    },
    "thought_leader": {
        "style": "Industry thought leadership",
        "opening_templates": [
            "The future of {topic} is being shaped by those willing to challenge conventions.",
            "Hot take: Most people are thinking about {topic} wrong.",
            "What if everything we knew about {topic} was just the beginning?",
        ],
        "cta_options": ["I'd love to hear your counterpoints", "Who else is seeing this trend?", "Tag someone who needs to see this"]
    },
    "humble": {
        "style": "Humble and authentic",
        "opening_templates": [
            "Still learning about {topic}, but here's what's resonated with me so far.",
            "I'm not an expert, but I wanted to share my experience with {topic}.",
            "A reminder to myself (and maybe you) about {topic}.",
        ],
        "cta_options": ["What would you add?", "I'm still figuring this out - would love advice", "What have you learned?"]
    }
}


def generate_post(topic: str, tone: str, include_statistics: bool, call_to_action: str = None) -> str:
    """Generate a LinkedIn post."""
    config = TONE_CONFIGS.get(tone, TONE_CONFIGS["professional"])

    hash_val = int(hashlib.md5((topic + tone + datetime.now().isoformat()).encode()).hexdigest()[:8], 16)

    opening_templates = config["opening_templates"]
    opening_template = opening_templates[hash_val % len(opening_templates)]
    opening = opening_template.format(topic=topic)

    stats_block = ""
    if include_statistics:
        stats = [
            f"📊 Companies prioritizing {topic} have seen a 47% increase in efficiency",
            f"📈 73% of professionals say {topic} is critical to their success",
            f"🎯 Organizations focusing on {topic} outperform peers by 3x",
            f"💡 Teams with strong {topic} foundations are 2x more likely to innovate"
        ]
        stats_block = f"\n\n{stats[hash_val % len(stats)]}"

    body_options = {
        "professional": [
            "The key insight here is that sustainable success requires both strategic vision and operational excellence.\n\nWhat aspect of this resonates most with your experience?",
            "I've seen firsthand how prioritizing this approach creates ripple effects throughout an organization.\n\nWould love to hear how others are approaching this.",
            "The data continues to support what many leaders have suspected: this matters more than ever.\n\nDrop a comment if you're navigating similar challenges."
        ],
        "thought_leader": [
            "Here's the uncomfortable truth: the traditional approach isn't working anymore.\n\nWe're at an inflection point where adaptation isn't optional - it's survival.\n\nThe organizations that embrace this shift will define the next decade of our industry.\n\nWho's with me?",
            "Let me break this down:\n\n1. The old playbook is broken\n2. The window for transformation is now\n3. Those who act decisively will lead\n\nThis isn't about keeping up. It's about setting the pace.\n\nChallenge conventional wisdom. Lead fearlessly.",
        ],
        "humble": [
            "I wanted to share this because open conversation helps us all grow.\n\nStill learning every day, and I appreciate the community here that makes space for genuine dialogue.\n\nWhat's been your experience?",
            "We're all works in progress, aren't we?\n\nHere's something I'm still processing about all this. Would genuinely love to hear your perspectives.",
            "No grand conclusions here - just some thoughts I've been sitting with.\n\nSometimes the most valuable thing we can do is share our ongoing journey.",
        ]
    }

    body_list = body_options.get(tone, body_options["professional"])
    body = body_list[hash_val % len(body_list)]

    if call_to_action:
        cta = call_to_action
    else:
        cta_options = config["cta_options"]
        cta = cta_options[hash_val % len(cta_options)]

    full_post = f"{opening}{stats_block}\n\n{body}\n\n{cta}"

    if len(full_post) > 3000:
        full_post = full_post[:2997] + "..."

    return full_post


def generate_hashtags(topic: str, count: int = 5) -> list:
    """Generate relevant hashtags."""
    topic_clean = topic.lower().replace(" ", "")
    hashtags = [
        f"#{topic_clean}",
        "#leadership",
        "#professionaldevelopment",
        "#growth",
        "#thoughtleadership",
        "#innovation",
        "#strategy",
        "#careers"
    ]

    hash_val = int(hashlib.md5(topic.encode()).hexdigest()[:8], 16)
    selected = []
    for i in range(count):
        idx = (hash_val + i) % len(hashtags)
        tag = hashtags[idx]
        if tag not in selected:
            selected.append(tag)

    return selected[:count]


def handler(event: dict) -> dict:
    """Generate a LinkedIn post."""
    try:
        topic = event.get("topic")
        include_statistics = event.get("include_statistics", False)
        tone = event.get("tone", "professional")
        call_to_action = event.get("call_to_action")

        if not topic:
            return {"ok": False, "error": "topic is required"}
        if tone not in ["professional", "thought_leader", "humble"]:
            return {"ok": False, "error": "tone must be one of: professional, thought_leader, humble"}

        if len(topic) < 3:
            return {"ok": False, "error": "topic must be at least 3 characters"}

        post_text = generate_post(topic, tone, include_statistics, call_to_action)
        hashtags = generate_hashtags(topic, 4)

        engagement_tips = [
            "Post during business hours (Tuesday-Thursday, 8-10am is optimal)",
            "Engage with comments in the first hour to boost algorithm visibility",
            "Ask open-ended questions to encourage meaningful comments",
            "Add a relevant image or document to increase engagement by 2x",
            "Respond to all comments within 24 hours"
        ]

        return {
            "ok": True,
            "topic": topic,
            "tone": tone,
            "post_text": post_text,
            "character_count": len(post_text),
            "hashtags": hashtags,
            "engagement_tips": engagement_tips,
            "generated_at": datetime.now().isoformat()
        }

    except Exception as e:
        return {"ok": False, "error": f"Failed to generate LinkedIn post: {str(e)}"}


if __name__ == "__main__":
    result = handler({
        "topic": "Remote Work Culture",
        "include_statistics": True,
        "tone": "thought_leader"
    })
    print(result)
