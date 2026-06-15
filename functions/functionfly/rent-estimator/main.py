"""Rent Estimator - Estimate rent based on property features."""


def handler(event):
    try:
        bedrooms = int(event.get("bedrooms", 0))
        bathrooms = float(event.get("bathrooms", 0))
        sqft = int(event.get("sqft", 0))
        city = event.get("city", "")
        state = event.get("state", "")

        if bedrooms < 0 or bedrooms > 20:
            return {"ok": False, "error": "bedrooms must be between 0 and 20"}
        if bathrooms < 0 or bathrooms > 20:
            return {"ok": False, "error": "bathrooms must be between 0 and 20"}
        if sqft <= 0 or sqft > 50000:
            return {"ok": False, "error": "sqft must be between 1 and 50000"}
        if not city or not isinstance(city, str):
            return {"ok": False, "error": "city is required"}
        if not state or not isinstance(state, str) or len(state) != 2:
            return {"ok": False, "error": "state must be a 2-letter code"}

        base = 1000.0
        bedroom_rate = 300.0
        bathroom_rate = 150.0
        sqft_rate = 0.50

        estimated_rent = base + (bedrooms * bedroom_rate) + (bathrooms * bathroom_rate) + (sqft * sqft_rate)
        estimated_rent = round(estimated_rent, 2)
        price_per_sqft = round(estimated_rent / sqft, 2) if sqft > 0 else 0

        range_min = round(estimated_rent * 0.85, 2)
        range_max = round(estimated_rent * 1.15, 2)

        return {
            "ok": True,
            "estimated_rent": estimated_rent,
            "price_per_sqft": price_per_sqft,
            "range_min": range_min,
            "range_max": range_max
        }
    except (ValueError, TypeError) as e:
        return {"ok": False, "error": f"Invalid input: {str(e)}"}
    except Exception as e:
        return {"ok": False, "error": f"Internal error: {str(e)}"}
