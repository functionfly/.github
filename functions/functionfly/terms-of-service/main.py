"""Terms of Service Generator - Generate Terms of Service documents."""
from datetime import datetime
from typing import Any


def handler(event: dict) -> dict:
    """Generate Terms of Service."""
    try:
        company_name = event.get("company_name")
        website_url = event.get("website_url")
        services_description = event.get("services_description")
        user_obligations = event.get("user_obligations", [])

        if not company_name:
            return {"ok": False, "error": "company_name is required"}
        if not website_url:
            return {"ok": False, "error": "website_url is required"}
        if not services_description:
            return {"ok": False, "error": "services_description is required"}

        today = datetime.now()
        effective_date = today.strftime("%B %d, %Y")

        sections = [
            "1. Acceptance of Terms",
            "2. Description of Services",
            "3. User Accounts and Eligibility",
            "4. User Obligations",
            "5. Intellectual Property Rights",
            "6. Privacy and Data Handling",
            "7. User Content and Conduct",
            "8. Third-Party Links and Services",
            "9. Termination",
            "10. Disclaimer of Warranties",
            "11. Limitation of Liability",
            "12. Indemnification",
            "13. Governing Law and Dispute Resolution",
            "14. Changes to Terms",
            "15. Contact Information"
        ]

        obligations_text = "\n".join([f"   • {obligation}" for obligation in user_obligations]) if user_obligations else "   • Use the services in accordance with these Terms\n   • Provide accurate and complete information\n   • Maintain the security of your account\n   • Comply with all applicable laws and regulations"

        liability_limitations = """NO WARRANTIES: THE SERVICES ARE PROVIDED "AS IS" AND "AS AVAILABLE" WITHOUT WARRANTIES OF ANY KIND, WHETHER EXPRESS, IMPLIED, OR STATUTORY.

LIMITATION OF LIABILITY: TO THE MAXIMUM EXTENT PERMITTED BY LAW, {COMPANY} SHALL NOT BE LIABLE FOR ANY INDIRECT, INCIDENTAL, SPECIAL, CONSEQUENTIAL, OR PUNITIVE DAMAGES, OR ANY LOSS OF PROFITS, REVENUE, DATA, OR GOODWILL, ARISING OUT OF OR RELATED TO YOUR USE OF THE SERVICES.

MAXIMUM LIABILITY: IN NO EVENT SHALL {COMPANY}'S TOTAL LIABILITY ARISING OUT OF OR RELATED TO THESE TERMS EXCEED THE AMOUNT OF FEES PAID BY YOU TO {COMPANY} IN THE TWELVE (12) MONTHS PRECEDING THE CLAIM.

SOME JURISDICTIONS DO NOT ALLOW THE EXCLUSION OF IMPLIED WARRANTIES OR LIMITATIONS ON LIABILITY, SO THE ABOVE EXCLUSIONS MAY NOT APPLY TO YOU.""".format(COMPANY=company_name.upper())

        tos_text = f"""
TERMS OF SERVICE

Effective Date: {effective_date}
Last Updated: {effective_date}

Welcome to {website_url}. These Terms of Service ("Terms") govern your access to and use of the services provided by {company_name} ("we," "us," or "our").

Please read these Terms carefully before accessing or using our services.

{'='*70}
{sections[0]}
{'='*70}

By accessing or using our services, you agree to be bound by these Terms. If you do not agree to these Terms, you may not access or use the services.

{'='*70}
{sections[1]}
{'='*70}

{services_description}

We reserve the right to modify, suspend, or discontinue any part of the services at any time.

{'='*70}
{sections[2]}
{'='*70}

To access certain features, you may need to create an account. You agree to:
   • Provide accurate, current, and complete information
   • Maintain and update your information to keep it accurate
   • Keep your password secure and confidential
   • Notify us immediately of any unauthorized access

You must be at least 18 years old to create an account. By creating an account, you represent that you meet this requirement.

{'='*70}
{sections[3]}
{'='*70}

As a user of our services, you agree to:
{obligations_text}

{'='*70}
{sections[4]}
{'='*70}

All content, features, and functionality of our services are owned by {company_name} and are protected by copyright, trademark, and other intellectual property laws.

{'='*70}
{sections[5]}
{'='*70}

Your privacy is important to us. Our Privacy Policy explains how we collect, use, and protect your information. By using our services, you agree to our Privacy Policy.

{'='*70}
{sections[6]}
{'='*70}

You are responsible for all content you post and your conduct on our services. You agree not to:
   • Post content that is harmful, offensive, or inappropriate
   • Violate any person's privacy or intellectual property rights
   • Use the services for any illegal purpose
   • Attempt to gain unauthorized access to any systems

{'='*70}
{sections[7]}
{'='*70}

Our services may contain links to third-party websites or services. We do not endorse or assume responsibility for any third-party content or services.

{'='*70}
{sections[8]}
{'='*70}

We may terminate or suspend your access to the services at any time, with or without cause, including if you violate these Terms.

{'='*70}
{sections[9]}
{'='*70}

NO WARRANTIES: THE SERVICES ARE PROVIDED "AS IS" AND "AS AVAILABLE" WITHOUT WARRANTIES OF ANY KIND, WHETHER EXPRESS, IMPLIED, OR STATUTORY, INCLUDING WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE, AND NON-INFRINGEMENT.

{'='*70}
{sections[10]}
{'='*70}

{liability_limitations}

{'='*70}
{sections[11]}
{'='*70}

You agree to indemnify and hold harmless {company_name} and its officers, directors, employees, and agents from any claims, damages, losses, or expenses arising out of your use of the services or your violation of these Terms.

{'='*70}
{sections[12]}
{'='*70}

These Terms shall be governed by the laws of the State of [State], without regard to its conflict of law provisions. Any disputes shall be resolved in the courts of [State].

{'='*70}
{sections[13]}
{'='*70}

We may update these Terms from time to time. We will notify you of material changes by posting the new Terms on this page and updating the "Last Updated" date.

{'='*70}
{sections[14]}
{'='*70}

If you have questions about these Terms, please contact us at:

{company_name}
Email: legal@{website_url.replace('www.', '').replace('https://', '').replace('http://', '')}
Address: [Company Address]

{'='*70}
"""

        return {
            "ok": True,
            "company_name": company_name,
            "website_url": website_url,
            "tos_text": tos_text.strip(),
            "sections": sections,
            "liability_limitations": liability_limitations,
            "effective_date": effective_date,
            "generated_at": datetime.now().isoformat()
        }

    except Exception as e:
        return {"ok": False, "error": f"Failed to generate Terms of Service: {str(e)}"}


if __name__ == "__main__":
    result = handler({
        "company_name": "TechCorp Inc",
        "website_url": "https://techcorp.example.com",
        "services_description": "TechCorp provides cloud-based productivity tools and collaboration software designed to help teams work more efficiently.",
        "user_obligations": [
            "Use the platform in accordance with our guidelines",
            "Do not attempt to reverse engineer the software",
            "Report any security vulnerabilities you discover",
            "Respect other users and maintain professional conduct"
        ]
    })
    print(result)
