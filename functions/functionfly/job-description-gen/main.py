"""Job Description Generator - Generate comprehensive job descriptions."""
from datetime import datetime
from typing import Any


def generate_hiring_markers(job_title: str, responsibilities: list, requirements: list) -> list:
    """Generate ATS-friendly hiring markers."""
    markers = []

    markers.append(f"Position: {job_title}")
    markers.append(f"Role: {job_title}")

    for resp in responsibilities[:3]:
        words = resp.split()[:3]
        markers.append(" ".join(words))

    for req in requirements[:3]:
        words = req.split()[:3]
        markers.append(" ".join(words))

    markers.append(job_title.replace(" ", ""))
    markers.append(job_title.replace(" ", "-"))

    return list(set(markers))[:10]


def generate_interview_questions(job_title: str, num: int = 5) -> list:
    """Generate interview questions for the role."""
    return [
        f"Tell me about your experience as a {job_title}.",
        f"What specific skills make you a strong candidate for this {job_title} role?",
        f"Describe a challenging situation you faced in a similar position and how you resolved it.",
        f"Where do you see your career in 3-5 years, and how does this {job_title} position fit?",
        f"How do you stay current with industry trends and best practices?"
    ]


def handler(event: dict) -> dict:
    """Generate a job description."""
    try:
        job_title = event.get("job_title")
        department = event.get("department")
        responsibilities = event.get("responsibilities", [])
        requirements = event.get("requirements", [])
        benefits = event.get("benefits", [])

        if not job_title:
            return {"ok": False, "error": "job_title is required"}
        if not department:
            return {"ok": False, "error": "department is required"}
        if not responsibilities or len(responsibilities) == 0:
            return {"ok": False, "error": "responsibilities list is required"}
        if not requirements or len(requirements) == 0:
            return {"ok": False, "error": "requirements list is required"}

        if not isinstance(responsibilities, list) or not isinstance(requirements, list):
            return {"ok": False, "error": "responsibilities and requirements must be lists"}

        today = datetime.now()
        date_str = today.strftime("%B %d, %Y")

        job_description_html = f"""
<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <style>
        body {{ font-family: Arial, sans-serif; line-height: 1.8; color: #333; max-width: 800px; margin: 0 auto; padding: 40px; }}
        h1 {{ color: #2C3E50; border-bottom: 3px solid #3498DB; padding-bottom: 10px; }}
        h2 {{ color: #2980B9; margin-top: 30px; }}
        ul {{ margin-left: 20px; }}
        li {{ margin-bottom: 8px; }}
        .meta {{ background: #ECF0F1; padding: 15px; border-radius: 4px; margin-bottom: 20px; }}
        .section {{ margin-bottom: 30px; }}
    </style>
</head>
<body>
    <h1>{job_title}</h1>
    <div class="meta">
        <p><strong>Department:</strong> {department}</p>
        <p><strong>Posted:</strong> {date_str}</p>
        <p><strong>Location:</strong> [Location]</p>
        <p><strong>Employment Type:</strong> [Full-time/Part-time/Contract]</p>
    </div>

    <div class="section">
        <h2>About Us</h2>
        <p>[Company Name] is looking for a talented {job_title} to join our {department} team. This is an exciting opportunity to make a significant impact in a dynamic work environment.</p>
    </div>

    <div class="section">
        <h2>Position Summary</h2>
        <p>The {job_title} will be responsible for [brief summary of role]. This position offers the opportunity to work with a collaborative team and grow professionally.</p>
    </div>

    <div class="section">
        <h2>Key Responsibilities</h2>
        <ul>
"""

        for resp in responsibilities:
            job_description_html += f"            <li>{resp}</li>\n"

        job_description_html += """        </ul>
    </div>

    <div class="section">
        <h2>Requirements</h2>
        <ul>
"""

        for req in requirements:
            job_description_html += f"            <li>{req}</li>\n"

        if benefits:
            job_description_html += """        </ul>
    </div>

    <div class="section">
        <h2>Benefits</h2>
        <ul>
"""
            for benefit in benefits:
                job_description_html += f"            <li>{benefit}</li>\n"

        job_description_html += """        </ul>
    </div>

    <div class="section">
        <h2>How to Apply</h2>
        <p>Interested candidates should submit their resume and cover letter to [email]. Please include the job title in the subject line.</p>
    </div>
</body>
</html>"""

        job_description_text = f"""
================================================================================
                                JOB DESCRIPTION
================================================================================

POSITION: {job_title}
DEPARTMENT: {department}
POSTED: {date_str}
LOCATION: [Location]
EMPLOYMENT TYPE: [Full-time/Part-time/Contract]

--------------------------------------------------------------------------------
ABOUT US
--------------------------------------------------------------------------------

[Company Name] is looking for a talented {job_title} to join our {department} team.
This is an exciting opportunity to make a significant impact in a dynamic work environment.

--------------------------------------------------------------------------------
POSITION SUMMARY
--------------------------------------------------------------------------------

The {job_title} will be responsible for [brief summary of role].

--------------------------------------------------------------------------------
KEY RESPONSIBILITIES
--------------------------------------------------------------------------------

"""

        for i, resp in enumerate(responsibilities, 1):
            job_description_text += f"  {i}. {resp}\n"

        job_description_text += """
--------------------------------------------------------------------------------
REQUIREMENTS
--------------------------------------------------------------------------------

"""

        for i, req in enumerate(requirements, 1):
            job_description_text += f"  {i}. {req}\n"

        if benefits:
            job_description_text += """
--------------------------------------------------------------------------------
BENEFITS
--------------------------------------------------------------------------------

"""
            for i, benefit in enumerate(benefits, 1):
                job_description_text += f"  {i}. {benefit}\n"

        job_description_text += """
--------------------------------------------------------------------------------
HOW TO APPLY
--------------------------------------------------------------------------------

Interested candidates should submit their resume and cover letter to [email].
Please include the job title in the subject line.

================================================================================
"""

        hiring_markers = generate_hiring_markers(job_title, responsibilities, requirements)
        interview_questions = generate_interview_questions(job_title, 5)

        return {
            "ok": True,
            "job_title": job_title,
            "department": department,
            "job_description": job_description_text.strip(),
            "job_description_html": job_description_html.strip(),
            "hiring_markers": hiring_markers,
            "interview_questions": interview_questions,
            "posted_date": date_str,
            "generated_at": datetime.now().isoformat()
        }

    except Exception as e:
        return {"ok": False, "error": f"Failed to generate job description: {str(e)}"}


if __name__ == "__main__":
    result = handler({
        "job_title": "Senior Software Engineer",
        "department": "Engineering",
        "responsibilities": [
            "Design and implement scalable software solutions",
            "Lead technical projects and mentor junior developers",
            "Collaborate with cross-functional teams",
            "Participate in code reviews and ensure code quality"
        ],
        "requirements": [
            "5+ years of software development experience",
            "Proficiency in Python, Java, or Go",
            "Experience with cloud platforms (AWS, GCP)",
            "Strong problem-solving skills"
        ],
        "benefits": [
            "Competitive salary and equity",
            "Health, dental, and vision insurance",
            "401(k) matching",
            "Flexible work arrangements"
        ]
    })
    print(result)
