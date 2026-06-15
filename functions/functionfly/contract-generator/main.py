"""Contract Generator - Generate simple legal contracts."""
import re
from datetime import datetime
from typing import Any


def validate_contract_type(contract_type: str) -> bool:
    """Validate contract type."""
    return contract_type in ["nda", "employment", "service"]


def generate_nda_sections(disclosing_party: str, receiving_party: str, effective_date: str,
                          confidential_info_description: str, term_years: int = 2) -> tuple:
    """Generate NDA contract sections."""
    sections = [
        "1. DEFINITION OF CONFIDENTIAL INFORMATION",
        "2. OBLIGATIONS OF RECEIVING PARTY",
        "3. EXCLUSIONS FROM CONFIDENTIAL INFORMATION",
        "4. TERM AND TERMINATION",
        "5. RETURN OF CONFIDENTIAL INFORMATION",
        "6. REMEDIES",
        "7. GENERAL PROVISIONS"
    ]

    text = f"""
CONFIDENTIALITY AND NON-DISCLOSURE AGREEMENT

This Confidentiality and Non-Disclosure Agreement ("Agreement") is entered into as of {effective_date}
by and between:

DISCLOSING PARTY: {disclosing_party}
RECEIVING PARTY: {receiving_party}

1. PURPOSE
The parties wish to explore a business opportunity of mutual interest. In connection with this
opportunity, Disclosing Party may share certain confidential information with Receiving Party.

2. DEFINITION OF CONFIDENTIAL INFORMATION
"Confidential Information" means {confidential_info_description or "any and all information shared by the Disclosing Party"}
that is designated as confidential or that reasonably should be understood to be confidential
given the nature of the information and circumstances of disclosure.

3. OBLIGATIONS OF RECEIVING PARTY
The Receiving Party agrees to:
a) Hold the Confidential Information in strict confidence
b) Not disclose the Confidential Information to any third parties
c) Use the Confidential Information solely for the purpose of evaluating the business opportunity
d) Take reasonable measures to protect the secrecy of the Confidential Information

4. EXCLUSIONS
Confidential Information does not include information that:
a) Was publicly known prior to disclosure
b) Becomes publicly known through no fault of Receiving Party
c) Was rightfully in Receiving Party's possession prior to disclosure
d) Is independently developed by Receiving Party without use of Confidential Information

5. TERM
This Agreement shall remain in effect for a period of {term_years} year(s) from the Effective Date.

6. RETURN OF MATERIALS
Upon termination or request, Receiving Party shall return or destroy all Confidential Information.

7. REMEDIES
The parties acknowledge that monetary damages may be inadequate for breach of this Agreement.
Accordingly, either party may seek injunctive relief in addition to other available remedies.

8. GENERAL PROVISIONS
This Agreement shall be governed by the laws of the State of New York.
"""

    signature_lines = f"""
DISCLOSING PARTY: {disclosing_party}

By: _______________________________
Name:
Title:
Date:

RECEIVING PARTY: {receiving_party}

By: _______________________________
Name:
Title:
Date:
"""

    return text.strip(), sections, signature_lines


def generate_employment_sections(employee_name: str, employer_name: str, effective_date: str,
                                  job_title: str, responsibilities: list, requirements: list,
                                  benefits: list) -> tuple:
    """Generate employment contract sections."""
    sections = [
        "1. POSITION AND DUTIES",
        "2. COMPENSATION",
        "3. BENEFITS",
        "4. TERM OF EMPLOYMENT",
        "5. CONFIDENTIALITY",
        "6. NON-COMPETE",
        "7. TERMINATION",
        "8. GENERAL PROVISIONS"
    ]

    resp_text = "\n".join([f"   • {r}" for r in responsibilities]) if responsibilities else "   • Per job description"

    req_text = "\n".join([f"   • {r}" for r in requirements]) if requirements else "   • Per job requirements"
    ben_text = "\n".join([f"   • {b}" for b in benefits]) if benefits else "   • Per company policy"

    text = f"""
EMPLOYMENT AGREEMENT

This Employment Agreement ("Agreement") is entered into as of {effective_date}
by and between:

EMPLOYER: {employer_name}
EMPLOYEE: {employee_name}

1. POSITION AND DUTIES
Employee's Title: {job_title}

Employee agrees to perform the following responsibilities:
{resp_text}

2. COMPENSATION
Employee shall receive compensation as agreed upon separately in an offer letter.

3. BENEFITS
Employee shall be entitled to the following benefits:
{ben_text}

4. TERM OF EMPLOYMENT
This Agreement shall be for an indefinite term, subject to the termination provisions below.

5. CONFIDENTIALITY
Employee agrees to maintain the confidentiality of all proprietary information.

6. NON-COMPETE
Employee agrees not to engage in competitive employment during the term of this Agreement
and for a reasonable period thereafter.

7. TERMINATION
Either party may terminate this Agreement with notice as specified in company policy.

8. GENERAL PROVISIONS
This Agreement shall be governed by applicable state and federal laws.
"""

    signature_lines = f"""
EMPLOYER: {employer_name}

By: _______________________________
Name:
Title:
Date:

EMPLOYEE: {employee_name}

By: _______________________________
Name: {employee_name}
Title: {job_title}
Date:
"""

    return text.strip(), sections, signature_lines


