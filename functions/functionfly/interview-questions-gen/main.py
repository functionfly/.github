"""Interview Questions Generator - Generate interview questions."""
from datetime import datetime
from typing import Any


EXPERIENCE_LEVELS = {
    "entry": {
        "question_types": ["basic", "situational", "motivational"],
        "difficulty": "Foundational"
    },
    "mid": {
        "question_types": ["behavioral", "situational", "technical", "motivational"],
        "difficulty": "Intermediate"
    },
    "senior": {
        "question_types": ["behavioral", "leadership", "technical", "strategic", "cultural"],
        "difficulty": "Advanced"
    },
    "executive": {
        "question_types": ["vision", "leadership", "strategic", "cultural", "industry"],
        "difficulty": "Executive"
    }
}


QUESTION_TEMPLATES = {
    "basic": [
        {"question": "What interests you about this position?", "sample_answer": "I'm drawn to this role because it aligns with my career goals and allows me to utilize my existing skills while continuing to grow."},
        {"question": "Where do you see yourself in 5 years?", "sample_answer": "I see myself growing into a leadership role where I can mentor others and contribute to strategic decision-making."},
        {"question": "Why should we hire you?", "sample_answer": "I bring a unique combination of skills, experience, and enthusiasm that will add value to your team."},
    ],
    "behavioral": [
        {"question": "Tell me about a time when you had to work under pressure.", "sample_answer": "Describe a specific situation with the STAR method: Situation, Task, Action, Result."},
        {"question": "Describe a conflict you had with a coworker and how you resolved it.", "sample_answer": "Focus on the resolution process and what you learned from the experience."},
        {"question": "Give an example of a goal you reached and how you achieved it.", "sample_answer": "Be specific about the steps you took and the measurable outcomes."},
    ],
    "situational": [
        {"question": "How would you handle a difficult customer?", "sample_answer": "I would listen actively, empathize with their concerns, and find a solution that addresses their needs."},
        {"question": "What would you do if you missed a deadline?", "sample_answer": "I would communicate immediately, assess the situation, and create a plan to get back on track."},
        {"question": "How would you prioritize multiple urgent tasks?", "sample_answer": "I would assess urgency and importance, communicate with stakeholders, and tackle items systematically."},
    ],
    "technical": [
        {"question": "What technical skills qualify you for this role?", "sample_answer": "List specific technical competencies, tools, and methodologies you are proficient in."},
        {"question": "How do you stay current with industry developments?", "sample_answer": "Mention specific resources, communities, and continuous learning practices."},
        {"question": "Describe a technical challenge you faced and how you solved it.", "sample_answer": "Walk through the problem-solving process and the outcome."},
    ],
    "motivational": [
        {"question": "What motivates you to do your best work?", "sample_answer": "Discuss intrinsic and extrinsic motivators that drive your performance."},
        {"question": "What feedback have you received most frequently?", "sample_answer": "Be honest and focus on constructive feedback you've worked to address."},
        {"question": "How do you handle constructive criticism?", "sample_answer": "I view criticism as an opportunity for growth and actively work to implement feedback."},
    ],
    "leadership": [
        {"question": "How do you motivate team members who are underperforming?", "sample_answer": "I believe in understanding root causes, providing support, and setting clear expectations."},
        {"question": "Describe your management style.", "sample_answer": "Discuss your approach to delegation, feedback, and team development."},
        {"question": "How do you develop talent on your team?", "sample_answer": "I identify individual growth areas and create opportunities for skill development."},
    ],
    "strategic": [
        {"question": "How do you approach decision-making?", "sample_answer": "I gather data, consider stakeholder input, evaluate options, and make decisive choices."},
        {"question": "What would be your first 90-day plan?", "sample_answer": "Focus on listening, learning, building relationships, and delivering quick wins."},
        {"question": "How do you balance short-term needs with long-term vision?", "sample_answer": "I ensure quick wins support the broader strategy while maintaining focus on long-term goals."},
    ],
    "cultural": [
        {"question": "What type of work environment brings out your best work?", "sample_answer": "Be honest about your preferences while showing adaptability."},
        {"question": "How do you contribute to team culture?", "sample_answer": "Discuss specific ways you positively impact team dynamics."},
        {"question": "Describe your ideal manager.", "sample_answer": "Focus on collaboration, feedback, and support qualities."},
    ],
    "vision": [
        {"question": "Where do you see the industry heading in the next decade?", "sample_answer": "Demonstrate industry awareness and strategic thinking."},
        {"question": "What is your vision for this company?", "sample_answer": "Show you've researched the company and align your vision with theirs."},
        {"question": "How would you position this company against competitors?", "sample_answer": "Demonstrate competitive analysis and strategic positioning."},
    ],
    "industry": [
        {"question": "What trends are shaping our industry?", "sample_answer": "Showcase knowledge of market dynamics and emerging trends."},
        {"question": "What challenges does our company face?", "sample_answer": "Demonstrate understanding of specific company challenges."},
    ]
}


