"""Resume Builder - Generate professional resumes."""
from datetime import datetime
from typing import Any


def format_experience(experience: list) -> str:
    """Format work experience section."""
    if not experience:
        return "No work experience listed."

    sections = []
    for exp in experience:
        company = exp.get("company", "Company")
        title = exp.get("title", "Position")
        dates = exp.get("dates", "N/A")
        bullets = exp.get("bullets", [])

        section = f"""
{title}
{company} | {dates}
{"-" * 50}"""

        if bullets:
            for bullet in bullets:
                section += f"\n• {bullet}"
        else:
            section += "\n• Responsibilities and achievements as documented"

        sections.append(section)

    return "\n".join(sections)


def format_education(education: list) -> str:
    """Format education section."""
    if not education:
        return "No education listed."

    sections = []
    for edu in education:
        degree = edu.get("degree", "Degree")
        school = edu.get("school", "School")
        year = edu.get("year", "")
        gpa = edu.get("gpa", "")

        section = f"{degree}"
        if gpa:
            section += f" | GPA: {gpa}"
        section += f"\n{school}"
        if year:
            section += f" | Class of {year}"

        sections.append(section)

    return "\n\n".join(sections)


def format_skills(skills: list) -> str:
    """Format skills section."""
    if not skills:
        return "No skills listed."

    return " • ".join(skills)


def handler(event: dict) -> dict:
    """Build a resume."""
    try:
        name = event.get("name")
        email = event.get("email")
        phone = event.get("phone")
        experience = event.get("experience", [])
        education = event.get("education", [])
        skills = event.get("skills", [])

        if not name:
            return {"ok": False, "error": "name is required"}
        if not email:
            return {"ok": False, "error": "email is required"}
        if not phone:
            return {"ok": False, "error": "phone is required"}

        if not isinstance(experience, list):
            return {"ok": False, "error": "experience must be a list"}
        if not isinstance(education, list):
            return {"ok": False, "error": "education must be a list"}
        if not isinstance(skills, list):
            return {"ok": False, "error": "skills must be a list"}

        today = datetime.now()

        resume_text = f"""
{'='*70}
{CENTER_TEXT}
{'='*70}

{CENTER_TEXT}
{email} | {phone} | Location
{'='*70}

PROFESSIONAL SUMMARY
{"-" * 50}
Dynamic professional with {[str(len(experience)) + '+' if experience else '0'}] years of experience. Proven track record of delivering results and adding value. Strong communicator with excellent problem-solving abilities.

EXPERIENCE
{"=" * 70}
{format_experience(experience)}

EDUCATION
{"=" * 70}
{format_education(education)}

SKILLS
{"=" * 70}
{format_skills(skills)}

{"=" * 70}
Resume generated: {today.strftime('%B %d, %Y')}
""".replace(f"{CENTER_TEXT}", name)

        resume_html = f"""
<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <style>
        body {{ font-family: 'Helvetica Neue', Arial, sans-serif; font-size: 11pt; line-height: 1.5; color: #333; max-width: 8.5in; margin: 0 auto; padding: 40px; }}
        .header {{ text-align: center; margin-bottom: 30px; border-bottom: 2px solid #333; padding-bottom: 20px; }}
        .name {{ font-size: 24pt; font-weight: bold; margin-bottom: 10px; }}
        .contact {{ font-size: 10pt; color: #666; }}
        h2 {{ font-size: 12pt; text-transform: uppercase; border-bottom: 1px solid #333; padding-bottom: 5px; margin-top: 25px; color: #333; }}
        .section {{ margin-bottom: 20px; }}
        .job-title {{ font-weight: bold; font-size: 11pt; }}
        .company {{ font-style: italic; }}
        .date {{ float: right; }}
        ul {{ margin: 10px 0 10px 20px; }}
        li {{ margin-bottom: 5px; }}
        .skills {{ line-height: 1.8; }}
    </style>
</head>
<body>
    <div class="header">
        <div class="name">{name}</div>
        <div class="contact">{email} | {phone} | Location</div>
    </div>

    <div class="section">
        <h2>Professional Summary</h2>
        <p>Dynamic professional with {len(experience) if experience else 0} years of experience. Proven track record of delivering results and adding value. Strong communicator with excellent problem-solving abilities.</p>
    </div>

    <div class="section">
        <h2>Experience</h2>
"""

        for exp in experience:
            company = exp.get("company", "Company")
            title = exp.get("title", "Position")
            dates = exp.get("dates", "N/A")
            bullets = exp.get("bullets", [])

            resume_html += f"""
        <div class="job-title">{title} <span class="date">{dates}</span></div>
        <div class="company">{company}</div>
        <ul>
"""
            for bullet in bullets:
                resume_html += f"            <li>{bullet}</li>\n"
            resume_html += "        </ul>\n"

        resume_html += """
    </div>

    <div class="section">
        <h2>Education</h2>
"""

        for edu in education:
            degree = edu.get("degree", "Degree")
            school = edu.get("school", "School")
            year = edu.get("year", "")

            resume_html += f"""
        <p><strong>{degree}</strong><br>
        {school}"""
            if year:
                resume_html += f" | Class of {year}"
            resume_html += "</p>\n"

        resume_html += """
    </div>

    <div class="section">
        <h2>Skills</h2>
        <p class="skills">"""

        if skills:
            resume_html += " • ".join(skills)

        resume_html += """
        </p>
    </div>
</body>
</html>"""

        cover_letter_intro = f"""I am writing to express my strong interest in the [Position Name] at [Company Name]. With my background in {[exp.get('title', 'my field') for exp in experience[:1]][0] if experience else 'the industry'}, I am confident in my ability to contribute meaningfully to your team.

"""

        return {
            "ok": True,
            "name": name,
            "email": email,
            "phone": phone,
            "resume_text": resume_text.strip(),
            "resume_html": resume_html.strip(),
            "cover_letter_intro": cover_letter_intro.strip(),
            "experience_count": len(experience),
            "education_count": len(education),
            "skills_count": len(skills),
            "generated_at": datetime.now().isoformat()
        }

    except Exception as e:
        return {"ok": False, "error": f"Failed to build resume: {str(e)}"}


if __name__ == "__main__":
    result = handler({
        "name": "Sarah Johnson",
        "email": "sarah.johnson@email.com",
        "phone": "(555) 123-4567",
        "experience": [
            {
                "company": "TechCorp Inc",
                "title": "Senior Software Engineer",
                "dates": "2020 - Present",
                "bullets": [
                    "Led development of microservices architecture serving 1M+ users",
                    "Mentored team of 5 junior developers",
                    "Reduced deployment time by 60% through CI/CD improvements"
                ]
            },
            {
                "company": "StartupXYZ",
                "title": "Software Engineer",
                "dates": "2017 - 2020",
                "bullets": [
                    "Built core product features using React and Node.js",
                    "Implemented automated testing, achieving 90% code coverage"
                ]
            }
        ],
        "education": [
            {"degree": "B.S. Computer Science", "school": "University of Technology", "year": "2017"}
        ],
        "skills": ["Python", "JavaScript", "React", "Node.js", "AWS", "PostgreSQL", "Docker"]
    })
    print(result)
