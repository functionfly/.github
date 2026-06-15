"""Discount Calculator - Calculate discounted price with optional tax."""


def handler(event):
    try:
        original_price = float(event.get("original_price", 0))
        discount_percent = float(event.get("discount_percent", 0))
        tax_percent = event.get("tax_percent")

        if original_price < 0:
            return {"ok": False, "error": "original_price cannot be negative"}
        if discount_percent < 0 or discount_percent > 100:
            return {"ok": False, "error": "discount_percent must be between 0 and 100"}
        if tax_percent is not None:
            tax_percent = float(tax_percent)
            if tax_percent < 0 or tax_percent > 100:
                return {"ok": False, "error": "tax_percent must be between 0 and 100"}

        discount_amount = round(original_price * (discount_percent / 100.0), 2)
        final_price = round(original_price - discount_amount, 2)

        result = {
            "ok": True,
            "discount_amount": discount_amount,
            "final_price": final_price
        }

        if tax_percent is not None:
            tax_amount = round(final_price * (tax_percent / 100.0), 2)
            total = round(final_price + tax_amount, 2)
            result["tax_amount"] = tax_amount
            result["total"] = total
        else:
            result["total"] = final_price

        return result
    except (ValueError, TypeError) as e:
        return {"ok": False, "error": f"Invalid input: {str(e)}"}
    except Exception as e:
        return {"ok": False, "error": f"Internal error: {str(e)}"}
