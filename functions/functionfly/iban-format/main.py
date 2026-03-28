def format_iban(iban, format_type="groups"):
    """Format IBAN"""
    if not iban:
        return ""
    # Remove spaces and convert to uppercase
    iban = iban.replace(" ", "").upper()
    if format_type == "groups":
        # Format in groups of 4 characters
        groups = []
        for i in range(0, len(iban), 4):
            groups.append(iban[i:i+4])
        return " ".join(groups)
    elif format_type == "compact":
        # No spaces
        return iban
    else:
        # Default to groups
        groups = []
        for i in range(0, len(iban), 4):
            groups.append(iban[i:i+4])
        return " ".join(groups)

def handler(event):
    try:
        iban = event.get("iban", "") if isinstance(event, dict) else ""
        format_type = event.get("format", "groups") if isinstance(event, dict) else "groups"
        if not iban:
            return {"ok": False, "error": "iban is required"}
        formatted = format_iban(iban, format_type)
        return {"ok": True, "formatted": formatted}
    except Exception as e:
        return {"ok": False, "error": str(e)}
