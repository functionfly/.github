"""Podcast Outline Generator - Generate podcast episode outlines."""
from datetime import datetime
from typing import Any


SEGMENT_TEMPLATES = {
    "intro": {
        "title": "Introduction",
        "duration_minutes": 3,
        "talking_points": [
            "Welcome and greeting",
            "Episode topic introduction",
            "Brief host background on topic",
            "Preview of what's coming"
        ]
    },
    "main": [
        {
            "title": "Context & Background",
            "duration_minutes": 7,
            "talking_points": [
                "Set the stage for the topic",
                "Why this matters now",
                "Common misconceptions",
                "Key definitions needed"
            ]
        },
        {
            "title": "Deep Dive",
            "duration_minutes": 10,
            "talking_points": [
                "Core insights and analysis",
                "Supporting examples or case studies",
                "Relevant data or research",
                "Practical applications"
            ]
        },
        {
            "title": "Expert Insight",
            "duration_minutes": 8,
            "talking_points": [
                "Guest perspective or expert view",
                "Real-world experience",
                "Advice or recommendations",
                "Lessons learned"
            ]
        },
        {
            "title": "Listener Engagement",
            "duration_minutes": 5,
            "talking_points": [
                "Answer listener questions",
                "Discuss common challenges",
                "Interactive polling or feedback",
                "Community highlights"
            ]
        }
    ],
    "outro": {
        "title": "Conclusion",
        "duration_minutes": 5,
        "talking_points": [
            "Recap key takeaways",
            "Final thoughts or call to action",
            "Preview of next episode",
            "Social media handles and contact info"
        ]
    }
}


def handler(event: dict) -> dict:
    """Generate a podcast outline."""
    try:
        topic = event.get("topic")
        episode_number = event.get("episode_number")
        duration_minutes = event.get("duration_minutes", 30)
        guest_name = event.get("guest_name")

        if not topic:
            return {"ok": False, "error": "topic is required"}
        if episode_number is None:
            return {"ok": False, "error": "episode_number is required"}

        if not isinstance(episode_number, int) or episode_number < 1:
            return {"ok": False, "error": "episode_number must be a positive integer"}
        if not isinstance(duration_minutes, int) or duration_minutes < 5 or duration_minutes > 180:
            return {"ok": False, "error": "duration_minutes must be an integer between 5 and 180"}

        topic_clean = topic.strip()
        episode_title = f"Episode {episode_number}: {topic_clean}"

        intro = SEGMENT_TEMPLATES["intro"].copy()
        intro["talking_points"] = [
            f"Welcome to another episode of [Podcast Name], I'm [Host Name]!",
            f"Today we're diving into: {topic_clean}",
            "If you're new here, subscribe and hit the notification bell",
            f"Here's what we're covering today:"
        ]

        main_segments = []
        remaining_time = duration_minutes - intro["duration_minutes"] - SEGMENT_TEMPLATES["outro"]["duration_minutes"]

        if guest_name:
            guest_segment = {
                "title": f"Interview with {guest_name}",
                "duration_minutes": min(15, remaining_time // 2),
                "talking_points": [
                    f"Introduction of {guest_name} and their background",
                    "Key insights from guest's experience",
                    "Discussion of challenges and solutions",
                    "Advice for listeners"
                ]
            }
            main_segments.append(guest_segment)
            remaining_time -= guest_segment["duration_minutes"]

        num_main_segments = min(len(SEGMENT_TEMPLATES["main"]), max(2, remaining_time // 7))
        time_per_segment = remaining_time // num_main_segments

        for i in range(num_main_segments):
            segment = SEGMENT_TEMPLATES["main"][i % len(SEGMENT_TEMPLATES["main"])].copy()
            segment["duration_minutes"] = time_per_segment
            if segment["title"] == "Expert Insight" and not guest_name:
                segment["title"] = "Deep Dive Analysis"
            main_segments.append(segment)

        outro = SEGMENT_TEMPLATES["outro"].copy()
        outro["talking_points"] = [
            f"Recap our 3 key takeaways from today's episode:",
            f"1. [Key insight about {topic_clean}]",
            f"2. [Actionable tip listeners can apply]",
            f"3. [Resources or next steps]",
            f"Subscribe and leave a review if you enjoyed this episode",
            "Connect with us on Twitter @PodcastHandle",
            f"Next episode: [Preview of upcoming topic]",
            "Thanks for listening!"
        ]

        segments = [intro] + main_segments + [outro]

        intro_script = f"""[INTRO MUSIC FADES]

Host: "Hey everyone, welcome back to another episode of [Podcast Name]. I'm [Your Name], and today we've got an amazing episode lined up for you.

We're talking about {topic_clean} - this is something I've been really passionate about lately, and I think you're going to get a lot of value from it.

If you're new to the show, make sure to hit that subscribe button and stick around until the end because we've got some great stuff coming up.

Let's dive right in!"

[TRANSITION]"""

        outro_script = f"""[OUTRO MUSIC BEGINS]

Host: "Alright, that's a wrap for today! Before we go, let me leave you with three key takeaways from our conversation about {topic_clean}:

First, [KEY INSIGHT 1]...

Second, [KEY INSIGHT 2]...

And third, [KEY INSIGHT 3]...

If you enjoyed today's episode, it would mean the world to me if you could subscribe, leave a review, and share it with someone who might find it valuable.

You can connect with me on Twitter @YourHandle - I'd love to hear your thoughts on today's topic.

Next week, we're going to be discussing [TOPIC PREVIEW], so make sure you don't miss that.

Thanks for listening, and I'll catch you in the next episode!"

[OUTRO MUSIC FADES]"""

        sponsor_message_template = """[SPONSOR MESSAGE]

Host: "Today's episode is brought to you by [Sponsor Name]. [Sponsor Brief Description]. Visit [URL] to get [Offer/Details] and use code [CODE] for [Discount/Offer].

[BACK TO SHOW]"""

        total_duration = sum(seg["duration_minutes"] for seg in segments)

        return {
            "ok": True,
            "topic": topic_clean,
            "episode_number": episode_number,
            "episode_title": episode_title,
            "duration_minutes": total_duration,
            "intro_script": intro_script.strip(),
            "segments": segments,
            "outro_script": outro_script.strip(),
            "sponsor_message_template": sponsor_message_template.strip(),
            "guest_name": guest_name,
            "generated_at": datetime.now().isoformat()
        }

    except Exception as e:
        return {"ok": False, "error": f"Failed to generate podcast outline: {str(e)}"}


if __name__ == "__main__":
    result = handler({
        "topic": "The Future of AI in Healthcare",
        "episode_number": 42,
        "duration_minutes": 30,
        "guest_name": "Dr. Jane Smith"
    })
    print(result)
