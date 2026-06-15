import re
from typing import Any


RISK_KEYWORDS = {
    "high": [
        r"\b(penal(?:ty|ties)|fine|fines|damages|indemnif(?:y|ication)|liabilit(?:y|ies)|unlimited|risk|solely|entire)\b",
        r"\b(exclusively|waive(?:s|d|r)?|forfeit(?:s|ed)?|irrevocable|perpetual)\b",
        r"\b(guarantee|warranty|as-is|without warranty|hold harmless|defend\s+and\s+hold)\b",
    ],
    "medium": [
        r"\b(terminat(?:e|ion|ing)|cancel(?:lation)?|renewal|automatic\s+renew|price\s+increase)\b",
        r"\b(limit(?:s|ed|ation)|cap(?:ped|s)?|maximum|threshold|conditional)\b",
        r"\b(notice\s+period|breach|default|remed(?:y|ies)|dispute|arbitrat(?:e|ion))\b",
        r"\b(assign(?:ment)?|transfer|sublicense|change\s+of\s+control)\b",
    ],
    "low": [
        r"\b(confirm|acknowledge|agree|consent|accept|understand|hereby)\b",
        r"\b(shall|will|may|must|should|is\s+required|is\s+responsible)\b",
        r"\b(promptly|reasonable|reasonable\s+efforts|best\s+efforts|timely)\b",
    ]
}

KEY_TERM_PATTERNS = [
    (r"(?i)term(?:\s+is|\s+shall\s+be)?\s+(\d+\s+(?:months?|years?))", "Contract Term"),
    (r"(?i)renewal(?:\s+is|\s+shall\s+be)?\s*(.*?)(?:\.|;|$)", "Renewal Terms"),
    (r"(?i)payment(?:\s+is|\s+shall\s+be)?\s*(.*?)(?:\.|;|$)", "Payment Terms"),
    (r"(?i)liability(?:\s+is|\s+shall\s+be|\s+limited?\s+to)?\s*(.*?)(?:\.|;|$)", "Liability"),
    (r"(?i)termination(?:\s+is|\s+shall\s+be|by|\s+for)?\s*(.*?)(?:\.|;|$)", "Termination"),
    (r"(?i)indemnif(?:y|ication|ication\s+and\s+defense)\s*(.*?)(?:\.|;|$)", "Indemnification"),
    (r"(?i)confidentialit(?:y|is|ies)\s*(.*?)(?:\.|;|$)", "Confidentiality"),
    (r"(?i)intellectual\s+property\s*(.*?)(?:\.|;|$)", "IP Ownership"),
    (r"(?i)governing\s+law\s*(.*?)(?:\.|;|$)", "Governing Law"),
    (r"(?i)dispute\s+resolution\s*(.*?)(?:\.|;|$)", "Dispute Resolution"),
]


def extract_key_terms(text: str) -> list[str]:
    terms = []
    seen = set()
    
    for pattern, name in KEY_TERM_PATTERNS:
        matches = re.findall(pattern, text)
        for match in matches:
            term = f"{name}: {match.strip()}" if match else name
            term_lower = term.lower()
            if term_lower not in seen and len(term) > 5:
                seen.add(term_lower)
                terms.append(term)
    
    return terms[:10]


def identify_risks(text: str, risk_tolerance: str) -> list[dict]:
    risks = []
    
    tolerance_weights = {"low": 1.5, "medium": 1.0, "high": 0.7}
    weight = tolerance_weights.get(risk_tolerance, 1.0)
    
    text_lower = text.lower()
    
    for level, patterns in RISK_KEYWORDS.items():
        for pattern in patterns:
            matches = re.finditer(pattern, text_lower)
            for match in matches:
                start = max(0, match.start() - 30)
                end = min(len(text), match.end() + 30)
                snippet = text[start:end].strip()
                
                risk_level_numeric = {"high": 3, "medium": 2, "low": 1}[level]
                
                if level == "high" or (level == "medium" and weight < 1.2):
                    recommendation = get_recommendation(match.group(0), level)
                    risks.append({
                        "term": snippet,
                        "risk_level": level,
                        "recommendation": recommendation
                    })
    
    unique_risks = []
    seen_terms = set()
    for risk in risks:
        if risk["term"].lower() not in seen_terms:
            seen_terms.add(risk["term"].lower())
            unique_risks.append(risk)
    
    return unique_risks[:10]


