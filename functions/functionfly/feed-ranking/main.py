import math


def handler(event):
    items = event.get("items") if isinstance(event, dict) else None
    algorithm = event.get("algorithm", "edge_rank")
    if not items or not isinstance(items, list):
        return {"ok": False, "error": "items must be a non-empty list"}
    try:
        scored = []
        import time as _time
        now = _time.time()
        for item in items:
            if not isinstance(item, dict):
                continue
            likes = float(item.get("likes", 0))
            comments = float(item.get("comments", 0))
            shares = float(item.get("shares", 0))
            ts = float(item.get("created_at_unix", now))
            age_hours = max((now - ts) / 3600, 0.1)
            if algorithm == "edge_rank":
                affinity = float(item.get("affinity", 1.0))
                weight = likes * 1 + comments * 2 + shares * 3
                time_decay = 1 / math.log1p(age_hours)
                score = round(affinity * weight * time_decay, 6)
            elif algorithm == "time_decay":
                engagement = likes + comments * 2 + shares * 3
                half_life = float(event.get("half_life_hours", 24))
                score = round(engagement * math.pow(0.5, age_hours / half_life), 6)
            else:
                score = round((likes + comments + shares) / math.pow(age_hours + 2, 1.5), 6)
            scored.append({**item, "_score": score})
        ranked = sorted(scored, key=lambda x: x["_score"], reverse=True)
        return {
            "ok": True,
            "result": ranked,
            "items": ranked,
            "count": len(ranked),
            "algorithm": algorithm
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
