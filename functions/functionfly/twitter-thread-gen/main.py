"""Twitter Thread Generator - Generate Twitter/X threads."""
import hashlib
from datetime import datetime
from typing import Any


def generate_tweet(topic: str, index: int, is_first: bool, is_last: bool, include_media: bool) -> dict:
    """Generate a single tweet for a thread."""
    hash_val = int(hashlib.md5((topic + str(index) + datetime.now().isoformat()).encode()).hexdigest()[:8], 16)

    if is_first:
        templates = [
            f"🧵 A thread on {topic}:\n\nHere's what nobody tells you...",
            f"Let's talk about {topic}:\n\n(1/?)",
            f"{topic} is changing fast. Here's what you need to know:",
            f"I spent weeks researching {topic}. Let me share what I found:",
        ]
        text = templates[hash_val % len(templates)]
    elif is_last:
        templates = [
            f"That's a wrap on {topic}!\n\nLike this thread? Follow for more.\n\n#thread",
            f"Hope this helped you understand {topic} better.\n\nDrop a ❤️ if you found value.\n\n#insights",
            f"What did I miss? Share your thoughts below 👇\n\n#discussion",
            f"The key takeaway about {topic}:\n\n[Put your main insight here]\n\nThanks for reading!",
        ]
        text = templates[hash_val % len(templates)]
    else:
        templates = [
            f"({index}/) Here's the thing about {topic} most people get wrong...",
            f"({index}/) But here's what actually works:",
            f"({index}/) The real insight on {topic}:",
            f"({index}/) And the kicker about {topic}...",
            f"({index}/) What this means for you:\n\n[Actionable takeaway]",
        ]
        text = templates[hash_val % len(templates)]

    media_suggestion = None
    if include_media:
        media_options = [
            f"Screenshot of key statistics about {topic}",
            f"Infographic breaking down {topic}",
            f"Relevant image or chart",
            f"Screen recording showing {topic} in action",
            None
        ]
        media_suggestion = media_options[hash_val % len(media_options)]

    return {
        "tweet_number": index,
        "text": text,
        "media_suggestion": media_suggestion,
        "character_count": len(text)
    }


def handler(event: dict) -> dict:
    """Generate a Twitter thread."""
    try:
        topic = event.get("topic")
        num_tweets = event.get("num_tweets", 5)
        include_media_suggestion = event.get("include_media_suggestion", True)

        if not topic:
            return {"ok": False, "error": "topic is required"}
        if not isinstance(num_tweets, int) or num_tweets < 2 or num_tweets > 25:
            return {"ok": False, "error": "num_tweets must be an integer between 2 and 25"}

        if len(topic) < 3:
            return {"ok": False, "error": "topic must be at least 3 characters"}

        thread_intro = f"Starting a thread on {topic}! Here's what I learned:"

        tweets = []
        for i in range(1, num_tweets + 1):
            is_first = i == 1
            is_last = i == num_tweets

            tweet = generate_tweet(topic, i, is_first, is_last, include_media_suggestion)
            tweets.append(tweet)

        hashtags = []
        topic_clean = topic.lower().replace(" ", "")
        hashtags.append(f"#{topic_clean}")
        hashtags.append("#thread")
        hashtags.append("#tips")

        hash_val = int(hashlib.md5(topic.encode()).hexdigest()[:8], 16)
        extra_tags = ["#insights", "#learn", "#growth", "#knowledge"]
        hashtags.append(extra_tags[hash_val % len(extra_tags)])

        total_chars = sum(t["character_count"] for t in tweets)

        return {
            "ok": True,
            "topic": topic,
            "thread_intro": thread_intro,
            "tweets": tweets,
            "num_tweets": num_tweets,
            "total_characters": total_chars,
            "hashtags": hashtags,
            "thread_estimate": f"{num_tweets * 2}-{num_tweets * 3} min read",
            "generated_at": datetime.now().isoformat()
        }

    except Exception as e:
        return {"ok": False, "error": f"Failed to generate Twitter thread: {str(e)}"}


if __name__ == "__main__":
    result = handler({
        "topic": "Productivity Tips",
        "num_tweets": 5,
        "include_media_suggestion": True
    })
    print(result)
