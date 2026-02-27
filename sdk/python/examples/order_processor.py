"""
Order processing example for FlyPy.

This example demonstrates more complex schema usage with type hints
and business logic functions.
"""

from typing import List, Dict, Any, Optional
from datetime import datetime
import flypy


# Define types for better schema inference
class OrderItem:
    """An item in an order."""
    product_id: str
    name: str
    quantity: int
    price: float


class Order:
    """An order object."""
    order_id: str
    customer_id: str
    items: List[OrderItem]
    tax_rate: float
    discount: Optional[float] = 0.0


@flypy.function(
    name="calculate-order-total",
    description="Calculate total cost for an order including tax and discounts",
    deterministic=True,
    idempotent=True,
    pure=True,
    cache_ttl=1800
)
def calculate_order_total(order: Dict[str, Any]) -> Dict[str, float]:
    """
    Calculate the total cost for an order.

    Args:
        order: Order data with items, tax rate, and discount

    Returns:
        Dictionary with subtotal, tax, discount, and total
    """
    items = order.get("items", [])
    tax_rate = order.get("tax_rate", 0.08)
    discount = order.get("discount", 0.0)

    # Calculate subtotal
    subtotal = 0.0
    for item in items:
        subtotal += item["price"] * item["quantity"]

    # Apply discount
    discount_amount = subtotal * discount

    # Calculate tax on pre-discount amount
    taxable_amount = subtotal - discount_amount
    tax_amount = taxable_amount * tax_rate

    # Calculate final total
    total = taxable_amount + tax_amount

    return {
        "subtotal": round(subtotal, 2),
        "discount_amount": round(discount_amount, 2),
        "taxable_amount": round(taxable_amount, 2),
        "tax_amount": round(tax_amount, 2),
        "total": round(total, 2)
    }


@flypy.function(
    name="validate-order",
    description="Validate order data and check business rules",
    deterministic=True,
    idempotent=True,
    pure=True
)
def validate_order(order: Dict[str, Any]) -> Dict[str, Any]:
    """
    Validate an order and check business rules.

    Args:
        order: Order data to validate

    Returns:
        Validation result with errors if any
    """
    errors = []
    warnings = []

    # Check required fields
    if not order.get("order_id"):
        errors.append("order_id is required")
    if not order.get("customer_id"):
        errors.append("customer_id is required")

    items = order.get("items", [])
    if not items:
        errors.append("order must contain at least one item")

    # Validate items
    total_quantity = 0
    for i, item in enumerate(items):
        if not item.get("product_id"):
            errors.append(f"item {i}: product_id is required")
        if not item.get("name"):
            errors.append(f"item {i}: name is required")

        quantity = item.get("quantity", 0)
        if not isinstance(quantity, int) or quantity <= 0:
            errors.append(f"item {i}: quantity must be a positive integer")
        else:
            total_quantity += quantity

        price = item.get("price", 0)
        if not isinstance(price, (int, float)) or price < 0:
            errors.append(f"item {i}: price must be a non-negative number")

    # Business rules
    if total_quantity > 100:
        warnings.append("Large order - may require special handling")

    tax_rate = order.get("tax_rate", 0)
    if not (0 <= tax_rate <= 0.5):
        errors.append("tax_rate must be between 0 and 0.5")

    discount = order.get("discount", 0)
    if not (0 <= discount <= 1):
        errors.append("discount must be between 0 and 1")

    return {
        "valid": len(errors) == 0,
        "errors": errors,
        "warnings": warnings,
        "item_count": len(items),
        "total_quantity": total_quantity
    }


@flypy.function(
    name="process-order",
    description="Complete order processing workflow",
    deterministic=True,
    idempotent=False  # Not idempotent due to potential external calls
)
def process_order(order: Dict[str, Any]) -> Dict[str, Any]:
    """
    Process a complete order including validation and calculation.

    Args:
        order: Order data to process

    Returns:
        Processing result
    """
    # Validate order first
    validation = validate_order(order)
    if not validation["valid"]:
        return {
            "success": False,
            "errors": validation["errors"],
            "order_id": order.get("order_id")
        }

    # Calculate totals
    totals = calculate_order_total(order)

    # Generate order summary
    summary = {
        "success": True,
        "order_id": order["order_id"],
        "customer_id": order["customer_id"],
        "processed_at": "2024-01-01T00:00:00Z",  # Would be datetime.now() in real code
        "item_count": len(order["items"]),
        "validation": validation,
        "totals": totals
    }

    return summary