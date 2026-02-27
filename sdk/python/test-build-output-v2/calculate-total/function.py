@flypy.function(
    name="calculate-total",
    description="Calculate order total with tax",
    deterministic=True,
    idempotent=True,
    cache_ttl=3600
)
def calculate_total(order_data: Dict[str, Any]) -> Dict[str, float]:
    """Calculate order total with tax and discounts."""
    items = order_data.get("items", [])
    tax_rate = order_data.get("tax_rate", 0.08)
    discount_percent = order_data.get("discount_percent", 0.0)

    # Calculate subtotal
    subtotal = sum(item["price"] * item["quantity"] for item in items)

    # Apply discount
    discount_amount = subtotal * discount_percent
    discounted_subtotal = subtotal - discount_amount

    # Calculate tax on discounted amount
    tax = discounted_subtotal * tax_rate

    # Calculate final total
    total = discounted_subtotal + tax

    return {
        "subtotal": subtotal,
        "discount_amount": discount_amount,
        "tax": tax,
        "total": total
    }
