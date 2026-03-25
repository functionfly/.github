import math


def handler(event):
    helpful_votes = event.get("helpful_votes") if isinstance(event, dict) else None
    total_votes = event.get("total_votes")
    review_age_days = float(event.get("review_age_days", 1))
    review_length = int(event.get("review_length", 0))
    has_images = event.get("has_images", False)
    verified_purchase = event.get("verified_purchase", False)
    if helpful_votes is None or total_votes is None:
        return {"ok": False, "error": "helpful_votes and total_votes are required"}
    try:
        hv, tv = int(helpful_votes), int(total_votes)
        if tv == 0:
            raw_ratio = 0.5
        else:
            raw_ratio = hv / tv
        # Wilson score lower bound for helpfulness confidence
        z = 1.96  # 95% confidence
        if tv > 0:
            p_hat = raw_ratio
            wilson = (p_hat + z**2/(2*tv) - z * math.sqrt((p_hat*(1-p_hat) + z**2/(4*tv))/tv)) / (1 + z**2/tv)
        else:
            wilson = 0
        # Content quality bonuses
        length_score = min(1.0, review_length / 500) * 0.2
        image_score = 0.1 if has_images else 0
        verified_score = 0.1 if verified_purchase else 0
        # Recency decay
        recency = max(0.5, 1 - review_age_days / 3650)
        helpfulness_score = round((wilson * 0.6 + length_score + image_score + verified_score) * recency, 4)
        return {
            "ok": True,
            "result": helpfulness_score,
            "helpfulness_score": helpfulness_score,
            "wilson_lower_bound": round(wilson, 4),
            "helpful_ratio": round(raw_ratio, 4),
            "helpful_votes": hv,
            "total_votes": tv,
            "grade": "very_helpful" if helpfulness_score > 0.6 else ("helpful" if helpfulness_score > 0.3 else "not_helpful")
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
