import math
import random

def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be a dict"}
    initial = event.get("initial_value")
    mu = event.get("expected_return")
    sigma = event.get("volatility")
    years = event.get("years")
    if initial is None or mu is None or sigma is None or years is None:
        return {"ok": False, "error": "initial_value, expected_return, volatility, and years are required"}
    try:
        initial = float(initial)
        mu = float(mu)
        sigma = float(sigma)
        years = float(years)
        n_sim = int(event.get("simulations", 1000))
        seed = event.get("seed")
        if seed is not None:
            random.seed(int(seed))
        dt = 1 / 252
        steps = int(years * 252)
        final_values = []
        for _ in range(n_sim):
            value = initial
            for _ in range(steps):
                z = random.gauss(0, 1)
                value *= math.exp((mu - 0.5 * sigma ** 2) * dt + sigma * math.sqrt(dt) * z)
            final_values.append(value)
        final_values.sort()
        mean_val = sum(final_values) / n_sim
        median_val = final_values[n_sim // 2]
        p5 = final_values[int(0.05 * n_sim)]
        p95 = final_values[int(0.95 * n_sim)]
        return {
            "ok": True,
            "result": {
                "mean": round(mean_val, 2),
                "median": round(median_val, 2),
                "p5": round(p5, 2),
                "p95": round(p95, 2),
                "simulations": n_sim
            }
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
