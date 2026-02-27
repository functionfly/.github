# Advanced Schema Customization

This guide covers advanced schema features in FlyPy, including custom types, conditional validation, schema composition, and performance optimization.

## Table of Contents

- [Custom Type Validation](#custom-type-validation)
- [Conditional Schemas](#conditional-schemas)
- [Schema Composition and Reuse](#schema-composition-and-reuse)
- [Advanced Field Types](#advanced-field-types)
- [Schema Performance Optimization](#schema-performance-optimization)
- [Custom Validators](#custom-validators)

## Custom Type Validation

### Creating Custom Field Types

You can create custom field types by extending the base `Field` class:

```python
from flypy import Field, Schema
from typing import Any, Dict, List
import re

class EmailField(Field):
    """Custom field for email validation."""

    def __init__(self, **kwargs):
        super().__init__(
            type_="string",
            pattern=r'^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$',
            **kwargs
        )

class PhoneNumberField(Field):
    """Custom field for phone number validation."""

    def __init__(self, **kwargs):
        super().__init__(
            type_="string",
            pattern=r'^\+?1?[-.\s]?\(?([0-9]{3})\)?[-.\s]?([0-9]{3})[-.\s]?([0-9]{4})$',
            **kwargs
        )

class UUIDField(Field):
    """Custom field for UUID validation."""

    def __init__(self, version: int = None, **kwargs):
        pattern = r'^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$'
        if version:
            # UUID version-specific patterns
            if version == 4:
                pattern = r'^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$'

        super().__init__(
            type_="string",
            pattern=pattern,
            **kwargs
        )
```

### Using Custom Fields

```python
import flypy

@flypy.function(name="create-user")
def create_user(user_data: dict) -> dict:
    # Schema will be inferred, but you can also define custom schemas
    pass

# Manual schema with custom fields
user_schema = Schema("User Registration")
user_schema.add_field("email", EmailField(required=True, description="User email address"))
user_schema.add_field("phone", PhoneNumberField(required=False, description="Phone number"))
user_schema.add_field("user_id", UUIDField(version=4, required=True, description="Unique user ID"))

@flypy.input_schema(user_schema)
@flypy.function(name="register-user")
def register_user(user_data: dict) -> dict:
    # Function implementation
    return {"registered": True, "user_id": user_data["user_id"]}
```

## Conditional Schemas

### Dependent Fields

Use conditional logic to make fields required based on other field values:

```python
from flypy import Schema, Field

def create_conditional_schema() -> Schema:
    """Create a schema with conditional validation."""
    schema = Schema("Conditional Form")

    # Base fields
    schema.add_field("form_type", Field(
        "string",
        enum=["personal", "business"],
        required=True
    ))

    # Personal form fields
    schema.add_field("first_name", Field("string", required=True))
    schema.add_field("last_name", Field("string", required=True))

    # Business form fields (conditionally required)
    schema.add_field("company_name", Field("string", required=False))  # Will be validated conditionally
    schema.add_field("tax_id", Field("string", required=False))

    # Email (required for both)
    schema.add_field("email", Field("string", required=True))

    return schema

@flypy.function(name="submit-form")
def submit_form(form_data: dict) -> dict:
    form_type = form_data.get("form_type")

    if form_type == "business":
        # Additional validation for business forms
        if not form_data.get("company_name"):
            raise ValueError("company_name is required for business forms")
        if not form_data.get("tax_id"):
            raise ValueError("tax_id is required for business forms")

    return {"submitted": True, "form_type": form_type}
```

### Schema Variants

Create different schema variants for different use cases:

```python
def create_user_schema(variant: str = "full") -> Schema:
    """Create user schema based on variant."""

    if variant == "minimal":
        schema = Schema("Minimal User")
        schema.add_field("id", Field("string", required=True))
        schema.add_field("email", Field("string", required=True))

    elif variant == "profile":
        schema = Schema("User Profile")
        schema.add_field("id", Field("string", required=True))
        schema.add_field("email", Field("string", required=True))
        schema.add_field("name", Field("string", required=True))
        schema.add_field("avatar_url", Field("string", required=False))

    else:  # full
        schema = Schema("Full User")
        schema.add_field("id", Field("string", required=True))
        schema.add_field("email", Field("string", required=True))
        schema.add_field("name", Field("string", required=True))
        schema.add_field("avatar_url", Field("string", required=False))
        schema.add_field("preferences", Field("object", required=False))
        schema.add_field("created_at", Field("string", required=True))

    return schema

# Usage
@flypy.input_schema(create_user_schema("minimal"))
@flypy.function(name="get-user")
def get_user(user_id: str) -> dict:
    pass

@flypy.input_schema(create_user_schema("profile"))
@flypy.output_schema(create_user_schema("full"))
@flypy.function(name="update-profile")
def update_profile(profile_data: dict) -> dict:
    pass
```

## Schema Composition and Reuse

### Schema Inheritance

Create base schemas and extend them:

```python
class BaseEntitySchema(Schema):
    """Base schema for all entities."""

    def __init__(self, title: str):
        super().__init__(title)
        self.add_field("id", Field("string", required=True, description="Unique identifier"))
        self.add_field("created_at", Field("string", required=True, description="Creation timestamp"))
        self.add_field("updated_at", Field("string", required=True, description="Last update timestamp"))

class UserSchema(BaseEntitySchema):
    """User entity schema."""

    def __init__(self):
        super().__init__("User")
        self.add_field("email", Field("string", required=True, description="Email address"))
        self.add_field("name", Field("string", required=True, description="Full name"))
        self.add_field("role", Field("string", enum=["admin", "user", "guest"], required=True))

class ProductSchema(BaseEntitySchema):
    """Product entity schema."""

    def __init__(self):
        super().__init__("Product")
        self.add_field("name", Field("string", required=True, description="Product name"))
        self.add_field("price", Field("number", minimum=0, required=True))
        self.add_field("category", Field("string", required=True))
        self.add_field("in_stock", Field("boolean", required=True, default=True))
```

### Schema Mixins

Create reusable schema components:

```python
class AddressMixin:
    """Reusable address fields."""

    @staticmethod
    def add_address_fields(schema: Schema, prefix: str = ""):
        """Add address fields to a schema."""
        schema.add_field(f"{prefix}street", Field("string", required=True))
        schema.add_field(f"{prefix}city", Field("string", required=True))
        schema.add_field(f"{prefix}state", Field("string", required=True))
        schema.add_field(f"{prefix}zip_code", Field("string", required=True))
        schema.add_field(f"{prefix}country", Field("string", required=True, default="USA"))

class ContactMixin:
    """Reusable contact fields."""

    @staticmethod
    def add_contact_fields(schema: Schema):
        """Add contact fields to a schema."""
        schema.add_field("email", Field("string", required=True))
        schema.add_field("phone", Field("string", required=False))

class OrderSchema(Schema):
    """Order schema with mixins."""

    def __init__(self):
        super().__init__("Order")

        # Basic order fields
        self.add_field("order_id", Field("string", required=True))
        self.add_field("customer_id", Field("string", required=True))

        # Add address mixins for billing and shipping
        AddressMixin.add_address_fields(self, "billing_")
        AddressMixin.add_address_fields(self, "shipping_")

        # Add contact info
        ContactMixin.add_contact_fields(self)

        # Order-specific fields
        self.add_field("items", Field(
            "array",
            required=True,
            items={
                "type": "object",
                "properties": {
                    "product_id": {"type": "string"},
                    "quantity": {"type": "integer", "minimum": 1},
                    "price": {"type": "number", "minimum": 0}
                },
                "required": ["product_id", "quantity", "price"]
            }
        ))
```

## Advanced Field Types

### Union Types

Handle fields that can be multiple types:

```python
from typing import Union, List

# Using Python type hints (recommended)
@flypy.function(name="flexible-input")
def flexible_input(data: Union[str, int, List[str]]) -> dict:
    """Accept string, int, or list of strings."""
    if isinstance(data, str):
        return {"type": "string", "value": data}
    elif isinstance(data, int):
        return {"type": "number", "value": data}
    elif isinstance(data, list):
        return {"type": "array", "value": data}
    else:
        raise ValueError("Unsupported type")

# Manual schema for complex unions
flexible_schema = Schema("Flexible Input")
flexible_schema.add_field("data", Field(
    "string",  # Default type, but we'll validate in function
    description="Can be string, number, or array"
))
```

### Recursive Schemas

Schemas that reference themselves:

```python
class TreeNodeSchema(Schema):
    """Schema for tree node with recursive children."""

    def __init__(self):
        super().__init__("Tree Node")

        self.add_field("id", Field("string", required=True))
        self.add_field("name", Field("string", required=True))
        self.add_field("value", Field("number", required=False))

        # Recursive reference to children
        self.add_field("children", Field(
            "array",
            required=False,
            items=self.to_dict()  # Reference to this schema
        ))

# Usage
@flypy.input_schema(TreeNodeSchema())
@flypy.function(name="process-tree")
def process_tree(node: dict) -> dict:
    """Process a tree structure recursively."""

    def count_nodes(n):
        count = 1  # Current node
        for child in n.get("children", []):
            count += count_nodes(child)
        return count

    node_count = count_nodes(node)

    return {
        "node_count": node_count,
        "root_name": node["name"]
    }
```

### Polymorphic Schemas

Handle different object types in the same field:

```python
def create_event_schema() -> Schema:
    """Schema for different types of events."""
    schema = Schema("Event")

    schema.add_field("event_type", Field(
        "string",
        enum=["user_action", "system_event", "error"],
        required=True
    ))

    # Base fields
    schema.add_field("timestamp", Field("string", required=True))
    schema.add_field("source", Field("string", required=True))

    # Type-specific fields (conditionally required)
    schema.add_field("user_id", Field("string", required=False))  # For user_action
    schema.add_field("action", Field("string", required=False))   # For user_action
    schema.add_field("component", Field("string", required=False))  # For system_event
    schema.add_field("severity", Field("string", enum=["low", "medium", "high"], required=False))  # For error
    schema.add_field("error_message", Field("string", required=False))  # For error

    return schema

@flypy.input_schema(create_event_schema())
@flypy.function(name="log-event")
def log_event(event: dict) -> dict:
    """Log different types of events with appropriate validation."""

    event_type = event["event_type"]

    # Validate type-specific fields
    if event_type == "user_action":
        if not event.get("user_id"):
            raise ValueError("user_id is required for user_action events")
        if not event.get("action"):
            raise ValueError("action is required for user_action events")

    elif event_type == "system_event":
        if not event.get("component"):
            raise ValueError("component is required for system_event events")

    elif event_type == "error":
        if not event.get("severity"):
            raise ValueError("severity is required for error events")
        if not event.get("error_message"):
            raise ValueError("error_message is required for error events")

    # Process event
    return {
        "logged": True,
        "event_id": f"{event_type}_{event['timestamp']}",
        "event_type": event_type
    }
```

## Schema Performance Optimization

### Schema Caching

Cache compiled schemas to avoid recompilation:

```python
import functools
from flypy import Schema

@functools.lru_cache(maxsize=128)
def get_cached_schema(schema_type: str) -> Schema:
    """Get cached schema instance."""
    if schema_type == "user":
        schema = Schema("User")
        schema.add_field("id", Field("string", required=True))
        schema.add_field("email", Field("string", required=True))
        # ... more fields
        return schema
    # ... other schema types

# Usage
@flypy.input_schema(get_cached_schema("user"))
@flypy.function(name="cached-schema-function")
def cached_schema_function(data: dict) -> dict:
    pass
```

### Lazy Schema Compilation

Defer schema compilation until needed:

```python
class LazySchema:
    """Lazy schema that compiles only when accessed."""

    def __init__(self, schema_factory):
        self._schema_factory = schema_factory
        self._compiled_schema = None

    @property
    def schema(self) -> Schema:
        if self._compiled_schema is None:
            self._compiled_schema = self._schema_factory()
        return self._compiled_schema

# Usage
def create_complex_schema():
    schema = Schema("Complex Schema")
    # Expensive schema construction
    for i in range(100):
        schema.add_field(f"field_{i}", Field("string", required=True))
    return schema

lazy_schema = LazySchema(create_complex_schema)

@flypy.input_schema(lazy_schema.schema)  # Only compiled when first accessed
@flypy.function(name="lazy-schema-function")
def lazy_schema_function(data: dict) -> dict:
    pass
```

### Schema Fragmentation

Split large schemas into smaller, reusable fragments:

```python
class SchemaFragments:
    """Reusable schema fragments."""

    @staticmethod
    def personal_info(required: bool = True):
        """Personal information fields."""
        return {
            "first_name": Field("string", required=required),
            "last_name": Field("string", required=required),
            "email": Field("string", required=required),
            "phone": Field("string", required=False)
        }

    @staticmethod
    def address_info(prefix: str = "", required: bool = True):
        """Address fields with optional prefix."""
        return {
            f"{prefix}street": Field("string", required=required),
            f"{prefix}city": Field("string", required=required),
            f"{prefix}state": Field("string", required=required),
            f"{prefix}zip_code": Field("string", required=required)
        }

def create_customer_schema() -> Schema:
    """Create customer schema from fragments."""
    schema = Schema("Customer")

    # Add personal info
    for field_name, field in SchemaFragments.personal_info().items():
        schema.add_field(field_name, field)

    # Add address info
    for field_name, field in SchemaFragments.address_info().items():
        schema.add_field(field_name, field)

    # Add customer-specific fields
    schema.add_field("customer_id", Field("string", required=True))
    schema.add_field("account_type", Field("string", enum=["basic", "premium"], required=True))

    return schema
```

## Custom Validators

### Field-Level Validators

Create custom validation functions for fields:

```python
from flypy import Field
import re

class ValidatedField(Field):
    """Field with custom validation."""

    def __init__(self, validator=None, **kwargs):
        super().__init__(**kwargs)
        self.validator = validator

    def validate(self, value):
        """Validate field value."""
        # Run standard JSON schema validation first
        super().validate(value)

        # Run custom validator
        if self.validator:
            self.validator(value)

# Custom validators
def validate_credit_card(number: str):
    """Validate credit card number using Luhn algorithm."""
    if not number:
        return

    # Remove spaces and dashes
    number = re.sub(r'[\s-]', '', number)

    if not number.isdigit():
        raise ValueError("Credit card number must contain only digits")

    if len(number) < 13 or len(number) > 19:
        raise ValueError("Credit card number must be 13-19 digits long")

    # Luhn algorithm
    total = 0
    reverse_digits = number[::-1]

    for i, digit in enumerate(reverse_digits):
        d = int(digit)
        if i % 2 == 1:
            d *= 2
            if d > 9:
                d -= 9
        total += d

    if total % 10 != 0:
        raise ValueError("Invalid credit card number")

def validate_password_strength(password: str):
    """Validate password strength."""
    if len(password) < 8:
        raise ValueError("Password must be at least 8 characters long")

    if not re.search(r'[A-Z]', password):
        raise ValueError("Password must contain at least one uppercase letter")

    if not re.search(r'[a-z]', password):
        raise ValueError("Password must contain at least one lowercase letter")

    if not re.search(r'[0-9]', password):
        raise ValueError("Password must contain at least one digit")

# Usage
@flypy.function(name="create-payment-method")
def create_payment_method(payment_data: dict) -> dict:
    # Note: In practice, you'd use the ValidatedField in schema definitions
    card_number = payment_data.get("card_number")
    if card_number:
        validate_credit_card(card_number)

    password = payment_data.get("password")
    if password:
        validate_password_strength(password)

    return {"created": True}
```

### Schema-Level Validators

Validate relationships between fields:

```python
def validate_date_range(data: dict):
    """Validate that start_date is before end_date."""
    start_date = data.get("start_date")
    end_date = data.get("end_date")

    if start_date and end_date:
        from datetime import datetime
        start = datetime.fromisoformat(start_date)
        end = datetime.fromisoformat(end_date)

        if start >= end:
            raise ValueError("start_date must be before end_date")

def validate_age_restrictions(data: dict):
    """Validate age restrictions for content."""
    content_rating = data.get("content_rating")
    min_age = data.get("min_age")

    rating_age_map = {
        "G": 0,
        "PG": 8,
        "PG-13": 13,
        "R": 17,
        "NC-17": 18
    }

    if content_rating and content_rating in rating_age_map:
        expected_min_age = rating_age_map[content_rating]
        if min_age is not None and min_age < expected_min_age:
            raise ValueError(f"Content rating {content_rating} requires minimum age {expected_min_age}")

@flypy.function(name="create-event")
def create_event(event_data: dict) -> dict:
    """Create event with cross-field validation."""

    # Run cross-field validations
    validate_date_range(event_data)
    validate_age_restrictions(event_data)

    return {
        "event_id": f"event_{hash(str(event_data))}",
        "created": True
    }
```

This advanced schema customization guide shows how to create robust, reusable, and performant schemas for complex FlyPy applications. The techniques demonstrated here can significantly improve code maintainability and validation reliability.