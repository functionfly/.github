import time, random, string


def handler(event):
    prefix = event.get("prefix", "INV") if isinstance(event, dict) else "INV"
    sequence = event.get("sequence")
    date_based = event.get("date_based", True)

    try:
        from datetime import datetime
        now = datetime.utcnow()
        if date_based:
            date_part = now.strftime("%Y%m%d")
            if sequence is not None:
                result = f"{prefix}-{date_part}-{int(sequence):05d}"
            else:
                rand_part = "".join(random.choices(string.digits, k=5))
                result = f"{prefix}-{date_part}-{rand_part}"
        else:
            if sequence is not None:
                result = f"{prefix}-{int(sequence):08d}"
            else:
                ts = str(int(time.time() * 1000))[-8:]
                rand = "".join(random.choices(string.ascii_uppercase + string.digits, k=4))
                result = f"{prefix}-{ts}-{rand}"
        return {"ok": True, "result": result, "invoice_number": result, "prefix": prefix}
    except Exception as e:
        return {"ok": False, "error": str(e)}
