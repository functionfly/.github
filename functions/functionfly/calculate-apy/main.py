def handler(event):
    rate = event.get("rate") if isinstance(event, dict) else None
    compounds_per_year = int(event.get("compounds_per_year", 12))
    if rate is None:
        return {"ok": False, "error": "rate is required (annual rate %)"}
    try:
        r = float(rate) / 100
        n = compounds_per_year
        apy = round(((1 + r / n) ** n - 1) * 100, 6)
        return {"ok": True, "result": apy, "apy_pct": apy, "apr_pct": float(rate), "compounds_per_year": n}
    except Exception as e:
        return {"ok": False, "error": str(e)}
