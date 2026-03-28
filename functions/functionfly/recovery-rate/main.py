def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be a dict"}
    face_value = event.get("face_value")
    recovery_amount = event.get("recovery_amount")
    if face_value is None or recovery_amount is None:
        return {"ok": False, "error": "face_value and recovery_amount are required"}
    try:
        face_value = float(face_value)
        recovery_amount = float(recovery_amount)
        accrued = float(event.get("accrued_interest", 0))
        total_claim = face_value + accrued
        if total_claim == 0:
            return {"ok": False, "error": "face_value cannot be zero"}
        recovery_rate = recovery_amount / total_claim
        lgd = 1 - recovery_rate
        loss_amount = total_claim - recovery_amount
        return {
            "ok": True,
            "result": round(recovery_rate, 6),
            "lgd": round(lgd, 6),
            "loss_amount": round(loss_amount, 2)
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
