from collections import Counter


def handler(event):
    ratings = event.get("ratings") if isinstance(event, dict) else None
    max_stars = int(event.get("max_stars", 5))
    if not ratings:
        return {"ok": False, "error": "ratings is required (list of numeric ratings)"}
    try:
        vals = [float(r) for r in ratings]
        n = len(vals)
        avg = round(sum(vals) / n, 2)
        dist = Counter(int(r) for r in vals)
        distribution = {str(star): dist.get(star, 0) for star in range(1, max_stars + 1)}
        pct_dist = {str(k): round(v / n * 100, 1) for k, v in distribution.items()}
        var = sum((v - avg) ** 2 for v in vals) / n
        std = round(var ** 0.5, 4)
        sorted_vals = sorted(vals)
        median = sorted_vals[n // 2] if n % 2 else round((sorted_vals[n // 2 - 1] + sorted_vals[n // 2]) / 2, 2)
        five_star_pct = round(dist.get(max_stars, 0) / n * 100, 1)
        return {
            "ok": True,
            "result": avg,
            "average_rating": avg,
            "total_reviews": n,
            "distribution": distribution,
            "distribution_pct": pct_dist,
            "five_star_pct": five_star_pct,
            "median_rating": median,
            "std_dev": std
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
