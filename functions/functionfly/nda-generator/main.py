"""NDA Generator - Generate Non-Disclosure Agreements."""
import hashlib
from datetime import datetime
from typing import Any


def handler(event: dict) -> dict:
    """Generate a Non-Disclosure Agreement."""
    try:
        disclosing_party = event.get("disclosing_party")
        receiving_party = event.get("receiving_party")
        effective_date = event.get("effective_date")
        confidential_info_description = event.get("confidential_info_description")
        term_years = event.get("term_years", 2)

        if not disclosing_party:
            return {"ok": False, "error": "disclosing_party is required"}
        if not receiving_party:
            return {"ok": False, "error": "receiving_party is required"}
        if not effective_date:
            return {"ok": False, "error": "effective_date is required"}
        if not confidential_info_description:
            return {"ok": False, "error": "confidential_info_description is required"}

        try:
            parsed_date = datetime.strptime(effective_date, "%Y-%m-%d")
        except ValueError:
            return {"ok": False, "error": "effective_date must be in YYYY-MM-DD format"}

        if not isinstance(term_years, int) or term_years < 1:
            return {"ok": False, "error": "term_years must be a positive integer"}

        hash_val = int(hashlib.md5((disclosing_party + receiving_party).encode()).hexdigest()[:8], 16)
        nda_id = f"NDA-{hash_val:04d}-{datetime.now().strftime('%Y%m%d')}"

        jurisdiction = "State of New York, County of New York"

        exclusions = [
            "Information that is publicly known at the time of disclosure or becomes publicly known through no fault of the Receiving Party",
            "Information that was rightfully in the Receiving Party's possession prior to disclosure",
            "Information that is independently developed by the Receiving Party without use of the Confidential Information",
            "Information that is rightfully obtained by the Receiving Party from a third party without restriction",
            "Information that is disclosed pursuant to court order or legal requirement, provided the Receiving Party gives prompt notice"
        ]

        nda_text = f"""
═══════════════════════════════════════════════════════════════════════════════
          CONFIDENTIALITY AND NON-DISCLOSURE AGREEMENT (NDA)
═══════════════════════════════════════════════════════════════════════════════

NDA Reference No.: {nda_id}
Effective Date: {effective_date}

═══════════════════════════════════════════════════════════════════════════════
                                    PARTIES
═══════════════════════════════════════════════════════════════════════════════

DISCLOSING PARTY:
{disclosing_party}
(hereinafter referred to as the "Disclosing Party")

RECEIVING PARTY:
{receiving_party}
(hereinafter referred to as the "Receiving Party")

(Each a "Party" and collectively the "Parties")

═══════════════════════════════════════════════════════════════════════════════
                                  RECITALS
═══════════════════════════════════════════════════════════════════════════════

WHEREAS, the Disclosing Party possesses certain confidential and proprietary
information relating to its business; and

WHEREAS, the Receiving Party desires to receive certain Confidential Information
for the purpose of evaluating a potential business relationship;

NOW, THEREFORE, in consideration of the mutual covenants and agreements herein
contained, the Parties agree as follows:

═══════════════════════════════════════════════════════════════════════════════
                           ARTICLE 1: DEFINITIONS
═══════════════════════════════════════════════════════════════════════════════

1.1 "Confidential Information" means any and all information disclosed by the
Disclosing Party to the Receiving Party, including but not limited to:

{confidential_info_description}

1.2 Confidential Information includes, without limitation:
- Trade secrets, inventions, and proprietary knowledge
- Business plans, strategies, and financial information
- Customer lists, supplier information, and pricing data
- Technical data, designs, and know-how
- Any other information marked or identified as confidential

═══════════════════════════════════════════════════════════════════════════════
                        ARTICLE 2: OBLIGATIONS
═══════════════════════════════════════════════════════════════════════════════

2.1 The Receiving Party agrees to:
(a) Hold all Confidential Information in strict confidence
(b) Not disclose Confidential Information to any third parties without prior
    written consent from the Disclosing Party
(c) Use the Confidential Information solely for the Purpose
(d) Take reasonable measures to protect the secrecy of the Confidential
    Information, including all precautions the Receiving Party employs with
    respect to its own confidential materials

2.2 The Receiving Party may disclose Confidential Information only to:
(a) Employees and contractors who have a need to know
(b) Professional advisors bound by confidentiality obligations
(c) As required by applicable law or court order

═══════════════════════════════════════════════════════════════════════════════
                         ARTICLE 3: EXCLUSIONS
═══════════════════════════════════════════════════════════════════════════════

3.1 The obligations under this Agreement do not apply to information that:

(a) Is or becomes publicly known through no breach of this Agreement
(b) Was rightfully in the Receiving Party's possession prior to disclosure
(c) Is independently developed without use of the Confidential Information
(d) Is rightfully obtained from a third party without restriction
(e) Is disclosed pursuant to court order, provided notice is given to Disclosing
    Party (when legally permitted) and the Receiving Party cooperates in seeking
    protective measures

═══════════════════════════════════════════════════════════════════════════════
                            ARTICLE 4: TERM
═══════════════════════════════════════════════════════════════════════════════

4.1 This Agreement shall remain in effect for a period of {term_years} year(s)
from the Effective Date.

4.2 The confidentiality obligations shall survive termination for a period of
{term_years + 1} years.

═══════════════════════════════════════════════════════════════════════════════
                    ARTICLE 5: RETURN OF MATERIALS
═══════════════════════════════════════════════════════════════════════════════

5.1 Upon termination or request by the Disclosing Party, the Receiving Party
shall promptly return or destroy all Confidential Information and any copies
thereof, and shall certify such destruction in writing upon request.

═══════════════════════════════════════════════════════════════════════════════
                           ARTICLE 6: REMEDIES
═══════════════════════════════════════════════════════════════════════════════

6.1 The Parties acknowledge that monetary damages may be inadequate for breach
of this Agreement. Accordingly, either Party may seek injunctive relief in
addition to any other available remedies.

6.2 The Receiving Party shall notify the Disclosing Party promptly upon
discovery of any unauthorized use or disclosure of Confidential Information.

═══════════════════════════════════════════════════════════════════════════════
                        ARTICLE 7: GENERAL PROVISIONS
═══════════════════════════════════════════════════════════════════════════════

7.1 GOVERNING LAW: This Agreement shall be governed by the laws of {jurisdiction}.

7.2 ENTIRE AGREEMENT: This Agreement constitutes the entire agreement between
the Parties with respect to the subject matter hereof.

7.3 AMENDMENT: This Agreement may not be modified except by written instrument
signed by both Parties.

7.4 ASSIGNMENT: Neither Party may assign this Agreement without prior written
consent of the other Party.

7.5 SEVERABILITY: If any provision is found unenforceable, the remaining
provisions shall remain in full force and effect.

═══════════════════════════════════════════════════════════════════════════════
                              SIGNATURES
═══════════════════════════════════════════════════════════════════════════════

IN WITNESS WHEREOF, the Parties have executed this Agreement as of the date
first written above.

DISCLOSING PARTY: {disclosing_party}

By: _________________________________
Name:
Title:
Date:

RECEIVING PARTY: {receiving_party}

By: _________________________________
Name:
Title:
Date:

═══════════════════════════════════════════════════════════════════════════════
                              EXHIBIT A
                         EXCLUSIONS LIST
═══════════════════════════════════════════════════════════════════════════════

The following categories of information are excluded from the definition of
Confidential Information:

1. Publicly available information not obtained through breach of this Agreement
2. Information previously known to Receiving Party before disclosure
3. Information independently developed without reference to Confidential Information
4. Information obtained from third parties without confidentiality restrictions
5. Information disclosed pursuant to valid court order or legal requirement

═══════════════════════════════════════════════════════════════════════════════
"""

        return {
            "ok": True,
            "nda_id": nda_id,
            "nda_text": nda_text.strip(),
            "disclosing_party": disclosing_party,
            "receiving_party": receiving_party,
            "effective_date": effective_date,
            "term_years": term_years,
            "jurisdiction": jurisdiction,
            "exclusions": exclusions,
            "generated_at": datetime.now().isoformat()
        }

    except Exception as e:
        return {"ok": False, "error": f"Failed to generate NDA: {str(e)}"}


if __name__ == "__main__":
    result = handler({
        "disclosing_party": "Tech Innovations Inc.",
        "receiving_party": "Strategic Partners LLC",
        "effective_date": "2026-06-15",
        "confidential_info_description": "Proprietary software algorithms, source code, technical specifications, product roadmaps, customer data, and business strategies",
        "term_years": 3
    })
    print(result)
