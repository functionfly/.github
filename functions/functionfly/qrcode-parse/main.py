def parse_qrcode_text(qrcode):
    """Parse a simple text-based QR code representation"""
    if not qrcode:
        return ""
    lines = qrcode.split('\n')
    # Find the middle line with data
    for line in lines:
        if ' ' in line and '█' in line:
            # Extract text between borders
            parts = line.split('█')
            for part in parts:
                part = part.strip()
                if part:
                    return part
    return ""

def handler(event):
    try:
        qrcode = event.get("qrcode", "") if isinstance(event, dict) else ""
        if not qrcode:
            return {"ok": False, "error": "qrcode is required"}
        data = parse_qrcode_text(qrcode)
        return {"ok": True, "data": data}
    except Exception as e:
        return {"ok": False, "error": str(e)}
