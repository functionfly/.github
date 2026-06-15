"""
Performance Review Generator - Generate structured performance review documents.
"""


RATING_SCALES = {
    1: "Needs Significant Improvement",
    2: "Needs Improvement",
    3: "Meets Some Expectations",
    4: "Meets Expectations",
    5: "Exceeds Expectations",
}


def handler(event):
    if isinstance(event, dict):
        employee_name = event.get("employee_name", "")
        job_title = event.get("job_title", "")
        review_period = event.get("review_period", "")
        achievements = event.get("achievements", [])
        areas_for_improvement = event.get("areas_for_improvement", [])
        rating = event.get("rating")
        goals = event.get("goals", [])
        manager_name = event.get("manager_name", "")
    else:
        employee_name, job_title, review_period, achievements, areas_for_improvement, rating, goals, manager_name = "", "", "", [], [], None, [], ""

    if not employee_name:
        return {"ok": False, "error": "employee_name is required"}
    if not job_title:
        return {"ok": False, "error": "job_title is required"}
    if not review_period:
        return {"ok": False, "error": "review_period is required"}

    if rating is None:
        return {"ok": False, "error": "rating is required"}

    try:
        rating = int(rating)
        if rating < 1 or rating > 5:
            return {"ok": False, "error": "rating must be between 1 and 5"}
    except (ValueError, TypeError):
        return {"ok": False, "error": "rating must be an integer (1-5)"}

    if not isinstance(achievements, list):
        return {"ok": False, "error": "achievements must be a list"}
    if not isinstance(areas_for_improvement, list):
        return {"ok": False, "error": "areas_for_improvement must be a list"}

    try:
        rating_label = RATING_SCALES.get(rating, "Unknown")

        # Build review text
        review_lines = [
            f"PERFORMANCE REVIEW",
            f"=" * 50,
            f"",
            f"Employee: {employee_name}",
            f"Position: {job_title}",
            f"Review Period: {review_period}",
            f"Overall Rating: {rating} - {rating_label}",
            f"",
            f"-" * 50,
            f"",
        ]

        # Achievements section
        if achievements:
            review_lines.append("KEY ACHIEVEMENTS")
            review_lines.append("-" * 25)
            for i, achievement in enumerate(achievements, 1):
                review_lines.append(f"  {i}. {achievement}")
            review_lines.append("")

        # Areas for improvement section
        if areas_for_improvement:
            review_lines.append("AREAS FOR DEVELOPMENT")
            review_lines.append("-" * 25)
            for i, area in enumerate(areas_for_improvement, 1):
                review_lines.append(f"  {i}. {area}")
            review_lines.append("")

        # Goals section
        if goals:
            review_lines.append("GOALS FOR NEXT REVIEW PERIOD")
            review_lines.append("-" * 25)
            for i, goal in enumerate(goals, 1):
                review_lines.append(f"  {i}. {goal}")
            review_lines.append("")

        # Overall assessment
        review_lines.append("OVERALL ASSESSMENT")
        review_lines.append("-" * 25)

        if rating >= 4:
            assessment = (
                f"{employee_name} has demonstrated strong performance in their role as {job_title} "
                f"during the {review_period} review period. Their contributions have positively impacted "
                f"the team and organization. Continued growth in the identified areas will position them "
                f"for additional responsibilities and advancement opportunities."
            )
        elif rating == 3:
            assessment = (
                f"{employee_name} has met the core expectations of their role as {job_title} "
                f"during the {review_period} review period. There are opportunities for growth in "
                f"the areas identified above. With focused development, we expect to see increased "
                f"contributions in the next review period."
            )
        else:
            assessment = (
                f"{employee_name} has faced challenges in meeting expectations for their role as {job_title} "
                f"during the {review_period} review period. The areas for improvement identified above require "
                f"immediate attention and focused effort. A performance improvement plan may be appropriate. "
                f"We are committed to supporting {employee_name} in meeting expectations."
            )

        review_lines.append(assessment)
        review_lines.append("")

        # Development goals
        development_goals = []
        for area in areas_for_improvement[:3]:
            development_goals.append({
                "goal": f"Improve in {area}",
                "actions": [
                    "Work with manager to create specific improvement plan",
                    "Seek feedback regularly",
                    "Complete relevant training if available",
                ],
                "timeline": "90 days",
            })

        # Signature lines
        review_lines.append("-" * 50)
        review_lines.append(f"Employee Signature: _____________________ Date: _________")
        review_lines.append(f"Manager Signature: {manager_name or '_________________'} Date: _________")
        review_lines.append(f"HR Signature: _____________________ Date: _________")

        review_text = "\n".join(review_lines)

        return {
            "ok": True,
            "employee_name": employee_name,
            "job_title": job_title,
            "review_period": review_period,
            "overall_rating": rating,
            "rating_label": rating_label,
            "review_text": review_text,
            "strengths": achievements if achievements else [],
            "development_goals": development_goals,
            "requires_improvement_plan": rating <= 2,
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}