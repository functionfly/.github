import math

def _norm_cdf(x):
    return 0.5 * (1 + math.erf(x / math.sqrt(2)))

def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be a dict"}
    S = event.get("spot_price")
    K = event.get("strike_price")
    T = event.get("time_to_expiry")
    r = event.get("risk_free_rate")
    if S is None or K is None or T is None or r is None:
        return {"ok": False, "error": "spot_price, strike_price, time_to_expiry, and risk_free_rate are required"}
    try:
        S = float(S)
        K = float(K)
        T = float(T)
        r = float(r)
        call_price = event.get("call_price")
        put_price = event.get("put_price")
        pv_strike = K * math.exp(-r * T)
        # C - P = S - PV(K)
        if call_price is not None and put_price is not None:
            call_price = float(call_price)
            put_price = float(put_price)
            lhs = call_price - put_price
            rhs = S - pv_strike
            parity_holds = abs(lhs - rhs) < 0.01
            theoretical_call = put_price + S - pv_strike
            theoretical_put = call_price - S + pv_strike
        elif call_price is not None:
            call_price = float(call_price)
            theoretical_put = call_price - S + pv_strike
            theoretical_call = call_price
            parity_holds = True
        elif put_price is not None:
            put_price = float(put_price)
            theoretical_call = put_price + S - pv_strike
            theoretical_put = put_price
            parity_holds = True
        else:
            return {"ok": False, "error": "At least one of call_price or put_price is required"}
        return {
            "ok": True,
            "result": {
                "parity_holds": parity_holds,
                "theoretical_call": round(theoretical_call, 6),
                "theoretical_put": round(theoretical_put, 6),
                "pv_strike": round(pv_strike, 6)
            }
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
