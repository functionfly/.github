"""Affidavit Generator - Generate legal affidavits."""
import re
from datetime import datetime
from typing import Any


def handler(event: dict) -> dict:
    """Generate a legal affidavit document."""
    try:
        full_name = event.get("full_name")
        statement = event.get("statement")
        date = event.get("date")
        notary_state = event.get("notary_state")

        if not full_name:
            return {"ok": False, "error": "full_name is required"}
        if not statement:
            return {"ok": False, "error": "statement is required"}
        if not date:
            return {"ok": False, "error": "date is required"}
        if not notary_state:
            return {"ok": False, "error": "notary_state is required"}

        if len(full_name) < 2:
            return {"ok": False, "error": "full_name must be at least 2 characters"}
        if len(statement) < 10:
            return {"ok": False, "error": "statement must be at least 10 characters"}

        valid_states = [
            "AL", "AK", "AZ", "AR", "CA", "CO", "CT", "DE", "FL", "GA", "HI", "ID", "IL", "IN", "IA",
            "KS", "KY", "LA", "ME", "MD", "MA", "MI", "MN", "MS", "MO", "MT", "NE", "NV", "NH", "NJ",
            "NM", "NY", "NC", "ND", "OH", "OK", "OR", "PA", "RI", "SC", "SD", "TN", "TX", "UT", "VT",
            "VA", "WA", "WV", "WI", "WY", "DC"
        ]
        if notary_state.upper() not in valid_states:
            return {"ok": False, "error": f"notary_state must be a valid US state code, got: {notary_state}"}

        try:
            parsed_date = datetime.strptime(date, "%Y-%m-%d")
        except ValueError:
            return {"ok": False, "error": "date must be in YYYY-MM-DD format"}

        affidavit_id = f"AFF-{datetime.now().strftime('%Y%m%d')}-{hash(full_name) % 10000:04d}"

        affidavit_text = f"""
═══════════════════════════════════════════════════════════════════════════════
                                 AFFIDAVIT
═══════════════════════════════════════════════════════════════════════════════

AFFIDAVIT NO.: {affidavit_id}

STATE OF {notary_state.upper()}
COUNTY OF ________________

Before me, the undersigned notary public, personally appeared:

    {full_name.upper()}

who, being duly sworn, deposes and says:

    {statement}

The undersigned notary public certifies that the above-named person appeared
before me on this {parsed_date.strftime('%drd' if parsed_date.day == 3 else 'th') if parsed_date.day in [1,21,31] else parsed_date.strftime('%dth')} day of {parsed_date.strftime('%B, %Y')},
and is personally known to me (or proved to me on the basis of satisfactory
evidence) to be the person whose name is subscribed to this instrument.

WITNESS my hand and official seal:




_________________________________________________
Notary Public, {notary_state.upper()}
My Commission Expires: _________________________

[NOTARY SEAL]

═══════════════════════════════════════════════════════════════════════════════
"""

        jurat_text = f"""
JURAT

State of {notary_state.upper()}
County of ________________

Subscribed and sworn to before me on this {parsed_date.strftime('%drd' if parsed_date.day == 3 else 'th') if parsed_date.day in [1,21,31] else parsed_date.strftime('%dth')} day of {parsed_date.strftime('%B, %Y')}.


_________________________________________________
Notary Public Signature

_________________________________________________
Notary Public Name (Print)

My Commission Number: _________________________
My Commission Expires: ________________________
"""

        return {
            "ok": True,
            "affidavit_id": affidavit_id,
            "affidavit_text": affidavit_text.strip(),
            "jurat_text": jurat_text.strip(),
            "full_name": full_name,
            "notary_state": notary_state.upper(),
            "date": date,
            "generated_at": datetime.now().isoformat()
        }

    except Exception as e:
        return {"ok": False, "error": f"Failed to generate affidavit: {str(e)}"}


if __name__ == "__main__":
    result = handler({
        "full_name": "John Michael Smith",
        "statement": "I hereby certify that the information provided in this document is true and accurate to the best of my knowledge.",
        "date": "2026-06-15",
        "notary_state": "CA"
    })
    print(result)
