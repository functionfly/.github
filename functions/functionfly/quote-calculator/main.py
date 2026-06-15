"""Quote Calculator - Generate a quote with line items."""


def handler(event):
    try:
        items = event.get("items", [])
        discount_percent = event.get("discount_percent")
        tax_percent = event.get("tax_percent")
        valid_days = int(event.get("valid_days", 30))

        if not isinstance(items, list):
            return {"ok": False, "error": "items must be a list"}
        if len(items) == 0:
            return {"ok": False, "error": "items cannot be empty"}
        if valid_days < 1 or valid_days > 365:
            return {"ok": False, "error": "valid_days must be between 1 and 365"}

        for i, item in enumerate(items):
            if not isinstance(item, dict):
                return {"ok": False, "error": f"item {i} must be an object"}
            desc = item.get("description")
            qty = item.get("quantity")
            unit_price = item.get("unit_price")
            if not desc or not isinstance(desc, str):
                return {"ok": False, "error": f"item {i} description is required"}
            if qty is None or not isinstance(qty, (int, float)) or qty <= 0:
                return {"ok": False, "error": f"item {i} quantity must be positive"}
            if unit_price is None or not isinstance(unit_price, (int, float)) or unit_price < 0:
                return {"ok": False, "error": f"item {i} unit_price must be non-negative"}

        if discount_percent is not None:
            discount_percent = float(discount_percent)
            if discount_percent < 0 or discount_percent > 100:
                return {"ok": False, "error": "discount_percent must be between 0 and 100"}

        if tax_percent is not None:
            tax_percent = float(tax_percent)
            if tax_percent < 0 or tax_percent > 100:
                return {"ok": False, "error": "tax_percent must be between 0 and 100"}

        import datetime
        quote_number = f"Q-{datetime.datetime.now().strftime('%Y%m%d%H%M%S')}"
        valid_until = (datetime.datetime.now() + datetime.timedelta(days=valid_days)).isoformat()

        subtotal = sum(item["quantity"] * item["unit_price"] for item in items)
        subtotal = round(subtotal, 2)

        result = {
            "ok": True,
            "subtotal": subtotal,
            "quote_number": quote_number,
            "valid_until": valid_until
        }

        if discount_percent is not None and discount_percent > 0:
            discount = round(subtotal * (discount_percent / 100.0), 2)
            result["discount"] = discount
            taxable_amount = subtotal - discount
        else:
            taxable_amount = subtotal

        if tax_percent is not None and tax_percent > 0:
            tax = round(taxable_amount * (tax_percent / 100.0), 2)
            result["tax"] = tax
            total = round(taxable_amount + tax, 2)
        else:
            total = taxable_amount

        result["total"] = round(total, 2)

        return result
    except (ValueError, TypeError) as e:
        return {"ok": False, "error": f"Invalid input: {str(e)}"}
    except Exception as e:
        return {"ok": False, "error": f"Internal error: {str(e)}"}
