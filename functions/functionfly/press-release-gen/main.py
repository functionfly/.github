"""Press Release Generator - Generate AP-style press releases."""
from datetime import datetime
from typing import Any


ANNOUNCEMENT_CONFIG = {
    "product": {
        "boilerplate_addition": "leading the industry in product innovation",
        "quote_prompt": "how this product solves customer problems"
    },
    "award": {
        "boilerplate_addition": "recognized for excellence and innovation",
        "quote_prompt": "what this award means to the company"
    },
    "partnership": {
        "boilerplate_addition": "building strategic alliances",
        "quote_prompt": "the strategic value of this partnership"
    },
    "milestone": {
        "boilerplate_addition": "celebrating important milestones",
        "quote_prompt": "what this milestone means for stakeholders"
    }
}


def handler(event: dict) -> dict:
    """Generate a press release."""
    try:
        company_name = event.get("company_name")
        announcement_type = event.get("announcement_type")
        headline = event.get("headline")
        details = event.get("details")
        quote = event.get("quote")

        if not company_name:
            return {"ok": False, "error": "company_name is required"}
        if not announcement_type:
            return {"ok": False, "error": "announcement_type is required"}
        if not headline:
            return {"ok": False, "error": "headline is required"}
        if not details:
            return {"ok": False, "error": "details is required"}

        valid_types = ["product", "award", "partnership", "milestone"]
        if announcement_type not in valid_types:
            return {"ok": False, "error": f"announcement_type must be one of: {', '.join(valid_types)}"}

        today = datetime.now()
        date_str = today.strftime("%B %d, %Y")
        timestamp = f"{date_str} / {company_name}"

        config = ANNOUNCEMENT_CONFIG.get(announcement_type, ANNOUNCEMENT_CONFIG["milestone"])

        boilerplate = f"""
About {company_name}
{company_name} is {config['boilerplate_addition']}. Founded in [Year], the company has been dedicated to [mission statement]. For more information, visit [website URL] or follow us on [social media handles].
"""

        media_contact = f"""
Media Contact:
[Contact Name]
[Title]
{company_name}
[email@company.com]
[Phone Number]
"""

        if not quote:
            quote = f"\"[This {announcement_type} represents our commitment to excellence and innovation. We are excited about the opportunities it creates for our customers and stakeholders.]\""

        press_release_text = f"""
FOR IMMEDIATE RELEASE
{date_str}

{headline.upper()}

{company_name} Announces {announcement_type.title()}

{company_name} today announced [details about the {announcement_type}].

{details}

"{quote.replace('"', '')}" said [Name], [Title] at {company_name}.

About {company_name}
{boilerplate}

###
{media_contact}
"""

        return {
            "ok": True,
            "company_name": company_name,
            "announcement_type": announcement_type,
            "headline": headline,
            "press_release_text": press_release_text.strip(),
            "boilerplate": boilerplate.strip(),
            "media_contact": media_contact.strip(),
            "date": date_str,
            "generated_at": datetime.now().isoformat()
        }

    except Exception as e:
        return {"ok": False, "error": f"Failed to generate press release: {str(e)}"}


if __name__ == "__main__":
    result = handler({
        "company_name": "TechCorp Inc",
        "announcement_type": "product",
        "headline": "TechCorp Launches Revolutionary AI-Powered Analytics Platform",
        "details": "The new platform leverages advanced machine learning algorithms to provide real-time insights for enterprise customers, reducing analysis time from hours to seconds.",
        "quote": "This platform represents a fundamental shift in how businesses will interact with their data."
    })
    print(result)
