def generate_qrcode_text(data, size=21):
    """Generate a simple text-based QR code representation"""
    if not data:
        return ""
    # Simple text representation - just show data in a box
    width = max(len(data) + 4, size)
    height = max(3, size // 3)
    lines = []
    lines.append("█" * width)
    for i in range(height - 2):
        if i == height // 2 - 1:
            padding = (width - len(data) - 2) // 2
            lines.append("█" + " " * padding + data + " " * (width - len(data) - padding - 2) + "█")
        else:
            lines.append("█" + " " * (width - 2) + "█")
    lines.append("█" * width)
    return "\n".join(lines)

def handler(event):
    try:
        data = event.get("data", "") if isinstance(event, dict) else ""
        size = event.get("size", 21) if isinstance(event, dict) else 21
        if not data:
            return {"ok": False, "error": "data is required"}
        if size < 1:
            return {"ok": False, "error": "size must be at least 1"}
        qrcode = generate_qrcode_text(data, size)
        return {"ok": True, "qrcode": qrcode}
    except Exception as e:
        return {"ok": False, "error": str(e)}