def generate_service_sections(party_a: str, party_b: str, effective_date: str,
                                services: list, terms: str) -> tuple:
    """Generate service contract sections."""
    sections = [
        "1. SERVICES TO BE PROVIDED",
        "2. PAYMENT TERMS",
        "3. TIMELINE",
        "4. DELIVERABLES",
        "5. INTELLECTUAL PROPERTY",
        "6. CONFIDENTIALITY",
        "7. LIMITATION OF LIABILITY",
        "8. TERMINATION",
        "9. GENERAL PROVISIONS"
    ]

    services_text = "\n".join([f"   {i+1}. {s}" for i, s in enumerate(services)]) if services else "   Per attached scope of work"

    text = f"""
SERVICE AGREEMENT

This Service Agreement ("Agreement") is entered into as of {effective_date}
by and between:

SERVICE PROVIDER: {party_a}
CLIENT: {party_b}

1. SERVICES TO BE PROVIDED
The Service Provider agrees to perform the following services:
{services_text}

2. PAYMENT TERMS
Payment shall be made according to the schedule specified in Exhibit A.

3. TIMELINE
Services shall commence on the Effective Date and continue per the agreed timeline.

4. DELIVERABLES
All deliverables shall be provided as specified in the attached statement of work.

5. INTELLECTUAL PROPERTY
Upon full payment, intellectual property rights shall transfer to the Client.

6. CONFIDENTIALITY
Both parties agree to maintain confidentiality of proprietary information.

7. LIMITATION OF LIABILITY
Liability shall not exceed the total fees paid under this Agreement.

8. TERMINATION
Either party may terminate with 30 days written notice.

9. GENERAL PROVISIONS
This Agreement shall be governed by the laws of the State of New York.

TERMS: {terms or "Standard terms and conditions as attached."}
"""

    signature_lines = f"""
SERVICE PROVIDER: {party_a}

By: _______________________________
Name:
Title:
Date:

CLIENT: {party_b}

By: _______________________________
Name:
Title:
Date:
"""

    return text.strip(), sections, signature_lines


def handler(event: dict) -> dict:
    """Generate a simple contract document."""
    try:
        contract_type = event.get("contract_type")
        party_a = event.get("party_a")
        party_b = event.get("party_b")
        effective_date = event.get("effective_date")
        terms = event.get("terms")

        if not contract_type:
            return {"ok": False, "error": "contract_type is required (nda/employment/service)"}
        if not party_a:
            return {"ok": False, "error": "party_a is required"}
        if not party_b:
            return {"ok": False, "error": "party_b is required"}
        if not effective_date:
            return {"ok": False, "error": "effective_date is required"}

        if not validate_contract_type(contract_type):
            return {"ok": False, "error": "contract_type must be one of: nda, employment, service"}

        try:
            parsed_date = datetime.strptime(effective_date, "%Y-%m-%d")
        except ValueError:
            return {"ok": False, "error": "effective_date must be in YYYY-MM-DD format"}

        contract_id = f"CONTRACT-{contract_type.upper()}-{datetime.now().strftime('%Y%m%d')}-{hash(party_a + party_b) % 10000:04d}"

        if contract_type == "nda":
            confidential_info = event.get("confidential_info_description", "")
            term_years = event.get("term_years", 2)
            if not isinstance(term_years, int) or term_years < 1:
                return {"ok": False, "error": "term_years must be a positive integer"}
            contract_text, sections, signature_lines = generate_nda_sections(
                party_a, party_b, effective_date, confidential_info, term_years
            )

        elif contract_type == "employment":
            job_title = event.get("job_title", "Employee")
            responsibilities = event.get("responsibilities", [])
            requirements = event.get("requirements", [])
            benefits = event.get("benefits", [])
            contract_text, sections, signature_lines = generate_employment_sections(
                party_b, party_a, effective_date, job_title, responsibilities, requirements, benefits
            )

        else:
            services = event.get("services", terms.split('\n') if terms else [])
            contract_text, sections, signature_lines = generate_service_sections(
                party_a, party_b, effective_date, services, terms or ""
            )

        return {
            "ok": True,
            "contract_id": contract_id,
            "contract_type": contract_type,
            "contract_text": contract_text,
            "sections": sections,
            "signature_lines": signature_lines.strip(),
            "effective_date": effective_date,
            "generated_at": datetime.now().isoformat()
        }

    except Exception as e:
        return {"ok": False, "error": f"Failed to generate contract: {str(e)}"}


if __name__ == "__main__":
    result = handler({
        "contract_type": "nda",
        "party_a": "Acme Corp",
        "party_b": "Beta LLC",
        "effective_date": "2026-06-15",
        "confidential_info_description": "Trade secrets, business strategies, customer lists, and financial information",
        "term_years": 2
    })
    print(result)
