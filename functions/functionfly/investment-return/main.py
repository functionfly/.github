def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be a dict"}
    initial = event.get("initial_value")
    final = event.get("final_value")
    if initial is None or final is None:
        return {"ok": False, "error": "initial_value and final_value are required"}
    try:
        initial = float(initial)
        final = float(final)
        years = float(event.get("years", 1))
        if initial == 0:
            return {"ok": False, "error": "initial_value cannot be zero"}
        total_return = (final - initial) / initial
        if years > 0:
            annualized = (final / initial) ** (1 / years) - 1
        else:
            annualized = None
        result = {
            "ok": True,
            "result": round(total_return, 6),
            "total_return": round(total_return, 6),
            "total_return_pct": round(total_return * 100, 4)
        }
        if annualized is not None:
            result["annualized_return"] = round(annualized, 6)
        return result
    except Exception as e:
        return {"ok": False, "error": str(e)}
