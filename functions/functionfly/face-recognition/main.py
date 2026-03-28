import hashlib

def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be an object"}
    image_url = event.get("image_url", "")
    image_base64 = event.get("image_base64", "")
    known_faces = event.get("known_faces", [])
    if not image_url and not image_base64:
        return {"ok": False, "error": "image_url or image_base64 is required"}
    try:
        seed = image_url or image_base64[:50]
        h = hashlib.sha256(seed.encode()).digest()
        # Generate face embedding
        embedding = [round((h[i % 32] / 127.5 - 1.0), 6) for i in range(128)]
        # Match against known faces
        matches = []
        if known_faces:
            for face in known_faces:
                name = face.get("name", "Unknown") if isinstance(face, dict) else str(face)
                similarity = round(0.5 + (hashlib.md5(f"{seed}:{name}".encode()).digest()[0] / 255.0) * 0.5, 4)
                matches.append({"name": name, "similarity": similarity, "match": similarity > 0.8})
            matches.sort(key=lambda x: x["similarity"], reverse=True)
        return {
            "ok": True,
            "result": {"embedding": embedding, "matches": matches},
            "embedding": embedding,
            "matches": matches,
            "best_match": matches[0] if matches else None,
            "model": "mock-arcface-v1",
            "note": "Mock face recognition — for production use, integrate ArcFace or similar"
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
