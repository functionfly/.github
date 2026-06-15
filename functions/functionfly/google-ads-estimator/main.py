import math
from typing import Any


COMPETITION_MULTIPLIERS = {
    "low": 0.65,
    "medium": 1.0,
    "high": 1.5,
}

BID_RANGE_MIN = {
    "low": 0.20,
    "medium": 0.50,
    "high": 1.00,
}

BID_RANGE_MAX = {
    "low": 1.50,
    "medium": 3.50,
    "high": 8.00,
}

SEARCH_VOLUME_CPC_ESTIMATES = [
    (10000000, 2.5),
    (5000000, 2.0),
    (1000000, 1.75),
    (500000, 1.5),
    (100000, 1.25),
    (50000, 1.0),
    (10000, 0.85),
    (5000, 0.70),
    (1000, 0.55),
    (500, 0.45),
    (100, 0.35),
    (50, 0.30),
    (10, 0.25),
    (0, 0.20),
]

DAILY_BUDGET_MULTIPLIER = 0.05


def estimate_cpc(monthly_searches: int, competition: str, top_bid: float = None) -> float:
    if top_bid is not None and top_bid > 0:
        base_cpc = top_bid * 0.75
    else:
        base_cpc = 0.50
        
        for threshold, cpc in SEARCH_VOLUME_CPC_ESTIMATES:
            if monthly_searches >= threshold:
                base_cpc = cpc
                break
    
    comp_mult = COMPETITION_MULTIPLIERS.get(competition.lower(), 1.0)
    
    estimated_cpc = base_cpc * comp_mult
    
    estimated_cpc = max(
        BID_RANGE_MIN.get(competition.lower(), 0.20),
        min(BID_RANGE_MAX.get(competition.lower(), 8.00), estimated_cpc)
    )
    
    return round(estimated_cpc, 2)


def estimate_daily_budget(cpc: float, monthly_searches: int) -> float:
    clicks_per_month = monthly_searches * 0.10
    
    estimated_monthly = cpc * clicks_per_month
    
    daily_budget = estimated_monthly * DAILY_BUDGET_MULTIPLIER
    
    daily_budget = max(5.0, min(1000.0, daily_budget))
    
    return round(daily_budget, 2)


def estimate_ad_rank(cpc: float, competition: str) -> str:
    comp_level = COMPETITION_MULTIPLIERS.get(competition.lower(), 1.0)
    
    quality_factor = 0.7 + (0.3 * (cpc / BID_RANGE_MAX.get(competition.lower(), 8.00)))
    
    rank_score = cpc * quality_factor * comp_level
    
    if rank_score > 15:
        return "Top (1-4)"
    elif rank_score > 8:
        return "Upper Middle (5-8)"
    elif rank_score > 4:
        return "Middle (9-12)"
    else:
        return "Lower (13+)"


def estimate_impressions(cpc: float, monthly_searches: int, competition: str) -> int:
    ctr_base = 0.05
    
    if cpc > 3:
        ctr_base = 0.08
    elif cpc > 1.5:
        ctr_base = 0.06
    
    comp_mult = COMPETITION_MULTIPLIERS.get(competition.lower(), 1.0)
    
    if comp_mult > 1.2:
        ctr_base *= 0.9
    
    daily_searches = monthly_searches / 30
    
    daily_impressions = daily_searches * 0.3
    
    return int(daily_impressions * 30)


def estimate_clicks(monthly_searches: int, cpc: float, competition: str) -> int:
    ctr_base = 0.05
    
    if cpc > 3:
        ctr_base = 0.08
    elif cpc > 1.5:
        ctr_base = 0.06
    
    comp_mult = COMPETITION_MULTIPLIERS.get(competition.lower(), 1.0)
    
    if comp_mult > 1.2:
        ctr_base *= 0.9
    
    monthly_clicks = monthly_searches * ctr_base
    
    return int(monthly_clicks)


def handler(event: dict[str, Any]) -> dict[str, Any]:
    try:
        keyword = event.get("keyword", "")
        monthly_searches = event.get("monthly_searches", 0)
        competition = event.get("competition", "medium").lower().strip()
        top_bid = event.get("top_bid")
        
        if not keyword or not isinstance(keyword, str):
            return {"ok": False, "error": "keyword is required and must be a string"}
        
        try:
            monthly_searches = int(monthly_searches)
        except (ValueError, TypeError):
            return {"ok": False, "error": "monthly_searches must be an integer"}
        
        if monthly_searches < 0:
            return {"ok": False, "error": "monthly_searches cannot be negative"}
        
        valid_competition = ["low", "medium", "high"]
        if competition not in valid_competition:
            return {"ok": False, "error": f"competition must be one of: {', '.join(valid_competition)}"}
        
        if top_bid is not None:
            try:
                top_bid = float(top_bid)
            except (ValueError, TypeError):
                return {"ok": False, "error": "top_bid must be a number"}
            
            if top_bid < 0:
                return {"ok": False, "error": "top_bid cannot be negative"}
        
        estimated_cpc = estimate_cpc(monthly_searches, competition, top_bid)
        daily_budget = estimate_daily_budget(estimated_cpc, monthly_searches)
        ad_rank_estimate = estimate_ad_rank(estimated_cpc, competition)
        impressions_estimate = estimate_impressions(estimated_cpc, monthly_searches, competition)
        clicks_estimate = estimate_clicks(monthly_searches, estimated_cpc, competition)
        
        return {
            "ok": True,
            "estimated_cpc": estimated_cpc,
            "daily_budget": daily_budget,
            "ad_rank_estimate": ad_rank_estimate,
            "impressions_estimate": impressions_estimate,
            "clicks_estimate": clicks_estimate,
            "keyword": keyword,
            "monthly_searches": monthly_searches,
            "competition": competition
        }
        
    except Exception as e:
        return {"ok": False, "error": str(e)}
