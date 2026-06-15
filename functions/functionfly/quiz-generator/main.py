"""Quiz Generator - Generate quizzes with various question types."""
import random
from datetime import datetime
from typing import Any


QUIZ_TEMPLATES = {
    "science": [
        {"question": "What is the chemical symbol for gold?", "options": ["Au", "Ag", "Fe", "Cu"], "correct": 0, "explanation": "Au (Aurum) is the chemical symbol for gold."},
        {"question": "What planet is known as the Red Planet?", "options": ["Venus", "Jupiter", "Mars", "Saturn"], "correct": 2, "explanation": "Mars appears red due to iron oxide on its surface."},
        {"question": "What is the powerhouse of the cell?", "options": ["Nucleus", "Ribosome", "Mitochondria", "Golgi apparatus"], "correct": 2, "explanation": "Mitochondria produce ATP, the cell's energy currency."},
        {"question": "What gas do plants absorb from the atmosphere?", "options": ["Oxygen", "Nitrogen", "Carbon Dioxide", "Hydrogen"], "correct": 2, "explanation": "Plants use CO2 in photosynthesis to produce glucose."},
        {"question": "What is the speed of light in vacuum?", "options": ["300,000 km/s", "150,000 km/s", "500,000 km/s", "1,000,000 km/s"], "correct": 0, "explanation": "Light travels at approximately 299,792 km/s in vacuum."},
    ],
    "history": [
        {"question": "In what year did World War II end?", "options": ["1943", "1944", "1945", "1946"], "correct": 2, "explanation": "WWII ended in 1945 with the surrender of Japan."},
        {"question": "Who was the first President of the United States?", "options": ["Thomas Jefferson", "John Adams", "George Washington", "Benjamin Franklin"], "correct": 2, "explanation": "George Washington served as president from 1789-1797."},
        {"question": "What ancient wonder was located in Alexandria?", "options": ["Colossus", "Lighthouse", "Hanging Gardens", "Temple of Artemis"], "correct": 1, "explanation": "The Lighthouse of Alexandria was one of the Seven Wonders."},
        {"question": "What year did the Berlin Wall fall?", "options": ["1987", "1988", "1989", "1990"], "correct": 2, "explanation": "The Berlin Wall fell on November 9, 1989."},
        {"question": "Who wrote the Declaration of Independence?", "options": ["George Washington", "Benjamin Franklin", "Thomas Jefferson", "John Adams"], "correct": 2, "explanation": "Thomas Jefferson was the primary author."},
    ],
    "technology": [
        {"question": "What does CPU stand for?", "options": ["Central Processing Unit", "Computer Personal Unit", "Central Program Utility", "Core Processing Unit"], "correct": 0, "explanation": "CPU stands for Central Processing Unit."},
        {"question": "What programming language is known as the 'language of the web'?", "options": ["Python", "Java", "JavaScript", "C++"], "correct": 2, "explanation": "JavaScript is the primary language for web development."},
        {"question": "What does HTML stand for?", "options": ["Hyper Text Markup Language", "High Tech Modern Language", "Hyper Transfer Markup Language", "Home Tool Markup Language"], "correct": 0, "explanation": "HTML stands for Hyper Text Markup Language."},
        {"question": "Who founded Microsoft?", "options": ["Steve Jobs", "Bill Gates", "Mark Zuckerberg", "Elon Musk"], "correct": 1, "explanation": "Bill Gates co-founded Microsoft in 1975."},
        {"question": "What does RAM stand for?", "options": ["Random Access Memory", "Read Access Memory", "Rapid Access Module", "Runtime Application Memory"], "correct": 0, "explanation": "RAM stands for Random Access Memory."},
    ],
    "general": [
        {"question": "What is the largest ocean on Earth?", "options": ["Atlantic", "Indian", "Arctic", "Pacific"], "correct": 3, "explanation": "The Pacific Ocean is the largest and deepest."},
        {"question": "How many continents are there?", "options": ["5", "6", "7", "8"], "correct": 2, "explanation": "There are 7 continents: Asia, Africa, North America, South America, Antarctica, Europe, and Australia."},
        {"question": "What is the capital of Japan?", "options": ["Seoul", "Beijing", "Tokyo", "Bangkok"], "correct": 2, "explanation": "Tokyo has been Japan's capital since 1868."},
        {"question": "What is the currency of the United Kingdom?", "options": ["Euro", "Dollar", "Pound Sterling", "Franc"], "correct": 2, "explanation": "The Pound Sterling (GBP) is the UK's currency."},
        {"question": "How many sides does a hexagon have?", "options": ["5", "6", "7", "8"], "correct": 1, "explanation": "A hexagon has 6 sides."},
    ]
}


