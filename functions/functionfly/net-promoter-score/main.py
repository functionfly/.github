from collections import Counter


def handler(event):
    scores = event.get("scores") if isinstance(event, dict) else None
    if not scores:
        return {"ok": False, "error": "scores is required (list of 0-10 ratings)"}
    try:
        vals = [int(s) for s in scores]
        invalid = [v for v in vals if not 0 <= v <= 10]
        if invalid:
            return {"ok": False, "error": f"All scores must be 0-10. Invalid: {invalid[:5]}"}
        n = len(vals)
        promoters = sum(1 for v in vals if v >= 9)
        passives = sum(1 for v in vals if 7 <= v <= 8)
        detractors = sum(1 for v in vals if v <= 6)
        nps = round((promoters - detractors) / n * 100, 1)
        dist = Counter(vals)
        return {
            "ok": True,
            "result": nps,
            "nps": nps,
            "promoters": promoters,
            "passives": passives,
            "detractors": detractors,
            "promoter_pct": round(promoters / n * 100, 1),
            "passive_pct": round(passives / n * 100, 1),
            "detractor_pct": round(detractors / n * 100, 1),
            "total_responses": n,
            "grade": "Excellent" if nps >= 70 else ("Great" if nps >= 50 else ("Good" if nps >= 30 else ("Needs improvement" if nps >= 0 else "Poor")))
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
