import math


def handler(event):
    score = event.get("score", 0) if isinstance(event, dict) else 0
    upvotes = event.get("upvotes", 0)
    downvotes = event.get("downvotes", 0)
    comments = event.get("comments", 0)
    created_at_unix = event.get("created_at_unix")
    algorithm = event.get("algorithm", "hacker_news")
    try:
        u, d, c = float(upvotes), float(downvotes), float(comments)
        if algorithm == "reddit":
            net = u - d
            sign = 1 if net > 0 else (-1 if net < 0 else 0)
            order = math.log10(max(abs(net), 1))
            ts = created_at_unix or 0
            epoch = 1134028003
            seconds = ts - epoch
            result = round(sign * order + seconds / 45000, 7)
        elif algorithm == "wilson":
            n = u + d
            if n == 0:
                result = 0.0
            else:
                z = 1.96
                phat = u / n
                result = round((phat + z*z/(2*n) - z*math.sqrt((phat*(1-phat)+z*z/(4*n))/n)) / (1+z*z/n), 6)
        else:
            # Hacker News: gravity model
            gravity = 1.8
            ts = created_at_unix or 0
            import time as _time
            age_hours = max((_time.time() - ts) / 3600, 0.1) if ts else 1.0
            points = max(float(score) + u - d, 0)
            result = round(points / math.pow(age_hours + 2, gravity), 6)
        category = "rising" if result > 0 else ("stable" if result == 0 else "falling")
        return {
            "ok": True,
            "result": result,
            "trending_score": result,
            "algorithm": algorithm,
            "category": category
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
