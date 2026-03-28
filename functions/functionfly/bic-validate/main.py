def validate_bic(bic):
    """Validate BIC/SWIFT code"""
    if not bic:
        return False, ""
    # Remove spaces and convert to uppercase
    bic = bic.replace(" ", "").upper()
    # Check length (8 or 11 characters)
    if len(bic) not in [8, 11]:
        return False, ""
    # Check format: 4 letters (bank code) + 2 letters (country code) + 2 alphanumeric (location code) + optional 3 alphanumeric (branch code)
    bank_code = bic[:4]
    country_code = bic[4:6]
    location_code = bic[6:8]
    branch_code = bic[8:] if len(bic) == 11 else ""
    # Validate bank code (4 letters)
    if not bank_code.isalpha():
        return False, ""
    # Validate country code (2 letters)
    if not country_code.isalpha():
        return False, ""
    # Validate location code (2 alphanumeric)
    if not location_code.isalnum():
        return False, ""
    # Validate branch code (3 alphanumeric if present)
    if branch_code and not branch_code.isalnum():
        return False, ""
    return True, country_code

def handler(event):
    try:
        bic = event.get("bic", "") if isinstance(event, dict) else ""
        if not bic:
            return {"ok": False, "error": "bic is required"}
        valid, country = validate_bic(bic)
        return {"ok": True, "valid": valid, "country": country}
    except Exception as e:
        return {"ok": False, "error": str(e)}
