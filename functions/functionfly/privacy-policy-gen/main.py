"""Privacy Policy Generator - Generate privacy policies."""
from datetime import datetime
from typing import Any


def handler(event: dict) -> dict:
    """Generate a privacy policy."""
    try:
        company_name = event.get("company_name")
        website_url = event.get("website_url")
        data_collected = event.get("data_collected", [])
        third_party_sharing = event.get("third_party_sharing", [])
        user_rights = event.get("user_rights", [])

        if not company_name:
            return {"ok": False, "error": "company_name is required"}
        if not website_url:
            return {"ok": False, "error": "website_url is required"}
        if not data_collected or len(data_collected) == 0:
            return {"ok": False, "error": "data_collected list is required"}

        today = datetime.now()
        effective_date = today.strftime("%B %d, %Y")

        sections = [
            "1. Information We Collect",
            "2. How We Use Your Information",
            "3. Information Sharing and Disclosure",
            "4. Data Security",
            "5. Cookies and Tracking Technologies",
            "6. Your Rights and Choices",
            "7. Children's Privacy",
            "8. International Data Transfers",
            "9. Changes to This Policy",
            "10. Contact Us"
        ]

        data_collected_text = "\n".join([f"   • {item}" for item in data_collected])

        third_party_text = "\n".join([f"   • {item}" for item in third_party_sharing]) if third_party_sharing else "   • We do not sell or share your personal information with third parties for their marketing purposes."

        default_user_rights = [
            "Access your personal information",
            "Correct inaccurate information",
            "Delete your personal information",
            "Object to processing of your information",
            "Data portability",
            "Withdraw consent"
        ]

        user_rights_text = "\n".join([f"   • {right}" for right in (user_rights if user_rights else default_user_rights)])

        policy_text = f"""
PRIVACY POLICY

Effective Date: {effective_date}
Last Updated: {effective_date}

{company_name} ("we," "our," or "us") is committed to protecting your privacy. This Privacy Policy explains how we collect, use, disclose, and safeguard your information when you visit our website at {website_url}.

Please read this Privacy Policy carefully. By accessing or using {website_url}, you acknowledge that you have read, understood, and agree to be bound by all the terms of this Privacy Policy.

{'='*70}
{sections[0]}
{'='*70}

We collect information that you provide directly to us, including:
{data_collected_text}

We also automatically collect certain information when you visit our website, including:
   • IP address and device identifiers
   • Browser type and version
   • Operating system
   • Referring and exit pages
   • Date and time of your visit
   • Pages viewed and time spent on pages

{'='*70}
{sections[1]}
{'='*70}

We use the information we collect to:
   • Provide, maintain, and improve our services
   • Process transactions and send related information
   • Send you technical notices and support messages
   • Respond to your comments and questions
   • Communicate with you about products and services
   • Monitor and analyze trends and usage
   • Detect, investigate, and prevent fraudulent transactions
   • Personalize and improve your experience

{'='*70}
{sections[2]}
{'='*70}

We may share your information with:
{third_party_text}

We may also share information in the following circumstances:
   • With your consent or at your direction
   • To comply with legal obligations
   • To protect our rights, privacy, safety, or property
   • In connection with a merger, acquisition, or sale of assets

{'='*70}
{sections[3]}
{'='*70}

We implement appropriate technical and organizational security measures to protect your personal information against unauthorized access, alteration, disclosure, or destruction. However, no method of transmission over the Internet is 100% secure.

{'='*70}
{sections[4]}
{'='*70}

We use cookies and similar tracking technologies to collect and use information about your visits. You can control cookies through your browser settings.

{'='*70}
{sections[5]}
{'='*70}

You have the following rights regarding your personal information:
{user_rights_text}

To exercise these rights, please contact us using the information provided in the "Contact Us" section below.

{'='*70}
{sections[6]}
{'='*70}

Our services are not intended for children under 13 years of age. We do not knowingly collect personal information from children under 13.

{'='*70}
{sections[7]}
{'='*70}

Your information may be transferred to and processed in countries other than your country of residence. We ensure appropriate safeguards are in place for such transfers.

{'='*70}
{sections[8]}
{'='*70}

We may update this Privacy Policy from time to time. We will notify you of any changes by posting the new Privacy Policy on this page and updating the "Last Updated" date.

{'='*70}
{sections[9]}
{'='*70}

If you have questions about this Privacy Policy, please contact us at:

{company_name}
Email: privacy@{website_url.replace('www.', '').replace('https://', '').replace('http://', '')}
Address: [Company Address]

{'='*70}
"""

        compliance_notes = [
            "This policy template is designed to comply with GDPR and CCPA requirements",
            "Consult with legal counsel to ensure compliance with specific jurisdiction requirements",
            "Review and update this policy at least annually or when regulations change",
            "Maintain documentation of consent mechanisms and data processing activities"
        ]

        return {
            "ok": True,
            "company_name": company_name,
            "website_url": website_url,
            "policy_text": policy_text.strip(),
            "sections": sections,
            "compliance_notes": compliance_notes,
            "effective_date": effective_date,
            "generated_at": datetime.now().isoformat()
        }

    except Exception as e:
        return {"ok": False, "error": f"Failed to generate privacy policy: {str(e)}"}


if __name__ == "__main__":
    result = handler({
        "company_name": "TechCorp Inc",
        "website_url": "https://techcorp.example.com",
        "data_collected": [
            "Name and email address",
            "Billing and shipping address",
            "Payment information",
            "Phone number",
            "Company name (if applicable)"
        ],
        "third_party_sharing": [
            "Payment processors (Stripe, PayPal)",
            "Shipping providers",
            "Analytics services"
        ],
        "user_rights": [
            "Right to access your data",
            "Right to deletion",
            "Right to opt-out of marketing"
        ]
    })
    print(result)
