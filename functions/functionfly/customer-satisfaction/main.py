def handler(event):
    scores = event.get("scores") if isinstance(event, dict) else None
    max_score = int(event.get("max_score", 5))
    if not scores:
        return {"ok": False, "error": "scores is required"}
    try:
        vals = [float(s) for s in scores]
        n = len(vals)
        avg = round(sum(vals) / n, 4)
        csat_pct = round(avg / max_score * 100, 2)
        satisfied = sum(1 for v in vals if v >= max_score * 0.8)
        satisfied_pct = round(satisfied / n * 100, 2)
        std = round((sum((v - avg) ** 2 for v in vals) / n) ** 0.5, 4)
        if csat_pct >= 85:
            grade = "Excellent"
        elif csat_pct >= 70:
            grade = "Good"
        elif csat_pct >= 50:
            grade = "Average"
        else:
            grade = "Poor"
        return {
            "ok": True,
            "result": csat_pct,
            "csat_pct": csat_pct,
            "avg_score": avg,
            "max_score": max_score,
            "satisfied_pct": satisfied_pct,
            "total_responses": n,
            "std_dev": std,
            "grade": grade
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
