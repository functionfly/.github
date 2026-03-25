import random


def _ean13_checksum(digits_12):
    total = sum(d * (3 if i % 2 else 1) for i, d in enumerate(digits_12))
    return (10 - (total % 10)) % 10


def _upc_checksum(digits_11):
    odd = sum(digits_11[i] for i in range(0, 11, 2))
    even = sum(digits_11[i] for i in range(1, 11, 2))
    return (10 - (odd * 3 + even) % 10) % 10


def handler(event):
    barcode_type = event.get("type", "EAN-13") if isinstance(event, dict) else "EAN-13"
    data = event.get("data")
    prefix = str(event.get("prefix", "")) if isinstance(event, dict) else ""
    try:
        if barcode_type == "EAN-13":
            if data:
                digits = [int(c) for c in str(data) if c.isdigit()][:12]
                while len(digits) < 12:
                    digits.append(random.randint(0, 9))
            elif prefix:
                pdigits = [int(c) for c in str(prefix) if c.isdigit()][:3]
                while len(pdigits) < 3:
                    pdigits.append(0)
                digits = pdigits + [random.randint(0, 9) for _ in range(9)]
            else:
                digits = [random.randint(0, 9) for _ in range(12)]
            check = _ean13_checksum(digits)
            result = "".join(map(str, digits)) + str(check)
            return {"ok": True, "result": result, "barcode": result, "type": "EAN-13", "check_digit": check}
        elif barcode_type == "UPC-A":
            if data:
                digits = [int(c) for c in str(data) if c.isdigit()][:11]
                while len(digits) < 11:
                    digits.append(random.randint(0, 9))
            else:
                digits = [random.randint(0, 9) for _ in range(11)]
            check = _upc_checksum(digits)
            result = "".join(map(str, digits)) + str(check)
            return {"ok": True, "result": result, "barcode": result, "type": "UPC-A", "check_digit": check}
        elif barcode_type == "CODE128":
            payload = str(data) if data else str(random.randint(10000000, 99999999))
            return {"ok": True, "result": payload, "barcode": payload, "type": "CODE128", "note": "CODE128 encoding requires a library like python-barcode for visual output"}
        else:
            return {"ok": False, "error": f"unsupported type: {barcode_type}. Use EAN-13, UPC-A, or CODE128"}
    except Exception as e:
        return {"ok": False, "error": str(e)}
