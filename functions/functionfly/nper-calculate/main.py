import math

def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be a dict"}
    rate = event.get("rate")
    payment = event.get("payment")
    pv = event.get("present_value")
    if rate is None or payment is None or pv is None:
        return {"ok": False, "error": "rate, payment, and present_value are required"}
    try:
        rate = float(rate)
        payment = float(payment)
        pv = float(pv)
        fv = float(event.get("future_value", 0))
        if rate == 0:
            if payment == 0:
                return {"ok": False, "error": "payment cannot be zero when rate is zero"}
            nper = -(pv + fv) / payment
        else:
            numerator = -fv * rate + payment
            denominator = pv * rate + payment
            if denominator == 0:
                return {"ok": False, "error": "Cannot compute nper with given inputs (denominator is zero)"}
            ratio = numerator / denominator
            if ratio <= 0:
                return {"ok": False, "error": "Cannot compute nper with given inputs (negative ratio)"}
            nper = math.log(ratio) / math.log(1 + rate)
        return {"ok": True, "result": round(nper, 6)}
    except Exception as e:
        return {"ok": False, "error": str(e)}
