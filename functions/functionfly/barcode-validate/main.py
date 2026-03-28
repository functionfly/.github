def validate_ean13(barcode):
    """Validate EAN-13 barcode"""
    if len(barcode) != 13:
        return False
    if not barcode.isdigit():
        return False
    # Calculate checksum
    total = 0
    for i in range(12):
        digit = int(barcode[i])
        if i % 2 == 0:
            total += digit
        else:
            total += digit * 3
    checksum = (10 - (total % 10)) % 10
    return checksum == int(barcode[12])

def validate_ean8(barcode):
    """Validate EAN-8 barcode"""
    if len(barcode) != 8:
        return False
    if not barcode.isdigit():
        return False
    # Calculate checksum
    total = 0
    for i in range(7):
        digit = int(barcode[i])
        if i % 2 == 0:
            total += digit * 3
        else:
            total += digit
    checksum = (10 - (total % 10)) % 10
    return checksum == int(barcode[7])

def validate_upc(barcode):
    """Validate UPC-A barcode"""
    if len(barcode) != 12:
        return False
    if not barcode.isdigit():
        return False
    # Calculate checksum
    total = 0
    for i in range(11):
        digit = int(barcode[i])
        if i % 2 == 0:
            total += digit * 3
        else:
            total += digit
    checksum = (10 - (total % 10)) % 10
    return checksum == int(barcode[11])

def handler(event):
    try:
        barcode = event.get("barcode", "") if isinstance(event, dict) else ""
        format_hint = event.get("format", "") if isinstance(event, dict) else ""
        if not barcode:
            return {"ok": False, "error": "barcode is required"}
        # Try to detect format
        if format_hint:
            if format_hint.upper() == "EAN-13":
                valid = validate_ean13(barcode)
                return {"ok": True, "valid": valid, "format": "EAN-13"}
            elif format_hint.upper() == "EAN-8":
                valid = validate_ean8(barcode)
                return {"ok": True, "valid": valid, "format": "EAN-8"}
            elif format_hint.upper() == "UPC":
                valid = validate_upc(barcode)
                return {"ok": True, "valid": valid, "format": "UPC-A"}
        # Auto-detect format
        if len(barcode) == 13:
            valid = validate_ean13(barcode)
            return {"ok": True, "valid": valid, "format": "EAN-13"}
        elif len(barcode) == 8:
            valid = validate_ean8(barcode)
            return {"ok": True, "valid": valid, "format": "EAN-8"}
        elif len(barcode) == 12:
            valid = validate_upc(barcode)
            return {"ok": True, "valid": valid, "format": "UPC-A"}
        else:
            return {"ok": True, "valid": False, "format": "unknown"}
    except Exception as e:
        return {"ok": False, "error": str(e)}
