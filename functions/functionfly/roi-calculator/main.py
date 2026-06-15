"""ROI Calculator - Calculate Return on Investment."""


def handler(event):
    try:
        initial_investment = float(event.get("initial_investment", 0))
        final_value = float(event.get("final_value", 0))
        holding_period_days = event.get("holding_period_days")

        if initial_investment <= 0:
            return {"ok": False, "error": "initial_investment must be positive"}
        if final_value < 0:
            return {"ok": False, "error": "final_value cannot be negative"}
        if holding_period_days is not None:
            holding_period_days = int(holding_period_days)
            if holding_period_days <= 0:
                return {"ok": False, "error": "holding_period_days must be positive"}

        gain = final_value - initial_investment
        roi_percent = round((gain / initial_investment) * 100, 2) if initial_investment > 0 else 0

        result = {
            "ok": True,
            "roi_percent": roi_percent,
            "gain": round(gain, 2),
            "holding_period_years": 0.0
        }

        if holding_period_days is not None:
            holding_period_years = holding_period_days / 365.0
            result["holding_period_years"] = round(holding_period_years, 4)

            if holding_period_years > 0:
                annualized_roi = round((((final_value / initial_investment) ** (1 / holding_period_years)) - 1) * 100, 2)
                result["annualized_roi"] = annualized_roi

        return result
    except (ValueError, TypeError) as e:
        return {"ok": False, "error": f"Invalid input: {str(e)}"}
    except Exception as e:
        return {"ok": False, "error": f"Internal error: {str(e)}"}
