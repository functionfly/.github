"""Product Recommender - Recommend products based on user preferences."""
from datetime import datetime
from typing import Any


def calculate_match_score(product: dict, preferences: dict) -> tuple:
    """Calculate how well a product matches user preferences."""
    score = 0.0
    max_score = 100.0
    reasons = []

    if "category" in preferences and product.get("category"):
        if preferences["category"].lower() == product["category"].lower():
            score += 30
            reasons.append(f"Matches preferred category: {product['category']}")
        elif preferences["category"].lower() in product.get("category", "").lower():
            score += 15

    if "price_range" in preferences and "price" in product:
        min_price, max_price = preferences["price_range"]
        price = float(product["price"])
        if min_price <= price <= max_price:
            score += 25
            reasons.append(f"Within budget range (${min_price}-${max_price})")
        elif price < min_price:
            score += 10
            reasons.append("Below budget (value option)")
        else:
            overlap = (max_price / price) * 20 if price > 0 else 0
            score += max(0, overlap)
            reasons.append("Above budget but may be worth it")

    if "features" in preferences and product.get("features"):
        product_features = set(f.lower() for f in product.get("features", []))
        user_features = set(f.lower() for f in preferences["features"])
        matching = product_features & user_features
        if matching:
            feature_score = (len(matching) / len(user_features)) * 25
            score += feature_score
            reasons.append(f"Matches {len(matching)} desired features")

    if "rating" in preferences and product.get("rating"):
        if product["rating"] >= preferences["rating"]:
            score += 20
            reasons.append(f"High rating: {product['rating']} stars")

    score = min(score, max_score)
    return round(score, 2), reasons


def handler(event: dict) -> dict:
    """Recommend products based on user preferences."""
    try:
        user_preferences = event.get("user_preferences", {})
        available_products = event.get("available_products", [])

        if not user_preferences:
            return {"ok": False, "error": "user_preferences is required"}
        if not available_products or len(available_products) == 0:
            return {"ok": False, "error": "available_products list is required and must not be empty"}

        if not isinstance(user_preferences, dict):
            return {"ok": False, "error": "user_preferences must be a dictionary"}

        if not isinstance(available_products, list):
            return {"ok": False, "error": "available_products must be a list"}

        for i, p in enumerate(available_products):
            if not isinstance(p, dict):
                return {"ok": False, "error": f"product at index {i} must be an object"}

        scored_products = []
        for product in available_products:
            score, reasons = calculate_match_score(product, user_preferences)
            scored_products.append({
                "product": product,
                "match_score": score,
                "match_reasons": reasons
            })

        scored_products.sort(key=lambda x: x["match_score"], reverse=True)

        top_5 = scored_products[:5]

        recommendations = []
        for item in top_5:
            rec = item["product"].copy()
            rec["match_score"] = item["match_score"]
            rec["match_reasons"] = item["match_reasons"]
            recommendations.append(rec)

        explanation_parts = []
        if "category" in user_preferences:
            explanation_parts.append(f"Category: {user_preferences['category']}")
        if "price_range" in user_preferences:
            pr = user_preferences["price_range"]
            explanation_parts.append(f"Price range: ${pr[0]}-${pr[1]}")
        if "features" in user_preferences:
            explanation_parts.append(f"Desired features: {', '.join(user_preferences['features'][:3])}")

        explanation = f"Based on your preferences ({', '.join(explanation_parts)}), these products were ranked by relevance score."

        return {
            "ok": True,
            "user_preferences": user_preferences,
            "recommendations": recommendations,
            "explanation": explanation,
            "total_products_analyzed": len(available_products),
            "generated_at": datetime.now().isoformat()
        }

    except Exception as e:
        return {"ok": False, "error": f"Failed to generate recommendations: {str(e)}"}


if __name__ == "__main__":
    result = handler({
        "user_preferences": {
            "category": "Electronics",
            "price_range": [50, 200],
            "features": ["wireless", "bluetooth", "long battery life"],
            "rating": 4.0
        },
        "available_products": [
            {"id": 1, "name": "Wireless Headphones", "category": "Electronics", "price": 79.99, "features": ["wireless", "bluetooth", "noise-canceling"], "rating": 4.5},
            {"id": 2, "name": "Bluetooth Speaker", "category": "Electronics", "price": 49.99, "features": ["bluetooth", "portable"], "rating": 4.2},
            {"id": 3, "name": "Smart Watch", "category": "Electronics", "price": 199.99, "features": ["wireless charging", "fitness tracking", "bluetooth"], "rating": 4.7},
            {"id": 4, "name": "Gaming Mouse", "category": "Electronics", "price": 59.99, "features": ["wireless", "ergonomic"], "rating": 4.3},
            {"id": 5, "name": "USB Hub", "category": "Electronics", "price": 29.99, "features": ["usb-c", "portable"], "rating": 4.1},
        ]
    })
    print(result)
