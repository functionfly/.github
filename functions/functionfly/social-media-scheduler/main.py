"""Social Media Scheduler - Schedule social media posts."""
from datetime import datetime, timedelta
from typing import Any


PLATFORM_CONFIGS = {
    "facebook": {
        "best_times": ["9:00 AM", "1:00 PM", "3:00 PM"],
        "optimal_days": ["Tuesday", "Wednesday", "Thursday"],
        "character_limit": 63200,
        "content_style": "conversational, community-focused"
    },
    "twitter": {
        "best_times": ["8:00 AM", "12:00 PM", "5:00 PM"],
        "optimal_days": ["Tuesday", "Wednesday", "Thursday"],
        "character_limit": 280,
        "content_style": "concise, engaging, uses hashtags"
    },
    "instagram": {
        "best_times": ["11:00 AM", "2:00 PM", "7:00 PM"],
        "optimal_days": ["Tuesday", "Wednesday", "Thursday"],
        "character_limit": 2200,
        "content_style": "visual, uses emojis, hashtag-heavy"
    },
    "linkedin": {
        "best_times": ["8:00 AM", "12:00 PM", "5:00 PM"],
        "optimal_days": ["Tuesday", "Wednesday", "Thursday"],
        "character_limit": 3000,
        "content_style": "professional, thought leadership"
    }
}


def parse_optimal_times(times_str: list) -> list:
    """Parse optimal times from input."""
    valid_times = []
    for t in times_str:
        try:
            datetime.strptime(t.strip(), "%I:%M %p")
            valid_times.append(t.strip())
        except ValueError:
            continue
    return valid_times


def generate_schedule(posts: list, week_start_date: str) -> tuple:
    """Generate social media schedule."""
    try:
        start_date = datetime.strptime(week_start_date, "%Y-%m-%d")
    except ValueError:
        raise ValueError("week_start_date must be in YYYY-MM-DD format")

    schedule = []
    used_slots = {}

    for post in posts:
        platform = post.get("platform", "").lower()
        content = post.get("content", "")

        if not content:
            continue

        if platform not in PLATFORM_CONFIGS:
            continue

        config = PLATFORM_CONFIGS[platform]

        optimal_times = post.get("optimal_times", config["best_times"])

        day_offset = len(schedule) % 5
        day_name = config["optimal_days"][day_offset % len(config["optimal_days"])]

        target_date = start_date + timedelta(days=day_offset)

        if platform not in used_slots:
            used_slots[platform] = []

        time_slot = optimal_times[len(used_slots[platform]) % len(optimal_times)]
        used_slots[platform].append(time_slot)

        schedule.append({
            "date": target_date.strftime("%Y-%m-%d"),
            "time": time_slot,
            "platform": platform,
            "content": content[:200] + "..." if len(content) > 200 else content,
            "full_content": content,
            "day_of_week": day_name,
            "character_count": len(content)
        })

    return schedule


def handler(event: dict) -> dict:
    """Schedule social media posts."""
    try:
        posts = event.get("posts", [])
        week_start_date = event.get("week_start_date")

        if not posts or len(posts) == 0:
            return {"ok": False, "error": "posts list is required and must not be empty"}
        if not week_start_date:
            return {"ok": False, "error": "week_start_date is required"}

        for i, post in enumerate(posts):
            if not isinstance(post, dict):
                return {"ok": False, "error": f"post at index {i} must be an object"}
            if "content" not in post:
                return {"ok": False, "error": f"post at index {i} must have content"}
            if "platform" not in post:
                return {"ok": False, "error": f"post at index {i} must have platform"}

        try:
            parsed_date = datetime.strptime(week_start_date, "%Y-%m-%d")
        except ValueError:
            return {"ok": False, "error": "week_start_date must be in YYYY-MM-DD format"}

        schedule = generate_schedule(posts, week_start_date)

        best_times_by_platform = {}
        for platform, config in PLATFORM_CONFIGS.items():
            best_times_by_platform[platform] = {
                "optimal_times": config["best_times"],
                "best_days": config["optimal_days"],
                "character_limit": config["character_limit"]
            }

        schedule_by_platform = {}
        for item in schedule:
            platform = item["platform"]
            if platform not in schedule_by_platform:
                schedule_by_platform[platform] = []
            schedule_by_platform[platform].append({
                "date": item["date"],
                "time": item["time"],
                "day_of_week": item["day_of_week"],
                "preview": item["content"]
            })

        return {
            "ok": True,
            "week_start_date": week_start_date,
            "schedule": schedule,
            "schedule_by_platform": schedule_by_platform,
            "best_times_by_platform": best_times_by_platform,
            "total_posts": len(schedule),
            "platforms_used": list(set(p["platform"] for p in schedule)),
            "generated_at": datetime.now().isoformat()
        }

    except ValueError as e:
        return {"ok": False, "error": str(e)}
    except Exception as e:
        return {"ok": False, "error": f"Failed to schedule posts: {str(e)}"}


if __name__ == "__main__":
    result = handler({
        "week_start_date": "2026-06-15",
        "posts": [
            {"content": "Excited to announce our new product launch! Check out the link in bio for more details.", "platform": "instagram", "optimal_times": ["11:00 AM"]},
            {"content": "Big news! We've just released our latest feature. Thread below for all the details.", "platform": "twitter", "optimal_times": ["8:00 AM"]},
            {"content": "We're hiring! Join our amazing team. Link to apply in comments.", "platform": "facebook", "optimal_times": ["9:00 AM"]},
            {"content": "Thought leadership piece: The future of remote work and its impact on productivity.", "platform": "linkedin", "optimal_times": ["8:00 AM"]},
        ]
    })
    print(result)
