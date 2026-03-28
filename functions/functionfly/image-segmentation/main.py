import hashlib

SEGMENT_CLASSES = ["background","sky","ground","building","wall","floor","tree","grass","water",
    "mountain","road","sidewalk","person","car","bicycle","motorcycle","bus","truck","boat",
    "chair","table","sofa","bed","window","door","plant","animal","food","clothing","furniture"]

def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be an object"}
    image_url = event.get("image_url", "")
    image_base64 = event.get("image_base64", "")
    mode = event.get("mode", "semantic")
    if not image_url and not image_base64:
        return {"ok": False, "error": "image_url or image_base64 is required"}
    try:
        seed = image_url or image_base64[:50]
        h = hashlib.sha256(seed.encode()).digest()
        num_segments = (h[0] % 4) + 2
        segments = []
        for i in range(num_segments):
            class_idx = (h[i % 32] + i * 7) % len(SEGMENT_CLASSES)
            area = round(0.05 + (h[(i+1) % 32] / 255.0) * 0.4, 4)
            segments.append({
                "id": i,
                "label": SEGMENT_CLASSES[class_idx],
                "area_ratio": area,
                "confidence": round(0.6 + (h[(i+2) % 32] / 255.0) * 0.4, 4),
                "bbox": {
                    "x1": round((h[(i+3) % 32] / 255.0) * 0.5, 4),
                    "y1": round((h[(i+4) % 32] / 255.0) * 0.5, 4),
                    "x2": round(0.5 + (h[(i+5) % 32] / 255.0) * 0.5, 4),
                    "y2": round(0.5 + (h[(i+6) % 32] / 255.0) * 0.5, 4)
                }
            })
        return {
            "ok": True,
            "result": segments,
            "segments": segments,
            "count": len(segments),
            "mode": mode,
            "model": "mock-segnet-v1",
            "note": "Mock segmentation — for production use, integrate DeepLab or similar"
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
