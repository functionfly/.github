"""Flashcard Generator - Generate study flashcards."""
import random
from datetime import datetime
from typing import Any


DIFFICULTY_CONFIG = {
    "basic": {
        "complexity": "simple definitions and facts",
        "hint_level": "explicit",
        "front_template": "What is {topic}?",
        "back_template": "{definition}"
    },
    "intermediate": {
        "complexity": "concepts with examples",
        "hint_level": "moderate",
        "front_template": "Explain the relationship between {topic} and {related}.",
        "back_template": "{explanation}\n\nExample: {example}"
    },
    "advanced": {
        "complexity": "analysis and application",
        "hint_level": "minimal",
        "front_template": "How would you apply {topic} to solve {scenario}?",
        "back_template": "{analysis}\n\nKey considerations:\n{considerations}"
    }
}


def generate_topic_content(topic: str, difficulty: str) -> list:
    """Generate flashcard content based on topic and difficulty."""
    topic_lower = topic.lower()

    content_db = {
        "basic": {
            "science": [
                {"front": "What is photosynthesis?", "back": "The process by which plants convert sunlight, water, and CO2 into glucose and oxygen.", "hint": "Think about what plants need to survive"},
                {"front": "What is the periodic table?", "back": "A systematic arrangement of chemical elements organized by atomic number.", "hint": "Organized by atomic number"},
            ],
            "history": [
                {"front": "What was the Renaissance?", "back": "A period of cultural and intellectual rebirth in Europe (14th-17th century).", "hint": "Rebirth of classical culture"},
                {"front": "What caused World War I?", "back": "A combination of militarism, alliances, imperialism, and nationalism.", "hint": "M-A-I-N acronym"},
            ],
            "default": [
                {"front": f"What is {topic}?", "back": f"{topic} is an important concept that relates to key principles in its field.", "hint": "Focus on the definition"},
                {"front": f"What are the main characteristics of {topic}?", "back": f"Key characteristics include: 1) Fundamental nature, 2) Practical applications, 3) Theoretical framework.", "hint": "Three main aspects"},
            ]
        },
        "intermediate": {
            "science": [
                {"front": "How do enzymes function in biological systems?", "back": "Enzymes act as catalysts, lowering activation energy and speeding up chemical reactions without being consumed.", "example": "Like a key opening a lock - the enzyme (key) fits specific substrates (locks)"},
            ],
            "history": [
                {"front": "Why did the Industrial Revolution begin in Britain?", "back": "Combination of resources (coal, iron), agricultural revolution providing labor, colonial markets, and technological innovation.", "example": "Watt's steam engine was a game-changer for manufacturing"},
            ],
            "default": [
                {"front": f"How does {topic} relate to modern practice?", "back": f"{topic} provides foundational principles that inform contemporary approaches and methodologies.", "example": f"Application in real-world {topic_lower} scenarios"},
            ]
        },
        "advanced": {
            "default": [
                {"front": f"How would you evaluate the effectiveness of {topic} in complex scenarios?", "back": f"Consider multiple factors: context, implementation approach, stakeholder perspectives, and measurable outcomes.", "considerations": "1) Cost-benefit, 2) Sustainability, 3) Scalability, 4) Ethics"},
            ]
        }
    }

    category = "default"
    for cat in ["science", "history"]:
        if cat in topic_lower:
            category = cat
            break

    return content_db.get(difficulty, content_db["basic"]).get(category, content_db["default"])


def handler(event: dict) -> dict:
    """Generate flashcards for a topic."""
    try:
        topic = event.get("topic")
        num_cards = event.get("num_cards", 10)
        difficulty = event.get("difficulty", "basic")

        if not topic:
            return {"ok": False, "error": "topic is required"}
        if not isinstance(num_cards, int) or num_cards < 1 or num_cards > 50:
            return {"ok": False, "error": "num_cards must be an integer between 1 and 50"}
        if difficulty not in ["basic", "intermediate", "advanced"]:
            return {"ok": False, "error": "difficulty must be one of: basic, intermediate, advanced"}

        topic_clean = topic.strip()
        if len(topic_clean) < 2:
            return {"ok": False, "error": "topic must be at least 2 characters"}

        base_content = generate_topic_content(topic_clean, difficulty)

        cards = []
        for i in range(num_cards):
            if i < len(base_content):
                card_data = base_content[i].copy()
            else:
                idx = i % len(base_content)
                card_data = base_content[idx].copy()
                card_data["front"] = card_data["front"].replace(card_data["front"].split()[1], f"{card_data['front'].split()[1]} #{i+1}")

            cards.append({
                "front": card_data["front"],
                "back": card_data["back"],
                "hint": card_data.get("hint", f"Review the concept of {topic_clean}"),
                "card_number": i + 1
            })

    study_tips = {
        "basic": "Review cards daily. Focus on memorizing key definitions and facts. Use the hint if stuck, then try to recall without it.",
        "intermediate": "Try explaining each concept aloud in your own words. Connect new information to things you already know.",
        "advanced": "Apply concepts to real scenarios. Create your own questions. Teach the material to someone else."
    }

    deck_name = f"{topic_clean} Study Deck"

    return {
        "ok": True,
        "topic": topic_clean,
        "num_cards": num_cards,
        "difficulty": difficulty,
        "cards": cards,
        "deck_name": deck_name,
        "study_tips": study_tips[difficulty],
        "generated_at": datetime.now().isoformat()
    }


if __name__ == "__main__":
    result = handler({
        "topic": "World War II",
        "num_cards": 5,
        "difficulty": "intermediate"
    })
    print(result)
