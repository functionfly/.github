"""Sales Forecast - Generate sales forecasts using simple statistical methods."""
import statistics
from datetime import datetime
from typing import Any


def calculate_moving_average(data: list, window: int = 3) -> float:
    """Calculate simple moving average."""
    if len(data) < window:
        return statistics.mean(data) if data else 0
    return statistics.mean(data[-window:])


def calculate_linear_regression(data: list) -> tuple:
    """Calculate linear regression (slope and intercept)."""
    if len(data) < 2:
        return 0, statistics.mean(data) if data else 0

    n = len(data)
    x = list(range(n))
    y = data

    x_mean = statistics.mean(x)
    y_mean = statistics.mean(y)

    numerator = sum((x[i] - x_mean) * (y[i] - y_mean) for i in range(n))
    denominator = sum((x[i] - x_mean) ** 2 for i in range(n))

    if denominator == 0:
        return 0, y_mean

    slope = numerator / denominator
    intercept = y_mean - slope * x_mean

    return slope, intercept


def detect_seasonality(data: list) -> dict:
    """Detect simple seasonality patterns."""
    if len(data) < 4:
        return {"has_seasonality": False, "pattern": None}

    n = len(data)
    quarter_means = []

    for i in range(0, min(4, n)):
        quarter_data = data[i::4]
        if quarter_data:
            quarter_means.append(statistics.mean(quarter_data))

    if len(quarter_means) >= 2:
        variance = statistics.variance(quarter_means) if len(quarter_means) > 1 else 0
        mean_val = statistics.mean(quarter_means)
        cv = (variance ** 0.5 / mean_val) if mean_val > 0 else 0

        return {
            "has_seasonality": cv > 0.15,
            "pattern": "quarterly" if len(quarter_means) >= 2 else None,
            "quarterly_indices": quarter_means
        }

    return {"has_seasonality": False, "pattern": None}


def handler(event: dict) -> dict:
    """Generate sales forecast."""
    try:
        historical_sales = event.get("historical_sales", [])
        future_months = event.get("future_months", 6)
        seasonality = event.get("seasonality", False)

        if not historical_sales or len(historical_sales) == 0:
            return {"ok": False, "error": "historical_sales list is required and must not be empty"}

        if not isinstance(historical_sales, list):
            return {"ok": False, "error": "historical_sales must be a list"}

        try:
            historical_sales = [float(x) for x in historical_sales]
        except (ValueError, TypeError):
            return {"ok": False, "error": "historical_sales must contain valid numbers"}

        for val in historical_sales:
            if val < 0:
                return {"ok": False, "error": "historical_sales values must be non-negative"}

        if not isinstance(future_months, int) or future_months < 1 or future_months > 24:
            return {"ok": False, "error": "future_months must be an integer between 1 and 24"}

        base_month = datetime.now().month
        base_year = datetime.now().year

        ma_3 = calculate_moving_average(historical_sales, 3)
        ma_6 = calculate_moving_average(historical_sales, 6)

        slope, intercept = calculate_linear_regression(historical_sales)

        seasonality_info = detect_seasonality(historical_sales)

        forecast = []
        current_values = list(historical_sales)

        for i in range(future_months):
            month_num = (base_month + i) % 12
            if month_num == 0:
                month_num = 12

            month_names = ["January", "February", "March", "April", "May", "June",
                          "July", "August", "September", "October", "November", "December"]
            month_name = month_names[month_num - 1]

            future_index = len(historical_sales) + i

            if seasonality and seasonality_info.get("has_seasonality"):
                quarter_idx = (month_num - 1) // 3
                if quarter_idx < len(seasonality_info.get("quarterly_indices", [])):
                    seasonal_factor = seasonality_info["quarterly_indices"][quarter_idx]
                    base_avg = statistics.mean(historical_sales)
                    seasonal_adjustment = (seasonal_factor / base_avg) if base_avg > 0 else 1.0
                else:
                    seasonal_adjustment = 1.0
            else:
                seasonal_adjustment = 1.0

            linear_prediction = slope * future_index + intercept

            ma_blend = (ma_3 + ma_6) / 2
            predicted = (linear_prediction * 0.6 + ma_blend * 0.4) * seasonal_adjustment
            predicted = max(0, predicted)

            predicted_rounded = round(predicted, 2)

            year_offset = (base_month + i - 1) // 12
            display_year = base_year + year_offset

            forecast.append({
                "month": month_name,
                "year": display_year,
                "month_index": i + 1,
                "predicted_sales": predicted_rounded
            })

            current_values.append(predicted)

        recent_trend = slope
        if recent_trend > statistics.mean(historical_sales) * 0.05:
            trend_direction = "increasing"
        elif recent_trend < -statistics.mean(historical_sales) * 0.05:
            trend_direction = "decreasing"
        else:
            trend_direction = "stable"

        all_predictions = [f["predicted_sales"] for f in forecast]
        avg_forecast = statistics.mean(all_predictions)
        std_forecast = statistics.stdev(all_predictions) if len(all_predictions) > 1 else 0

        confidence_interval = {
            "lower": round(avg_forecast - (1.96 * std_forecast), 2),
            "upper": round(avg_forecast + (1.96 * std_forecast), 2),
            "confidence_level": "95%"
        }

        historical_avg = statistics.mean(historical_sales)
        forecast_avg = statistics.mean(all_predictions)

        return {
            "ok": True,
            "historical_sales": historical_sales,
            "forecast": forecast,
            "trend_direction": trend_direction,
            "trend_slope": round(slope, 4),
            "moving_averages": {
                "3_month": round(ma_3, 2),
                "6_month": round(ma_6, 2) if len(historical_sales) >= 6 else None
            },
            "confidence_interval": confidence_interval,
            "seasonality_detected": seasonality_info.get("has_seasonality", False),
            "summary": {
                "historical_average": round(historical_avg, 2),
                "forecast_average": round(forecast_avg, 2),
                "predicted_total": round(sum(all_predictions), 2),
                "period": f"{future_months} months"
            },
            "generated_at": datetime.now().isoformat()
        }

    except Exception as e:
        return {"ok": False, "error": f"Failed to generate forecast: {str(e)}"}


if __name__ == "__main__":
    result = handler({
        "historical_sales": [10000, 12000, 11500, 14000, 16000, 15500, 18000, 17500, 19000, 21000, 20500, 22000],
        "future_months": 6,
        "seasonality": True
    })
    print(result)
