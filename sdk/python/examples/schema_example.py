"""
Schema example for FlyPy.

This example demonstrates explicit schema definition and validation.
"""

import flypy
from flypy import Schema, Field


# Define explicit schemas
input_schema = Schema("User Data Input")
input_schema.add_field("user_id", Field("string", required=True, min_length=1))
input_schema.add_field("name", Field("string", required=True, min_length=2, max_length=100))
input_schema.add_field("email", Field("string", required=True, pattern=r"^[^@]+@[^@]+\.[^@]+$"))
input_schema.add_field("age", Field("integer", required=True, minimum=0, maximum=150))
input_schema.add_field("preferences", Field("object", required=False, additional_properties=True))

output_schema = Schema("User Profile Output")
output_schema.add_field("user_id", Field("string", required=True))
output_schema.add_field("name", Field("string", required=True))
output_schema.add_field("email", Field("string", required=True))
output_schema.add_field("age", Field("integer", required=True))
output_schema.add_field("is_adult", Field("boolean", required=True))
output_schema.add_field("profile_complete", Field("boolean", required=True))
output_schema.add_field("preferences", Field("object", required=True))


@flypy.input_schema(input_schema)
@flypy.output_schema(output_schema)
@flypy.function(
    name="process-user-profile",
    description="Process and validate user profile data",
    deterministic=True,
    idempotent=True,
    pure=True
)
def process_user_profile(user_data: dict) -> dict:
    """
    Process user profile data with validation.

    Args:
        user_data: User data dictionary

    Returns:
        Processed user profile
    """
    user_id = user_data["user_id"]
    name = user_data["name"]
    email = user_data["email"]
    age = user_data["age"]
    preferences = user_data.get("preferences", {})

    # Business logic
    is_adult = age >= 18
    profile_complete = bool(name and email and age >= 0)

    # Process preferences
    processed_preferences = {
        "theme": preferences.get("theme", "light"),
        "notifications": preferences.get("notifications", True),
        "language": preferences.get("language", "en"),
    }

    return {
        "user_id": user_id,
        "name": name.strip(),
        "email": email.lower(),
        "age": age,
        "is_adult": is_adult,
        "profile_complete": profile_complete,
        "preferences": processed_preferences
    }


# Example with JSON schema dictionaries
input_schema_dict = {
    "type": "object",
    "properties": {
        "product_id": {"type": "string", "minLength": 1},
        "name": {"type": "string", "minLength": 1},
        "price": {"type": "number", "minimum": 0},
        "category": {"type": "string", "enum": ["electronics", "books", "clothing"]},
        "in_stock": {"type": "boolean"},
        "tags": {"type": "array", "items": {"type": "string"}}
    },
    "required": ["product_id", "name", "price"]
}

output_schema_dict = {
    "type": "object",
    "properties": {
        "product_id": {"type": "string"},
        "name": {"type": "string"},
        "price": {"type": "number"},
        "category": {"type": "string"},
        "in_stock": {"type": "boolean"},
        "tags": {"type": "array", "items": {"type": "string"}},
        "discounted_price": {"type": "number"},
        "is_on_sale": {"type": "boolean"}
    },
    "required": ["product_id", "name", "price", "is_on_sale"]
}


@flypy.input_schema(input_schema_dict)
@flypy.output_schema(output_schema_dict)
@flypy.function(
    name="process-product",
    description="Process product data with discount calculation",
    deterministic=True,
    idempotent=True,
    pure=True,
    cache_ttl=3600
)
def process_product(product: dict) -> dict:
    """
    Process product data and calculate discounts.

    Args:
        product: Product data dictionary

    Returns:
        Processed product with discount information
    """
    product_id = product["product_id"]
    name = product["name"]
    price = product["price"]
    category = product.get("category", "")
    in_stock = product.get("in_stock", True)
    tags = product.get("tags", [])

    # Calculate discount based on category
    discount_rate = 0.0
    if category == "electronics":
        discount_rate = 0.1  # 10% off electronics
    elif category == "books":
        discount_rate = 0.2  # 20% off books
    elif not in_stock:
        discount_rate = 0.5  # 50% off out of stock items

    discounted_price = price * (1 - discount_rate)
    is_on_sale = discount_rate > 0

    return {
        "product_id": product_id,
        "name": name,
        "price": price,
        "category": category,
        "in_stock": in_stock,
        "tags": tags,
        "discounted_price": round(discounted_price, 2),
        "is_on_sale": is_on_sale
    }