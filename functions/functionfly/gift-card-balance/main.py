def handler(event):
    balance = event.get("balance") if isinstance(event, dict) else None
    amount_to_charge = event.get("amount_to_charge")
    action = event.get("action", "check")
    if balance is None:
        return {"ok": False, "error": "balance is required"}
    try:
        bal = float(balance)
        if action == "check":
            return {"ok": True, "result": bal, "balance": bal, "can_use": bal > 0}
        elif action == "charge":
            if amount_to_charge is None:
                return {"ok": False, "error": "amount_to_charge required for charge action"}
            charge = float(amount_to_charge)
            if charge > bal:
                applied = bal
                remaining_to_pay = round(charge - bal, 2)
                new_balance = 0.0
            else:
                applied = round(charge, 2)
                remaining_to_pay = 0.0
                new_balance = round(bal - charge, 2)
            return {"ok": True, "result": new_balance, "previous_balance": bal, "applied": applied, "new_balance": new_balance, "remaining_to_pay": remaining_to_pay}
        else:
            return {"ok": False, "error": "action must be 'check' or 'charge'"}
    except Exception as e:
        return {"ok": False, "error": str(e)}
