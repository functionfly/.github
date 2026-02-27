#!/usr/bin/env python3
"""
Comprehensive example functions for testing FlyPy end-to-end compilation.

This file contains various example functions that test different scenarios:
- Basic arithmetic operations
- Data processing and transformation
- Schema validation
- Performance monitoring
- Optimization opportunities
- Error handling
"""

import flypy
from typing import Dict, List, Any


# Basic arithmetic function
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


# Data transformation function
@flypy.function(
    name="transform-user-data",
    description="Transform user data for analytics",
    deterministic=True,
    enable_performance_monitoring=True
)
def transform_user_data(user_data: Dict[str, Any]) -> Dict[str, Any]:
    """Transform raw user data into analytics-ready format."""
    # Extract and normalize user information
    user = {
        "user_id": user_data["id"],
        "full_name": f"{user_data['first_name']} {user_data['last_name']}",
        "email_domain": user_data["email"].split("@")[1],
        "age_group": "18-24" if user_data["age"] < 25 else "25+",
        "account_status": "active" if user_data["is_active"] else "inactive"
    }

    # Process activity data
    activities = user_data.get("activities", [])
    activity_summary = {
        "total_activities": len(activities),
        "last_activity": max(act["timestamp"] for act in activities) if activities else None,
        "activity_types": list(set(act["type"] for act in activities))
    }

    # Calculate engagement score (simplified)
    engagement_score = min(100, len(activities) * 10 + (user_data.get("login_count", 0) * 5))

    return {
        "user": user,
        "activity_summary": activity_summary,
        "engagement_score": engagement_score,
        "processed_at": "2024-01-01T00:00:00Z"  # Would be dynamic in real implementation
    }


# List processing function
@flypy.function(
    name="process-inventory",
    description="Process inventory data with filtering and aggregation",
    deterministic=True
)
def process_inventory(inventory_data: Dict[str, Any]) -> Dict[str, Any]:
    """Process inventory data with filtering and statistics."""
    items = inventory_data.get("items", [])
    filters = inventory_data.get("filters", {})

    # Apply filters
    filtered_items = []
    for item in items:
        # Category filter
        if "category" in filters and item.get("category") != filters["category"]:
            continue

        # Price range filter
        if "min_price" in filters and item.get("price", 0) < filters["min_price"]:
            continue
        if "max_price" in filters and item.get("price", 0) > filters["max_price"]:
            continue

        # Stock level filter
        if "min_stock" in filters and item.get("stock_quantity", 0) < filters["min_stock"]:
            continue

        filtered_items.append(item)

    # Calculate statistics
    if filtered_items:
        prices = [item["price"] for item in filtered_items]
        stocks = [item["stock_quantity"] for item in filtered_items]

        stats = {
            "total_items": len(filtered_items),
            "avg_price": sum(prices) / len(prices),
            "min_price": min(prices),
            "max_price": max(prices),
            "total_value": sum(price * stock for price, stock in zip(prices, stocks)),
            "low_stock_items": len([s for s in stocks if s < 10])
        }
    else:
        stats = {
            "total_items": 0,
            "avg_price": 0.0,
            "min_price": 0.0,
            "max_price": 0.0,
            "total_value": 0.0,
            "low_stock_items": 0
        }

    return {
        "filtered_items": filtered_items,
        "statistics": stats,
        "filters_applied": filters
    }


# Function with complex optimization opportunities
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


# Function with schema validation
@flypy.input_schema({
    "type": "object",
    "properties": {
        "name": {"type": "string"},
        "age": {"type": "integer", "minimum": 0},
        "email": {"type": "string", "format": "email"}
    },
    "required": ["name", "age"]
})
@flypy.output_schema({
    "type": "object",
    "properties": {
        "user_id": {"type": "string"},
        "profile_complete": {"type": "boolean"},
        "age_group": {"type": "string"},
        "validation_status": {"type": "string"}
    }
})
@flypy.function(
    name="validate-user-profile",
    description="Validate and process user profile data",
    deterministic=True
)
def validate_user_profile(user_data: Dict[str, Any]) -> Dict[str, Any]:
    """Validate user profile and return processed information."""

    # Generate a simple user ID
    user_id = f"user_{hash(user_data['name'] + str(user_data['age'])) % 10000}"

    # Determine if profile is complete
    required_fields = ["name", "age", "email"]
    profile_complete = all(field in user_data for field in required_fields)

    # Age group classification
    age = user_data["age"]
    if age < 18:
        age_group = "minor"
    elif age < 25:
        age_group = "young_adult"
    elif age < 65:
        age_group = "adult"
    else:
        age_group = "senior"

    # Validation status
    validation_status = "valid" if profile_complete else "incomplete"

    return {
        "user_id": user_id,
        "profile_complete": profile_complete,
        "age_group": age_group,
        "validation_status": validation_status
    }


# Error handling function
@flypy.function(
    name="safe-division",
    description="Perform safe division with error handling",
    deterministic=True,
    enable_performance_monitoring=True
)
def safe_division(calculation_data: Dict[str, Any]) -> Dict[str, Any]:
    """Perform division with comprehensive error handling."""

    try:
        numerator = calculation_data.get("numerator", 0)
        denominator = calculation_data.get("denominator", 1)

        # Validate inputs
        if not isinstance(numerator, (int, float)):
            raise ValueError("Numerator must be a number")
        if not isinstance(denominator, (int, float)):
            raise ValueError("Denominator must be a number")
        if denominator == 0:
            raise ZeroDivisionError("Division by zero")

        result = numerator / denominator

        return {
            "result": result,
            "status": "success",
            "error": None
        }

    except ZeroDivisionError:
        return {
            "result": None,
            "status": "error",
            "error": "division_by_zero"
        }
    except ValueError as e:
        return {
            "result": None,
            "status": "error",
            "error": str(e)
        }
    except Exception as e:
        return {
            "result": None,
            "status": "error",
            "error": f"unexpected_error: {str(e)}"
        }


if __name__ == "__main__":
    print("FlyPy Example Functions")
    print("=" * 50)

    # List all registered functions
    functions = flypy.get_registered_functions()
    print(f"Registered functions: {len(functions)}")

    for name, func_def in functions.items():
        print(f"- {name}: {func_def.metadata.description}")

    print("\n✅ Example functions loaded successfully!")