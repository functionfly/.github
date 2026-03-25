import math


def handler(event):
    principal = event.get("principal") if isinstance(event, dict) else None
    installments = event.get("installments")
    annual_rate = event.get("annual_rate", 0)
    if principal is None or installments is None:
        return {"ok": False, "error": "principal and installments are required"}
    try:
        p = float(principal)
        n = int(installments)
        r = float(annual_rate) / 100 / 12
        if r == 0:
            monthly = round(p / n, 2)
            total = round(monthly * n, 2)
            interest = 0
        else:
            monthly = round(p * r * (1 + r) ** n / ((1 + r) ** n - 1), 2)
            total = round(monthly * n, 2)
            interest = round(total - p, 2)
        return {"ok": True, "result": monthly, "monthly_payment": monthly, "total": total, "interest": interest, "installments": n}
    except Exception as e:
        return {"ok": False, "error": str(e)}
