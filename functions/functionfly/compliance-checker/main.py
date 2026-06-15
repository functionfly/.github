import re
from typing import Any


GDPR_REQUIREMENTS = {
    "data_collection": ["consent", "legitimate_interest", "contract", "legal_obligation", "vital_interest"],
    "rights": ["access", "rectification", "erasure", "restriction", "portability", "objection"],
    "principles": ["lawfulness_fairness", "transparency", "purpose_limitation", "data_minimization", "accuracy", "storage_limitation", "integrity_confidentiality", "accountability"],
}

CCPA_REQUIREMENTS = {
    "rights": ["know_data_collected", "know_sale_shared", "opt_out_sale", "non_discrimination", "delete_data", "correct_inaccurate"],
    "disclosures": ["categories_collected", "categories_sold", "purpose", "retention"],
}

HIPAA_REQUIREMENTS = {
    "phi_protection": ["administrative_safeguards", "physical_safeguards", "technical_safeguards"],
    "rights": ["access", "amendment", "accounting", "restriction", "confidentiality"],
}

SOX_REQUIREMENTS = {
    "financial": ["accurate_financial_records", "internal_controls", "audit_trail", "documentation"],
    "compliance": ["segregation_duties", "approval_procedures", "reconciliation"],
}


def check_gdpr(data_handling: list) -> tuple[list, list, list, int]:
    violations = []
    recommendations = []
    score = 100
    
    required_practices = set()
    for category, items in GDPR_REQUIREMENTS.items():
        for item in items:
            required_practices.add(item)
    
    found_practices = set(re.sub(r'[^a-z_]', '', p.lower()) for p in data_handling)
    
    for practice in required_practices:
        if practice not in found_practices:
            violations.append(f"Missing GDPR {practice.replace('_', ' ')} requirement")
            score -= 5
    
    if "consent" not in found_practices:
        recommendations.append("Implement explicit consent mechanism for data collection")
    if "erasure" not in found_practices:
        recommendations.append("Implement data erasure/right to be forgotten procedures")
    if "portability" not in found_practices:
        recommendations.append("Implement data portability in standard machine-readable format")
    if "transparency" not in found_practices and "lawfulness_fairness" not in found_practices:
        recommendations.append("Provide clear privacy notice explaining data processing")
    if "integrity_confidentiality" not in found_practices:
        recommendations.append("Implement encryption and security measures for data")
    
    return violations, recommendations, score


def check_ccpa(data_handling: list) -> tuple[list, list, list, int]:
    violations = []
    recommendations = []
    score = 100
    
    found_practices = set(re.sub(r'[^a-z_]', '', p.lower()) for p in data_handling)
    
    ccpa_essentials = {"opt_out_sale", "know_data_collected", "delete_data", "non_discrimination"}
    for essential in ccpa_essentials:
        if essential not in found_practices:
            violations.append(f"Missing CCPA {essential.replace('_', ' ')} requirement")
            score -= 5
    
    if "opt_out_sale" not in found_practices:
        recommendations.append("Implement 'Do Not Sell My Personal Information' link on website")
    if "know_data_collected" not in found_practices:
        recommendations.append("Create data collection disclosure at point of collection")
    if "delete_data" not in found_practices:
        recommendations.append("Implement consumer request submission process for data deletion")
    
    return violations, recommendations, score


def check_hipaa(data_handling: list) -> tuple[list, list, list, int]:
    violations = []
    recommendations = []
    score = 100
    
    found_practices = set(re.sub(r'[^a-z_]', '', p.lower()) for p in data_handling)
    
    hipaa_essentials = {"access", "administrative_safeguards", "technical_safeguards", "physical_safeguards"}
    for essential in hipaa_essentials:
        if essential not in found_practices:
            violations.append(f"Missing HIPAA {essential.replace('_', ' ')} requirement")
            score -= 5
    
    if "administrative_safeguards" not in found_practices:
        recommendations.append("Implement HIPAA security officer and risk analysis procedures")
    if "technical_safeguards" not in found_practices:
        recommendations.append("Implement access controls, audit controls, and encryption")
    if "physical_safeguards" not in found_practices:
        recommendations.append("Implement facility access controls and workstation security")
    if "access" not in found_practices:
        recommendations.append("Implement procedures for patient access to PHI")
    
    return violations, recommendations, score


def check_sox(data_handling: list) -> tuple[list, list, list, int]:
    violations = []
    recommendations = []
    score = 100
    
    found_practices = set(re.sub(r'[^a-z_]', '', p.lower()) for p in data_handling)
    
    sox_essentials = {"accurate_financial_records", "internal_controls", "audit_trail", "segregation_duties"}
    for essential in sox_essentials:
        if essential not in found_practices:
            violations.append(f"Missing SOX {essential.replace('_', ' ')} requirement")
            score -= 5
    
    if "accurate_financial_records" not in found_practices:
        recommendations.append("Implement controls to ensure financial data accuracy")
    if "internal_controls" not in found_practices:
        recommendations.append("Document and test internal controls over financial reporting")
    if "audit_trail" not in found_practices:
        recommendations.append("Implement comprehensive audit logging for financial transactions")
    if "segregation_duties" not in found_practices:
        recommendations.append("Implement segregation of duties in financial processes")
    
    return violations, recommendations, score


def handler(event: dict[str, Any]) -> dict[str, Any]:
    try:
        regulation = event.get("regulation", "").lower().strip()
        data_handling_practices = event.get("data_handling_practices", [])
        
        if not regulation:
            return {"ok": False, "error": "regulation is required (gdpr/ccpa/hipaa/sox)"}
        
        if not isinstance(data_handling_practices, list):
            return {"ok": False, "error": "data_handling_practices must be a list"}
        
        valid_regulations = ["gdpr", "ccpa", "hipaa", "sox"]
        if regulation not in valid_regulations:
            return {"ok": False, "error": f"regulation must be one of: {', '.join(valid_regulations)}"}
        
        if regulation == "gdpr":
            violations, recommendations, score = check_gdpr(data_handling_practices)
        elif regulation == "ccpa":
            violations, recommendations, score = check_ccpa(data_handling_practices)
        elif regulation == "hipaa":
            violations, recommendations, score = check_hipaa(data_handling_practices)
        else:
            violations, recommendations, score = check_sox(data_handling_practices)
        
        score = max(0, min(100, score))
        is_compliant = score >= 70 and len(violations) <= 2
        
        return {
            "ok": True,
            "is_compliant": is_compliant,
            "violations": violations,
            "recommendations": recommendations,
            "compliance_score": score,
            "regulation": regulation.upper()
        }
        
    except Exception as e:
        return {"ok": False, "error": str(e)}