def get_recommendation(term: str, level: str) -> str:
    term_lower = term.lower()
    
    if any(w in term_lower for w in ["penalty", "fine", "damages"]):
        return "Negotiate caps on penalties or convert to performance-based metrics"
    if any(w in term_lower for w in ["indemnif", "liabilit", "hold harmless"]):
        return "Ensure mutual indemnification or negotiate liability caps"
    if "unlimited" in term_lower:
        return "Negotiate specific liability limits"
    if "as-is" in term_lower or "without warranty" in term_lower:
        return "Negotiate basic warranties or representations"
    if "irrevocable" in term_lower or "perpetual" in term_lower:
        return "Ensure exit provisions or time limits are included"
    if "termination" in term_lower:
        return "Clarify termination rights and notice periods for both parties"
    if "auto-renewal" in term_lower or "automatic renewal" in term_lower:
        return "Add opt-out provisions and advance notice requirements"
    if "assignment" in term_lower or "transfer" in term_lower:
        return "Require written consent for assignment or specify conditions"
    if "price increase" in term_lower:
        return "Add caps on price increases or fixed pricing periods"
    
    return f"Review {level}-risk clause with legal counsel"


def calculate_risk_score(risks: list[dict], risk_tolerance: str) -> float:
    if not risks:
        return 0.0
    
    level_scores = {"high": 3, "medium": 2, "low": 1}
    
    total = sum(level_scores.get(r["risk_level"], 2) for r in risks)
    max_possible = len(risks) * 3
    
    base_score = (total / max_possible) * 100 if max_possible > 0 else 0
    
    tolerance_adjustment = {"low": 0.7, "medium": 1.0, "high": 1.3}.get(risk_tolerance, 1.0)
    
    adjusted_score = base_score * tolerance_adjustment
    
    return round(min(100, max(0, adjusted_score)), 1)


def summarize_contract(text: str) -> str:
    sentences = re.split(r'[.!?]+', text)
    sentences = [s.strip() for s in sentences if len(s.strip()) > 20]
    
    key_sentences = []
    important_keywords = ["agreement", "parties", "terms", "conditions", "obligations", "rights", "duties", "shall", "must", "will"]
    
    for sent in sentences[:15]:
        if any(kw in sent.lower() for kw in important_keywords):
            key_sentences.append(sent)
            if len(key_sentences) >= 3:
                break
    
    if len(key_sentences) < 2:
        key_sentences = sentences[:2] if len(sentences) >= 2 else sentences[:1]
    
    summary = ". ".join(key_sentences)
    return summary[:500] + "..." if len(summary) > 500 else summary


def handler(event: dict[str, Any]) -> dict[str, Any]:
    try:
        contract_text = event.get("contract_text", "")
        risk_tolerance = event.get("risk_tolerance", "medium").lower().strip()
        
        if not contract_text:
            return {"ok": False, "error": "contract_text is required"}
        
        if not contract_text.strip():
            return {"ok": False, "error": "contract_text cannot be empty or whitespace"}
        
        valid_tolerances = ["low", "medium", "high"]
        if risk_tolerance not in valid_tolerances:
            return {"ok": False, "error": f"risk_tolerance must be one of: {', '.join(valid_tolerances)}"}
        
        summary = summarize_contract(contract_text)
        key_terms = extract_key_terms(contract_text)
        risks = identify_risks(contract_text, risk_tolerance)
        overall_risk_score = calculate_risk_score(risks, risk_tolerance)
        
        return {
            "ok": True,
            "summary": summary,
            "key_terms": key_terms,
            "risks": risks,
            "overall_risk_score": overall_risk_score,
            "risk_tolerance": risk_tolerance
        }
        
    except Exception as e:
        return {"ok": False, "error": str(e)}
