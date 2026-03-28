def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be a dict"}
    cash_flows = event.get("cash_flows")
    discount_rate = event.get("discount_rate")
    if cash_flows is None or discount_rate is None:
        return {"ok": False, "error": "cash_flows and discount_rate are required"}
    try:
        cash_flows = [float(cf) for cf in cash_flows]
        r = float(discount_rate)
        g = float(event.get("terminal_growth_rate", 0.02))
        if r <= g:
            return {"ok": False, "error": "discount_rate must be greater than terminal_growth_rate"}
        # PV of explicit cash flows
        pv_cfs = sum(cf / (1 + r) ** (t + 1) for t, cf in enumerate(cash_flows))
        # Terminal value (Gordon Growth Model)
        last_cf = cash_flows[-1]
        terminal_value = last_cf * (1 + g) / (r - g)
        pv_terminal = terminal_value / (1 + r) ** len(cash_flows)
        total = pv_cfs + pv_terminal
        return {
            "ok": True,
            "result": round(total, 2),
            "pv_cash_flows": round(pv_cfs, 2),
            "pv_terminal_value": round(pv_terminal, 2),
            "terminal_value": round(terminal_value, 2)
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
