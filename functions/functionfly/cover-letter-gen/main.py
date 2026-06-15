"""Cover Letter Generator - Generate professional cover letters."""
import re
from datetime import datetime
from typing import Any


def handler(event: dict) -> dict:
    """Generate a professional cover letter."""
    try:
        applicant_name = event.get("applicant_name")
        job_title = event.get("job_title")
        company_name = event.get("company_name")
        key_qualifications = event.get("key_qualifications", [])
        years_experience = event.get("years_experience")

        if not applicant_name:
            return {"ok": False, "error": "applicant_name is required"}
        if not job_title:
            return {"ok": False, "error": "job_title is required"}
        if not company_name:
            return {"ok": False, "error": "company_name is required"}

        if not isinstance(applicant_name, str) or len(applicant_name) < 2:
            return {"ok": False, "error": "applicant_name must be a valid name"}

        if years_experience is not None:
            if not isinstance(years_experience, (int, float)) or years_experience < 0:
                return {"ok": False, "error": "years_experience must be a non-negative number"}

        today = datetime.now()
        date_str = today.strftime("%B %d, %Y")

        qualifications_text = ""
        if key_qualifications and len(key_qualifications) > 0:
            if len(key_qualifications) == 1:
                qualifications_text = f"I bring {years_experience or 'substantial'} years of experience in {key_qualifications[0]}."
            elif len(key_qualifications) == 2:
                qualifications_text = f"I bring {years_experience or 'substantial'} years of experience in both {key_qualifications[0]} and {key_qualifications[1]}."
            else:
                quals = ", ".join(key_qualifications[:-1])
                qualifications_text = f"I bring {years_experience or 'substantial'} years of experience in {quals}, and {key_qualifications[-1]}."
        else:
            qualifications_text = f"I bring {years_experience or 'several'} years of relevant experience that aligns well with this role."

        opening = f"Dear Hiring Manager at {company_name},"

        if years_experience:
            if years_experience < 1:
                exp_phrase = "as a recent graduate"
            elif years_experience < 3:
                exp_phrase = f"with {int(years_experience)} years of experience"
            elif years_experience < 5:
                exp_phrase = f"with {int(years_experience)} years of hands-on experience"
            else:
                exp_phrase = f"with over {int(years_experience)} years of proven expertise"
        else:
            exp_phrase = "with relevant industry experience"

        body = f"""I am writing to express my strong interest in the {job_title} position at {company_name}. {exp_phrase}, I am confident in my ability to contribute meaningfully to your team and organization.

{qualifications_text}

I am particularly drawn to {company_name} because of your reputation for excellence and innovation in the industry. I would welcome the opportunity to bring my skills and enthusiasm to your team.

I am excited about the possibility of discussing how my background, skills, and interests align with {company_name}'s goals. I am available at your convenience for an interview and look forward to learning more about this opportunity.

Thank you for considering my application. I appreciate your time and consideration.

Best regards,
{applicant_name}"""

        cover_letter = f"""Date: {date_str}

{opening}

{body}

Attachment: Resume"""

        cover_letter_clean = f"""{opening}

{body}

Sincerely,
{applicant_name}"""

        return {
            "ok": True,
            "cover_letter": cover_letter_clean,
            "full_cover_letter": cover_letter,
            "applicant_name": applicant_name,
            "job_title": job_title,
            "company_name": company_name,
            "date": date_str,
            "generated_at": datetime.now().isoformat()
        }

    except Exception as e:
        return {"ok": False, "error": f"Failed to generate cover letter: {str(e)}"}


if __name__ == "__main__":
    result = handler({
        "applicant_name": "Sarah Chen",
        "job_title": "Senior Software Engineer",
        "company_name": "TechCorp Inc",
        "key_qualifications": ["software development", "team leadership", "cloud architecture"],
        "years_experience": 7
    })
    print(result)
