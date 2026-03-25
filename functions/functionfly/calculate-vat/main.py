def handler(event):
    price = event.get("price") if isinstance(event, dict) else None
    rate = event.get("rate")
    inclusive = event.get("inclusive", False)
    if price is None or rate is None:
        return {"ok": False, "error": "price and rate are required"}
    try:
        p, r = float(price), float(rate)
        if inclusive:
            vat = round(p * r / (100 + r), 2)
            net = round(p - vat, 2)
            return {"ok": True, "result": vat, "vat": vat, "net": net, "gross": p, "rate": r}
        else:
            vat = round(p * r / 100, 2)
            gross = round(p + vat, 2)
            return {"ok": True, "result": vat, "vat": vat, "net": p, "gross": gross, "rate": r}
    except Exception as e:
        return {"ok": False, "error": str(e)}
