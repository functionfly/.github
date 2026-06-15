"""Study Schedule Generator - Generate personalized study schedules."""
from datetime import datetime, timedelta
from typing import Any


def handler(event: dict) -> dict:
    """Generate a study schedule."""
    try:
        exam_date = event.get("exam_date")
        topics = event.get("topics", [])
        daily_study_hours = event.get("daily_study_hours", 2)

        if not exam_date:
            return {"ok": False, "error": "exam_date is required"}
        if not topics or len(topics) == 0:
            return {"ok": False, "error": "topics list is required and must not be empty"}
        if not isinstance(daily_study_hours, (int, float)) or daily_study_hours <= 0 or daily_study_hours > 16:
            return {"ok": False, "error": "daily_study_hours must be between 0.5 and 16"}

        for i, topic in enumerate(topics):
            if not isinstance(topic, dict):
                return {"ok": False, "error": f"topic at index {i} must be an object"}

        try:
            exam = datetime.strptime(exam_date, "%Y-%m-%d")
        except ValueError:
            return {"ok": False, "error": "exam_date must be in YYYY-MM-DD format"}

        today = datetime.now()
        if exam <= today:
            return {"ok": False, "error": "exam_date must be in the future"}

        days_until_exam = (exam - today).days

        total_required_hours = sum(
            (topic.get("weight", 1) * 2) / (topic.get("mastery_level", 0.5) + 0.1)
            for topic in topics
        )

        adjusted_daily_hours = min(daily_study_hours, total_required_hours / days_until_exam * 1.2)

        topics_with_study_plan = []
        for topic in topics:
            name = topic.get("name", "Unknown Topic")
            weight = topic.get("weight", 1)
            mastery = topic.get("mastery_level", 0.5)

            priority = (weight * (1 - mastery)) / (mastery + 0.1)

            estimated_hours = (weight * 2) / (mastery + 0.1)

            topics_with_study_plan.append({
                "name": name,
                "weight": weight,
                "mastery_level": mastery,
                "priority": round(priority, 3),
                "estimated_hours": round(estimated_hours, 1)
            })

        topics_with_study_plan.sort(key=lambda x: x["priority"], reverse=True)

        schedule = []
        remaining_hours_per_topic = {t["name"]: t["estimated_hours"] for t in topics_with_study_plan}

        current_date = today
        day_num = 0

        while current_date < exam:
            if current_date.weekday() >= 5:
                current_date += timedelta(days=1)
                continue

            daily_schedule = {
                "date": current_date.strftime("%Y-%m-%d"),
                "day_of_week": current_date.strftime("%A"),
                "day_number": day_num + 1,
                "topics": [],
                "hours": 0
            }

            hours_left_today = adjusted_daily_hours

            for topic in topics_with_study_plan:
                if hours_left_today <= 0:
                    break

                topic_name = topic["name"]
                hours_for_topic = min(remaining_hours_per_topic[topic_name], hours_left_today)

                if hours_for_topic >= 0.25:
                    daily_schedule["topics"].append({
                        "name": topic_name,
                        "hours": round(hours_for_topic, 1),
                        "activities": suggest_activities(topic)
                    })
                    remaining_hours_per_topic[topic_name] -= hours_for_topic
                    hours_left_today -= hours_for_topic
                    daily_schedule["hours"] += hours_for_topic

            schedule.append(daily_schedule)
            day_num += 1
            current_date += timedelta(days=1)

            if sum(remaining_hours_per_topic.values()) <= 0:
                break

        total_study_hours = sum(day["hours"] for day in schedule)
        hours_covered = {t["name"]: t["estimated_hours"] - remaining_hours_per_topic.get(t["name"], 0) for t in topics_with_study_plan}

        recommended_daily_plan = {
            "morning": "Start with the most challenging topic when your mind is fresh",
            "afternoon": "Review and practice problems",
            "evening": "Light review or flashcards",
            "tips": [
                "Take a 5-10 minute break every hour",
                "Stay hydrated and snack healthy",
                "Review previous day's material briefly",
                "Practice with past exams or sample questions"
            ]
        }

        return {
            "ok": True,
            "exam_date": exam_date,
            "days_until_exam": days_until_exam,
            "schedule": schedule,
            "topics": topics_with_study_plan,
            "total_study_hours": round(total_study_hours, 1),
            "hours_covered": {k: round(v, 1) for k, v in hours_covered.items()},
            "recommended_daily_plan": recommended_daily_plan,
            "generated_at": datetime.now().isoformat()
        }

    except Exception as e:
        return {"ok": False, "error": f"Failed to generate study schedule: {str(e)}"}


def suggest_activities(topic: dict) -> list:
    """Suggest study activities based on mastery level."""
    mastery = topic.get("mastery_level", 0.5)

    if mastery < 0.3:
        return [
            "Read introduction material",
            "Watch video tutorials",
            "Take notes on key concepts"
        ]
    elif mastery < 0.6:
        return [
            "Practice problems",
            "Review notes",
            "Discuss with study group"
        ]
    else:
        return [
            "Review flashcards",
            "Take practice quizzes",
            "Teach concepts to someone else"
        ]


if __name__ == "__main__":
    result = handler({
        "exam_date": "2026-07-15",
        "topics": [
            {"name": "Calculus", "weight": 3, "mastery_level": 0.4},
            {"name": "Linear Algebra", "weight": 2, "mastery_level": 0.6},
            {"name": "Statistics", "weight": 2, "mastery_level": 0.3},
        ],
        "daily_study_hours": 3
    })
    print(result)
