import hashlib

def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be an object"}
    image_url = event.get("image_url", "")
    image_base64 = event.get("image_base64", "")
    min_confidence = float(event.get("min_confidence", 0.8))
    if not image_url and not image_base64:
        return {"ok": False, "error": "image_url or image_base64 is required"}
    try:
        seed = image_url or image_base64[:50]
        h = hashlib.sha256(seed.encode()).digest()
        num_faces = (h[0] % 3) + 1
        faces = []
        for i in range(num_faces):
            confidence = round(0.8 + (h[i % 32] / 255.0) * 0.2, 4)
            if confidence < min_confidence:
                continue
            x = round((h[(i+1) % 32] / 255.0) * 0.6, 4)
            y = round((h[(i+2) % 32] / 255.0) * 0.6, 4)
            size = round(0.1 + (h[(i+3) % 32] / 255.0) * 0.2, 4)
            faces.append({
                "id": i,
                "confidence": confidence,
                "bbox": {"x": x, "y": y, "width": size, "height": size},
                "landmarks": {
                    "left_eye": [round(x + size * 0.3, 4), round(y + size * 0.35, 4)],
                    "right_eye": [round(x + size * 0.7, 4), round(y + size * 0.35, 4)],
                    "nose": [round(x + size * 0.5, 4), round(y + size * 0.55, 4)],
                    "left_mouth": [round(x + size * 0.35, 4), round(y + size * 0.75, 4)],
                    "right_mouth": [round(x + size * 0.65, 4), round(y + size * 0.75, 4)]
                },
                "attributes": {
                    "age_estimate": (h[(i+4) % 32] % 50) + 18,
                    "gender": "male" if h[(i+5) % 32] > 127 else "female",
                    "smile": h[(i+6) % 32] > 127
                }
            })
        return {
            "ok": True,
            "result": faces,
            "faces": faces,
            "count": len(faces),
            "model": "mock-retinaface-v1",
            "note": "Mock face detection — for production use, integrate RetinaFace or similar"
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
