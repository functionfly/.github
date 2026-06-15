"""Keyword Rank Tracker - Track keyword ranking changes."""
import random


def calculate_traffic_estimate(rank, search_volume):
    if rank <= 0:
        return 0

    if rank == 1:
        click_rate = 0.32
    elif rank == 2:
        click_rate = 0.20
    elif rank == 3:
        click_rate = 0.14
    elif rank <= 5:
        click_rate = 0.08
    elif rank <= 10:
        click_rate = 0.04
    elif rank <= 20:
        click_rate = 0.02
    elif rank <= 50:
        click_rate = 0.01
    else:
        click_rate = 0.005

    estimated_traffic = int(search_volume * click_rate)
    return estimated_traffic


def calculate_opportunity_score(rank):
    if rank <= 0:
        return 0

    if rank <= 3:
        return random.randint(70, 85)
    elif rank <= 10:
        return random.randint(80, 95)
    elif rank <= 20:
        return random.randint(60, 80)
    elif rank <= 50:
        return random.randint(40, 60)
    else:
        return random.randint(20, 40)


def handler(event):
    try:
        keyword = event.get("keyword", "")
        url = event.get("url", "")
        current_rank = event.get("current_rank")
        previous_rank = event.get("previous_rank")

        if not keyword:
            return {"ok": False, "error": "keyword is required"}
        if current_rank is None:
            return {"ok": False, "error": "current_rank is required"}
        if not isinstance(current_rank, int) or current_rank < 1:
            return {"ok": False, "error": "current_rank must be a positive integer"}

        search_volume = event.get("search_volume", random.randint(500, 10000))

        if previous_rank is None:
            previous_rank = current_rank + random.randint(-5, 5)
            previous_rank = max(1, previous_rank)

        rank_change = previous_rank - current_rank

        if rank_change > 0:
            rank_direction = "up"
        elif rank_change < 0:
            rank_direction = "down"
        else:
            rank_direction = "stable"

        estimated_traffic = calculate_traffic_estimate(current_rank, search_volume)

        opportunity_score = calculate_opportunity_score(current_rank)

        return {
            "ok": True,
            "rank_change": rank_change,
            "rank_direction": rank_direction,
            "estimated_traffic": estimated_traffic,
            "opportunity_score": opportunity_score,
            "keyword": keyword,
            "url": url,
            "current_rank": current_rank,
            "previous_rank": previous_rank,
            "search_volume": search_volume
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