CASE_STUDIES = [
    {
        "title": "The Budget Challenge",
        "scenario": "Your team has a limited budget but must deliver a project with significant impact. How do you approach this?",
        "key_points": ["Resource optimization", "Prioritization", "Creative solutions", "Stakeholder management"]
    },
    {
        "title": "The Personnel Issue",
        "scenario": "A key team member is underperforming, and the project deadline is approaching. What do you do?",
        "key_points": ["Performance management", "Risk assessment", "Contingency planning", "Team communication"]
    },
    {
        "title": "The Strategic Pivot",
        "scenario": "Mid-project, leadership changes direction significantly. How do you adapt?",
        "key_points": ["Flexibility", "Communication", "Quick assessment", "Implementation planning"]
    }
]


def handler(event: dict) -> dict:
    """Generate interview questions."""
    try:
        job_title = event.get("job_title")
        experience_level = event.get("experience_level", "mid")
        num_questions = event.get("num_questions", 10)
        include_case_study = event.get("include_case_study", False)

        if not job_title:
            return {"ok": False, "error": "job_title is required"}
        if experience_level not in ["entry", "mid", "senior", "executive"]:
            return {"ok": False, "error": "experience_level must be one of: entry, mid, senior, executive"}
        if not isinstance(num_questions, int) or num_questions < 1 or num_questions > 30:
            return {"ok": False, "error": "num_questions must be an integer between 1 and 30"}

        level_config = EXPERIENCE_LEVELS[experience_level]
        question_types = level_config["question_types"]

        questions = []
        used_templates = set()

        for q_type in question_types:
            templates = QUESTION_TEMPLATES.get(q_type, [])
            for template in templates:
                template_key = f"{q_type}:{template['question']}"
                if template_key not in used_templates:
                    questions.append({
                        "question": template["question"],
                        "type": q_type,
                        "difficulty": level_config["difficulty"],
                        "sample_answer": template.get("sample_answer", "Develop your own response based on your experience.")
                    })
                    used_templates.add(template_key)
                    if len(questions) >= num_questions:
                        break
            if len(questions) >= num_questions:
                break

        while len(questions) < num_questions:
            idx = len(questions) % len(QUESTION_TEMPLATES["behavioral"])
            template = QUESTION_TEMPLATES["behavioral"][idx]
            questions.append({
                "question": template["question"],
                "type": "behavioral",
                "difficulty": level_config["difficulty"],
                "sample_answer": template.get("sample_answer", "Use the STAR method to structure your response.")
            })

        interview_tips = [
            "Research the company thoroughly before the interview.",
            "Prepare specific examples from your experience for each question.",
            "Use the STAR method (Situation, Task, Action, Result) for behavioral questions.",
            "Prepare thoughtful questions to ask the interviewer.",
            "Send a thank-you note within 24 hours of the interview.",
            "Practice your responses but avoid sounding rehearsed.",
            "Be honest about your weaknesses and how you're working to improve them."
        ]

        result = {
            "ok": True,
            "job_title": job_title,
            "experience_level": experience_level,
            "num_questions": len(questions),
            "questions": questions,
            "interview_tips": interview_tips,
            "generated_at": datetime.now().isoformat()
        }

        if include_case_study:
            import random
            result["case_study"] = random.choice(CASE_STUDIES)

        return result

    except Exception as e:
        return {"ok": False, "error": f"Failed to generate interview questions: {str(e)}"}


if __name__ == "__main__":
    result = handler({
        "job_title": "Product Manager",
        "experience_level": "senior",
        "num_questions": 10,
        "include_case_study": True
    })
    print(result)
