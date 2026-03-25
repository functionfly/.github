def handler(event):
    price = event.get("price") if isinstance(event, dict) else None
    rate = event.get("rate", 10)
    inclusive = event.get("inclusive", False)
    country = event.get("country", "AU")
    if price is None:
        return {"ok": False, "error": "price is required"}
    try:
        p, r = float(price), float(rate)
        if inclusive:
            gst = round(p * r / (100 + r), 2)
            net = round(p - gst, 2)
            return {"ok": True, "result": gst, "gst": gst, "net": net, "gross": p, "rate": r, "country": country}
        else:
            gst = round(p * r / 100, 2)
            return {"ok": True, "result": gst, "gst": gst, "net": p, "gross": round(p + gst, 2), "rate": r, "country": country}
    except Exception as e:
        return {"ok": False, "error": str(e)}
