"""Product Pricing Optimizer - Optimize product pricing strategy."""
import statistics
from datetime import datetime
from typing import Any


def handler(event: dict) -> dict:
    """Optimize product pricing."""
    try:
        product_name = event.get("product_name")
        base_cost = event.get("base_cost")
        competitor_prices = event.get("competitor_prices", [])
        target_margin_percent = event.get("target_margin_percent", 30)

        if not product_name:
            return {"ok": False, "error": "product_name is required"}
        if base_cost is None:
            return {"ok": False, "error": "base_cost is required"}
        if not competitor_prices or len(competitor_prices) == 0:
            return {"ok": False, "error": "competitor_prices list is required and must not be empty"}

        try:
            base_cost = float(base_cost)
            if base_cost <= 0:
                return {"ok": False, "error": "base_cost must be a positive number"}
        except (ValueError, TypeError):
            return {"ok": False, "error": "base_cost must be a valid number"}

        try:
            competitor_prices = [float(p) for p in competitor_prices]
            for p in competitor_prices:
                if p <= 0:
                    return {"ok": False, "error": "all competitor_prices must be positive numbers"}
        except (ValueError, TypeError):
            return {"ok": False, "error": "competitor_prices must be a list of valid numbers"}

        try:
            target_margin_percent = float(target_margin_percent)
            if target_margin_percent < 0 or target_margin_percent > 99:
                return {"ok": False, "error": "target_margin_percent must be between 0 and 99"}
        except (ValueError, TypeError):
            return {"ok": False, "error": "target_margin_percent must be a valid number"}

        min_competitor = min(competitor_prices)
        max_competitor = max(competitor_prices)
        avg_competitor = statistics.mean(competitor_prices)
        median_competitor = statistics.median(competitor_prices)

        target_price_by_margin = base_cost / (1 - (target_margin_percent / 100))

        price_range_min = base_cost * 1.15
        price_range_max = max_competitor * 1.1

        if target_price_by_margin < price_range_min:
            recommended_price = price_range_min
        elif target_price_by_margin > price_range_max:
            recommended_price = price_range_max
        else:
            recommended_price = target_price_by_margin

        recommended_rounded = round(recommended_price, 2)

        margin_at_recommended = ((recommended_rounded - base_cost) / recommended_rounded) * 100

        competitor_comparison = []
        for i, cp in enumerate(competitor_prices):
            diff = recommended_rounded - cp
            pct_diff = (diff / cp) * 100 if cp > 0 else 0
            competitor_comparison.append({
                "competitor": f"Competitor {i + 1}",
                "price": cp,
                "difference": round(diff, 2),
                "percentage_difference": round(pct_diff, 2),
                "position": "above" if diff > 0 else "below"
            })

        pricing_strategy = "premium" if recommended_rounded > avg_competitor * 1.05 else \
                          "competitive" if recommended_rounded < avg_competitor * 0.95 else \
                          "market matching"

        return {
            "ok": True,
            "product_name": product_name,
            "base_cost": base_cost,
            "recommended_price": recommended_rounded,
            "price_range": {
                "min": round(price_range_min, 2),
                "max": round(price_range_max, 2)
            },
            "target_margin_percent": target_margin_percent,
            "margin_at_recommended_price": round(margin_at_recommended, 2),
            "competitor_comparison": competitor_comparison,
            "market_analysis": {
                "min_competitor_price": min_competitor,
                "max_competitor_price": max_competitor,
                "average_competitor_price": round(avg_competitor, 2),
                "median_competitor_price": round(median_competitor, 2)
            },
            "pricing_strategy": pricing_strategy,
            "generated_at": datetime.now().isoformat()
        }

    except Exception as e:
        return {"ok": False, "error": f"Failed to optimize pricing: {str(e)}"}


if __name__ == "__main__":
    result = handler({
        "product_name": "Premium Wireless Headphones",
        "base_cost": 45.00,
        "competitor_prices": [99.99, 129.99, 89.99, 149.99, 109.99],
        "target_margin_percent": 40
    })
    print(result)
