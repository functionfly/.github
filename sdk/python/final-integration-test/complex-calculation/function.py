@flypy.function(
    name="complex-calculation",
    description="Complex calculation with optimization opportunities",
    deterministic=True
)
def complex_calculation(input_data: Dict[str, Any]) -> Dict[str, Any]:
    """Perform complex calculations with dead code and constant folding opportunities."""

    # Constants that can be folded
    TAX_RATE = 0.08
    DISCOUNT_THRESHOLD = 100.0
    PREMIUM_MULTIPLIER = 1.5

    # Dead code that should be removed
    unused_variable = "this will be removed"
    another_unused = 42
    never_called_function = lambda x: x * 2

    # Unused import that should be removed
    import math
    import json

    # Actual processing
    base_amount = input_data.get("amount", 0.0)
    is_premium = input_data.get("is_premium", False)

    # Calculations with constant folding opportunities
    discount = base_amount * 0.1 if base_amount > DISCOUNT_THRESHOLD else 0.0
    discounted_amount = base_amount - discount

    tax = discounted_amount * TAX_RATE
    final_amount = discounted_amount + tax

    if is_premium:
        final_amount = final_amount * PREMIUM_MULTIPLIER

    # More dead code
    unused_list = [1, 2, 3, 4, 5]
    unused_dict = {"key": "value"}

    return {
        "original_amount": base_amount,
        "discount": discount,
        "tax": tax,
        "final_amount": final_amount,
        "is_premium": is_premium
    }
