"""
Interview Questions Generator - Generate interview questions for job positions.
"""


# Question banks by category
QUESTION_TEMPLATES = {
    "behavioral": [
        "Tell me about a time when you had to work under pressure.",
        "Describe a situation where you had to deal with a difficult coworker.",
        "Give an example of a goal you reached and how you achieved it.",
        "Tell me about a time you made a mistake. How did you handle it?",
        "Describe a time when you had to persuade others to see your point of view.",
    ],
    "technical": [
        "What programming languages are you most proficient in?",
        "Describe your experience with version control systems.",
        "How do you approach debugging a complex issue?",
        "Explain a technical concept to someone non-technical.",
        "What is your process for code review?",
    ],
    "situational": [
        "How would you handle a project with an unrealistic deadline?",
        "What would you do if you disagreed with your manager's decision?",
        "How would you prioritize multiple urgent tasks?",
        "Describe how you would handle a scope creep scenario.",
        "What steps would you take if a project was failing?",
    ],
    "cultural": [
        "What type of work environment do you thrive in?",
        "How do you stay motivated during repetitive tasks?",
        "Describe your ideal team dynamic.",
        "How do you handle feedback you disagree with?",
        "What makes you want to work for our company?",
    ],
}


def handler(event):
    if isinstance(event, dict):
        job_title = event.get("job_title", "")
        experience_level = event.get("experience_level", "mid")
        num_questions = event.get("num_questions", 10)
        include_case_study = event.get("include_case_study", False)
        categories = event.get("categories", ["behavioral", "technical", "situational", "cultural"])
    else:
        job_title, experience_level, num_questions, include_case_study, categories = "", "mid", 10, False, ["behavioral", "technical", "situational", "cultural"]

    if not job_title:
        return {"ok": False, "error": "job_title is required"}

    try:
        num_questions = int(num_questions)
        if num_questions < 1 or num_questions > 50:
            return {"ok": False, "error": "num_questions must be between 1 and 50"}
    except (ValueError, TypeError):
        return {"ok": False, "error": "num_questions must be an integer"}

    experience_level = experience_level.lower()
    if experience_level not in ("entry", "mid", "senior", "executive"):
        return {"ok": False, "error": "experience_level must be entry/mid/senior/executive"}

    # Level modifiers
    level_modifiers = {
        "entry": "for an entry-level position",
        "mid": "for a mid-level position",
        "senior": "for a senior-level position",
        "executive": "for an executive-level position",
    }

    questions = []
    categories = [c.lower() for c in categories if c.lower() in QUESTION_TEMPLATES]

    if not categories:
        categories = list(QUESTION_TEMPLATES.keys())

    # Distribute questions across categories
    per_category = max(1, num_questions // len(categories))
    for category in categories:
        template_list = QUESTION_TEMPLATES[category]
        # Adjust questions based on experience level
        for i, q in enumerate(template_list[:per_category]):
            if experience_level in ("senior", "executive"):
                q = q.replace("Tell me about", "Share an example where you").replace("Describe", "Analyze a time when you")
            elif experience_level == "entry":
                q = q.replace("Describe a situation", "Walk me through a time")
            questions.append({
                "question": q,
                "type": category,
                "difficulty": "intermediate",
                "sample_answer": "Use the STAR method (Situation, Task, Action, Result)" if category == "behavioral" else None,
            })

    # Add case study if requested
    case_study = None
    if include_case_study:
        case_study = {
            "scenario": f"You are hired for a {job_title} {level_modifiers.get(experience_level, '')}. In your first week, you discover that the team has been using outdated processes that are causing delays. How would you approach this situation?",
            "evaluation_criteria": [
                "Problem identification",
                "Stakeholder communication",
                "Solution design",
                "Implementation planning",
                "Risk assessment",
            ],
            "recommended_format": "Present your approach in 5-7 minutes, covering: 1) How you would assess the situation, 2) Who you would involve, 3) What changes you would propose, 4) How you would implement them.",
        }

    return {
        "ok": True,
        "job_title": job_title,
        "experience_level": experience_level,
        "questions": questions[:num_questions],
        "case_study": case_study,
        "total_questions": min(num_questions, len(questions)),
        "interview_tips": [
            "Research the company beforehand",
            "Prepare concrete examples from your experience",
            "Use the STAR method for behavioral questions",
            "Ask clarifying questions if needed",
            "Follow up with thoughtful questions about the role",
        ],
    }