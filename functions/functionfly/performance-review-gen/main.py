"""Performance Review Generator - Generate employee performance reviews."""
from datetime import datetime
from typing import Any


RATING_DESCRIPTIONS = {
    1: "Does Not Meet Expectations - Performance is below acceptable standards. Significant improvement is required.",
    2: "Needs Improvement - Performance partially meets expectations. Clear gaps exist that must be addressed.",
    3: "Meets Expectations - Performance is satisfactory and meets job requirements. Solid, reliable work.",
    4: "Exceeds Expectations - Performance consistently exceeds requirements. High quality and quantity of work.",
    5: "Exceptional - Performance is outstanding and exemplary. Makes significant contributions beyond role requirements."
}


def handler(event: dict) -> dict:
    """Generate a performance review."""
    try:
        employee_name = event.get("employee_name")
        job_title = event.get("job_title")
        review_period = event.get("review_period")
        achievements = event.get("achievements", [])
        areas_for_improvement = event.get("areas_for_improvement", [])
        rating = event.get("rating")

        if not employee_name:
            return {"ok": False, "error": "employee_name is required"}
        if not job_title:
            return {"ok": False, "error": "job_title is required"}
        if not review_period:
            return {"ok": False, "error": "review_period is required"}
        if not achievements or len(achievements) == 0:
            return {"ok": False, "error": "achievements list is required and must not be empty"}
        if rating is None:
            return {"ok": False, "error": "rating is required"}

        if not isinstance(rating, int) or rating < 1 or rating > 5:
            return {"ok": False, "error": "rating must be an integer between 1 and 5"}

        today = datetime.now()
        review_id = f"PR-{today.strftime('%Y%m%d')}-{hash(employee_name) % 10000:04d}"

        rating_text = RATING_DESCRIPTIONS.get(rating, RATING_DESCRIPTIONS[3])

        strengths = achievements[:5]
        development_goals = []

        for area in areas_for_improvement[:3]:
            goal = f"Develop and demonstrate improved proficiency in {area}"
            development_goals.append(goal)

        if rating <= 3:
            development_goals.append("Complete required training and certification programs")
            development_goals.append("Schedule regular check-ins with manager to track progress")

        if len(achievements) > 5:
            development_goals.append("Continue building on demonstrated successes in current role")

        review_text = f"""
{'='*70}
                    PERFORMANCE REVIEW
{'='*70}

Review ID: {review_id}
Employee: {employee_name}
Title: {job_title}
Review Period: {review_period}
Date: {today.strftime('%B %d, %Y')}

{'='*70}
                         RATING
{'='*70}

Overall Rating: {rating} out of 5

{rating_text}

{'='*70}
                      KEY ACHIEVEMENTS
{'='*70}

"""

        for i, achievement in enumerate(achievements, 1):
            review_text += f"{i}. {achievement}\n\n"

        review_text += f"""
{'='*70}
                   AREAS FOR IMPROVEMENT
{'='*70}

"""

        for i, area in enumerate(areas_for_improvement, 1):
            review_text += f"{i}. {area}\n\n"

        review_text += f"""
{'='*70}
                    DEVELOPMENT GOALS
{'='*70}

"""

        for i, goal in enumerate(development_goals, 1):
            review_text += f"{i}. {goal}\n\n"

        review_text += f"""
{'='*70}
                       SUMMARY
{'='*70}

{employee_name} has demonstrated {'strong' if rating >= 4 else 'satisfactory'} performance
during the {review_period} review period. {'Notable achievements include ' + ', '.join(achievements[:2]) + '.' if achievements else ''}

{'The employee is encouraged to focus on the identified areas for improvement to reach their full potential.' if rating < 4 else 'The employee is on track for advancement opportunities.'}

{'='*70}
                    ACKNOWLEDGMENT
{'='*70}

Employee Signature: _______________________________  Date: ___________

Manager Signature: _______________________________  Date: ___________

{'='*70}
"""

        return {
            "ok": True,
            "review_id": review_id,
            "employee_name": employee_name,
            "job_title": job_title,
            "review_period": review_period,
            "review_date": today.strftime('%B %d, %Y'),
            "review_text": review_text.strip(),
            "strengths": strengths,
            "development_goals": development_goals,
            "overall_rating": rating,
            "rating_description": rating_text,
            "generated_at": datetime.now().isoformat()
        }

    except Exception as e:
        return {"ok": False, "error": f"Failed to generate performance review: {str(e)}"}


if __name__ == "__main__":
    result = handler({
        "employee_name": "Sarah Johnson",
        "job_title": "Senior Software Engineer",
        "review_period": "2025-2026",
        "achievements": [
            "Led the migration of legacy systems to cloud infrastructure, reducing costs by 40%",
            "Mentored 3 junior developers through code reviews and pair programming sessions",
            "Delivered the new payment processing feature 2 weeks ahead of schedule",
            "Reduced production incidents by 60% through implementation of automated testing"
        ],
        "areas_for_improvement": [
            "Cross-team communication during complex projects",
            "Documentation practices for complex technical solutions"
        ],
        "rating": 4
    })
    print(result)
