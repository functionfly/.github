import re


def handler(event):
    clicks = event.get("clicks", 0) if isinstance(event, dict) else 0
    impressions = event.get("impressions", 0)
    conversions = event.get("conversions", 0)
    spend = event.get("spend", 0)
    revenue = event.get("revenue", 0)
    campaign_name = event.get("campaign_name", "")
    try:
        c, imp, conv, sp, rev = float(clicks), float(impressions), float(conversions), float(spend), float(revenue)
        ctr = round((c / imp * 100), 4) if imp > 0 else 0
        conversion_rate = round((conv / c * 100), 4) if c > 0 else 0
        cpc = round(sp / c, 4) if c > 0 else 0
        cpm = round(sp / imp * 1000, 4) if imp > 0 else 0
        cpa = round(sp / conv, 4) if conv > 0 else 0
        roas = round(rev / sp, 4) if sp > 0 else 0
        roi = round((rev - sp) / sp * 100, 2) if sp > 0 else 0
        return {
            "ok": True,
            "result": {
                "ctr": ctr,
                "conversion_rate": conversion_rate,
                "cpc": cpc,
                "cpm": cpm,
                "cpa": cpa,
                "roas": roas,
                "roi": roi
            },
            "campaign_name": campaign_name,
            "clicks": c,
            "impressions": imp,
            "conversions": conv,
            "spend": sp,
            "revenue": rev,
            "ctr_percent": f"{ctr}%",
            "conversion_rate_percent": f"{conversion_rate}%",
            "roi_percent": f"{roi}%",
            "roas": roas,
            "cpc": cpc,
            "cpm": cpm,
            "cpa": cpa
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
