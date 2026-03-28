def validate_iban(iban):
    """Validate IBAN (International Bank Account Number)"""
    if not iban:
        return False, ""
    # Remove spaces and convert to uppercase
    iban = iban.replace(" ", "").upper()
    # Check minimum length
    if len(iban) < 4:
        return False, ""
    # Check country code
    country = iban[:2]
    if not country.isalpha():
        return False, ""
    # Check length
    check_digits = iban[2:4]
    if not check_digits.isdigit():
        return False, ""
    # Move first 4 characters to end
    rearranged = iban[4:] + iban[:4]
    # Convert letters to numbers
    numeric = ""
    for char in rearranged:
        if char.isdigit():
            numeric += char
        elif char.isalpha():
            numeric += str(ord(char) - ord('A') + 10)
        else:
            return False, ""
    # Check mod 97
    if int(numeric) % 97 == 1:
        return True, country
    return False, country

def handler(event):
    try:
        iban = event.get("iban", "") if isinstance(event, dict) else ""
        if not iban:
            return {"ok": False, "error": "iban is required"}
        valid, country = validate_iban(iban)
        return {"ok": True, "valid": valid, "country": country}
    except Exception as e:
        return {"ok": False, "error": str(e)}
