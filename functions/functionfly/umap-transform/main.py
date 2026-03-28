import hashlib
import math

def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be an object"}
    data = event.get("data")
    if not data or not isinstance(data, list):
        return {"ok": False, "error": "data (array of arrays) is required"}
    try:
        n_components = int(event.get("n_components", 2))
        n_neighbors = int(event.get("n_neighbors", 15))
        min_dist = float(event.get("min_dist", 0.1))
        processed = [[float(x) for x in row] for row in data if isinstance(row, list)]
        if len(processed) < 2:
            return {"ok": False, "error": "data must contain at least 2 samples"}
        # Mock UMAP: use hash-based deterministic projection
        transformed = []
        for i, row in enumerate(processed):
            seed = str(row[:4]) + str(i) + str(n_neighbors)
            h = hashlib.sha256(seed.encode()).digest()
            point = []
            for j in range(n_components):
                val = (h[j % 32] / 127.5 - 1.0) * 5
                point.append(round(val, 4))
            transformed.append(point)
        return {
            "ok": True,
            "result": transformed,
            "transformed": transformed,
            "n_components": n_components,
            "n_neighbors": n_neighbors,
            "min_dist": min_dist,
            "n_samples": len(processed),
            "note": "Mock UMAP — for production use, integrate umap-learn"
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
