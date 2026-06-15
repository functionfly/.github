"""Ebook Outline Generator - Generate structured ebook outlines."""
import re
from datetime import datetime
from typing import Any


CHAPTER_TEMPLATES = {
    "introduction": {
        "title_template": "Introduction: {topic}",
        "word_count_range": (1500, 3000),
        "key_topics_template": ["What is {topic}", "Why {topic} matters", "Who this book is for", "How to use this book"]
    },
    "chapter": {
        "title_template": "Chapter {num}: {subtopic}",
        "word_count_range": (3000, 5000),
        "key_topics_template": ["Understanding {subtopic}", "Key principles of {subtopic}", "Practical applications", "Common challenges and solutions"]
    },
    "conclusion": {
        "title_template": "Conclusion: Bringing It All Together",
        "word_count_range": (2000, 3500),
        "key_topics_template": ["Key takeaways", "Next steps for readers", "Final thoughts on {topic}", "Resources for continued learning"]
    }
}


def generate_chapter_title(chapter_type: str, topic: str, num: int = None, subtopic: str = None) -> str:
    """Generate chapter title based on template."""
    if chapter_type == "introduction":
        return CHAPTER_TEMPLATES["introduction"]["title_template"].format(topic=topic)
    elif chapter_type == "conclusion":
        return CHAPTER_TEMPLATES["conclusion"]["title_template"]
    else:
        if not subtopic:
            subtopic = f"Advanced {topic}" if num > 3 else f"Fundamentals of {topic}"
        return CHAPTER_TEMPLATES["chapter"]["title_template"].format(num=num, subtopic=subtopic)


def generate_key_topics(chapter_type: str, topic: str, subtopic: str = None) -> list:
    """Generate key topics for a chapter."""
    template = CHAPTER_TEMPLATES[chapter_type]["key_topics_template"]
    return [t.format(topic=topic, subtopic=subtopic or topic) for t in template]


def handler(event: dict) -> dict:
    """Generate an ebook outline."""
    try:
        topic = event.get("topic")
        target_audience = event.get("target_audience")
        num_chapters = event.get("num_chapters", 10)

        if not topic:
            return {"ok": False, "error": "topic is required"}
        if not target_audience:
            return {"ok": False, "error": "target_audience is required"}
        if not isinstance(num_chapters, int) or num_chapters < 3 or num_chapters > 30:
            return {"ok": False, "error": "num_chapters must be an integer between 3 and 30"}

        topic_clean = topic.strip()
        audience_clean = target_audience.strip()

        if len(topic_clean) < 3:
            return {"ok": False, "error": "topic must be at least 3 characters"}

        title = f"The Complete Guide to {topic_clean}"
        subtitle = f"A Comprehensive Resource for {audience_clean}"

        chapters = []

        chapters.append({
            "chapter_number": 1,
            "title": generate_chapter_title("introduction", topic_clean),
            "word_count": 2500,
            "key_topics": generate_key_topics("introduction", topic_clean),
            "type": "introduction"
        })

        main_chapters = num_chapters - 2
        subtopic_areas = [
            "Getting Started", "Core Principles", "Building Foundations",
            "Advanced Strategies", "Best Practices", "Common Pitfalls",
            "Tools and Resources", "Implementation Guide", "Case Studies",
            "Real-World Examples", "Expert Insights", "Deep Dive Analysis",
            "Practical Exercises", "Measuring Success", "Scaling Up",
            "Maintenance Tips", "Troubleshooting", "Future Trends"
        ]

        for i in range(main_chapters):
            chapter_num = i + 2
            subtopic = subtopic_areas[i % len(subtopic_areas)]
            word_range = CHAPTER_TEMPLATES["chapter"]["word_count_range"]
            word_count = (word_range[0] + word_range[1]) // 2

            chapters.append({
                "chapter_number": chapter_num,
                "title": generate_chapter_title("chapter", topic_clean, chapter_num, subtopic),
                "word_count": word_count,
                "key_topics": generate_key_topics("chapter", topic_clean, subtopic),
                "type": "main"
            })

        chapters.append({
            "chapter_number": num_chapters,
            "title": generate_chapter_title("conclusion", topic_clean),
            "word_count": 2800,
            "key_topics": generate_key_topics("conclusion", topic_clean),
            "type": "conclusion"
        })

        total_word_count = sum(ch["word_count"] for ch in chapters)

        estimated_reading_hours = round(total_word_count / 750, 1)

        return {
            "ok": True,
            "title": title,
            "subtitle": subtitle,
            "topic": topic_clean,
            "target_audience": audience_clean,
            "num_chapters": num_chapters,
            "chapters": chapters,
            "total_word_count_estimate": total_word_count,
            "estimated_reading_hours": estimated_reading_hours,
            "generated_at": datetime.now().isoformat()
        }

    except Exception as e:
        return {"ok": False, "error": f"Failed to generate ebook outline: {str(e)}"}


if __name__ == "__main__":
    result = handler({
        "topic": "Digital Marketing Strategy",
        "target_audience": "Small Business Owners",
        "num_chapters": 10
    })
    print(result)