def get_category_for_topic(topic: str) -> str:
    """Determine the category based on topic keywords."""
    topic_lower = topic.lower()
    if any(kw in topic_lower for kw in ["science", "physics", "chemistry", "biology", "space"]):
        return "science"
    if any(kw in topic_lower for kw in ["history", "war", "ancient", "civilization"]):
        return "history"
    if any(kw in topic_lower for kw in ["tech", "computer", "software", "programming", "code"]):
        return "technology"
    return "general"


def generate_true_false_questions(topic: str, num: int) -> list:
    """Generate true/false questions."""
    questions = []
    statements = [
        f"The concept of {topic} is fundamental to modern understanding.",
        f"{topic} has remained unchanged over the past century.",
        f"{topic} plays a crucial role in contemporary practice.",
        f"Understanding {topic} is essential for professionals in the field.",
        f"{topic} continues to evolve with new discoveries.",
    ]

    for i in range(min(num, len(statements))):
        is_true = i % 2 == 0
        questions.append({
            "question": statements[i],
            "options": ["True", "False"],
            "correct": 0 if is_true else 1,
            "explanation": f"This statement is {'true' if is_true else 'false'} based on current understanding of {topic}."
        })

    return questions


def handler(event: dict) -> dict:
    """Generate a quiz."""
    try:
        topic = event.get("topic")
        num_questions = event.get("num_questions", 10)
        difficulty = event.get("difficulty", "medium")
        quiz_type = event.get("quiz_type", "mixed")

        if not topic:
            return {"ok": False, "error": "topic is required"}
        if not isinstance(num_questions, int) or num_questions < 1 or num_questions > 50:
            return {"ok": False, "error": "num_questions must be an integer between 1 and 50"}
        if difficulty not in ["easy", "medium", "hard"]:
            return {"ok": False, "error": "difficulty must be one of: easy, medium, hard"}
        if quiz_type not in ["multiple_choice", "true_false", "mixed"]:
            return {"ok": False, "error": "quiz_type must be one of: multiple_choice, true_false, mixed"}

        category = get_category_for_topic(topic)
        template_questions = QUIZ_TEMPLATES.get(category, QUIZ_TEMPLATES["general"])

        questions = []

        if quiz_type in ["multiple_choice", "mixed"]:
            selected = template_questions[:num_questions]
            for q in selected:
                questions.append({
                    "question": q["question"],
                    "options": q["options"],
                    "correct_answer": q["options"][q["correct"]],
                    "correct_index": q["correct"],
                    "explanation": q["explanation"],
                    "type": "multiple_choice"
                })

        if quiz_type == "true_false" or (quiz_type == "mixed" and len(questions) < num_questions):
            tf_questions = generate_true_false_questions(topic, num_questions)
            for q in tf_questions:
                if len(questions) < num_questions:
                    q["correct_answer"] = q["options"][q["correct"]]
                    q["correct_index"] = q["correct"]
                    questions.append(q)

        random.seed(datetime.now().microsecond)
        random.shuffle(questions)

        for i, q in enumerate(questions):
            q["question_number"] = i + 1
            if difficulty == "easy":
                q["difficulty"] = "Easy"
            elif difficulty == "medium":
                q["difficulty"] = "Medium"
            else:
                q["difficulty"] = "Hard"

        difficulty_markers = {
            "easy": "Perfect for beginners",
            "medium": "For those with some knowledge",
            "hard": "Challenge yourself!"
        }

        quiz_metadata = {
            "topic": topic,
            "total_questions": len(questions),
            "difficulty": difficulty,
            "quiz_type": quiz_type,
            "time_estimate_minutes": len(questions) * 1.5,
            "passing_score": "70%" if difficulty != "hard" else "80%"
        }

        return {
            "ok": True,
            "topic": topic,
            "questions": questions,
            "quiz_metadata": quiz_metadata,
            "generated_at": datetime.now().isoformat()
        }

    except Exception as e:
        return {"ok": False, "error": f"Failed to generate quiz: {str(e)}"}


if __name__ == "__main__":
    result = handler({
        "topic": "World History",
        "num_questions": 5,
        "difficulty": "medium",
        "quiz_type": "multiple_choice"
    })
    print(result)
