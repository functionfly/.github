# Tutorial: E-commerce Order Processing

This tutorial shows how to build a complete e-commerce order processing system using FlyPy, including order validation, inventory management, and payment processing.

## Overview

We'll build several FlyPy functions that work together to process e-commerce orders:

1. **Order Validation** - Validate order data and customer information
2. **Inventory Check** - Verify product availability
3. **Tax Calculation** - Calculate taxes based on location
4. **Payment Processing** - Process payment (simulated)
5. **Order Fulfillment** - Create order confirmation and update inventory

## Prerequisites

```bash
pip install flypy
```

## Step 1: Define Data Models

First, let's create the data models for our e-commerce system:

```python
# models.py
from typing import List, Dict, Any, Optional
from pydantic import BaseModel
from enum import Enum

class OrderStatus(str, Enum):
    PENDING = "pending"
    CONFIRMED = "confirmed"
    PROCESSING = "processing"
    SHIPPED = "shipped"
    DELIVERED = "delivered"
    CANCELLED = "cancelled"

class Address(BaseModel):
    street: str
    city: str
    state: str
    zip_code: str
    country: str

class Customer(BaseModel):
    id: str
    name: str
    email: str
    shipping_address: Address
    billing_address: Optional[Address] = None

class OrderItem(BaseModel):
    product_id: str
    name: str
    price: float
    quantity: int
    sku: str

class Order(BaseModel):
    id: str
    customer: Customer
    items: List[OrderItem]
    subtotal: float
    tax_rate: float = 0.08
    shipping_cost: float = 9.99
    total: float
    status: OrderStatus = OrderStatus.PENDING
    created_at: str
```

## Step 2: Order Validation Function

Let's create a function to validate incoming orders:

```python
# order_validation.py
import flypy
from typing import Dict, Any, List
from models import Order, Customer, OrderItem, Address

@flypy.function(
    name="validate-order",
    description="Validate order data and customer information",
    deterministic=True,
    idempotent=True
)
def validate_order(order_data: Dict[str, Any]) -> Dict[str, Any]:
    """
    Validate an incoming order.

    Args:
        order_data: Raw order data from the client

    Returns:
        Validation result with order data or errors
    """
    errors = []

    # Validate required fields
    if not order_data.get("customer"):
        errors.append("Customer information is required")
        return {"valid": False, "errors": errors}

    customer_data = order_data["customer"]
    if not customer_data.get("id"):
        errors.append("Customer ID is required")
    if not customer_data.get("name"):
        errors.append("Customer name is required")
    if not customer_data.get("email"):
        errors.append("Customer email is required")
    if not customer_data.get("shipping_address"):
        errors.append("Shipping address is required")

    # Validate shipping address
    shipping_addr = customer_data.get("shipping_address", {})
    required_addr_fields = ["street", "city", "state", "zip_code", "country"]
    for field in required_addr_fields:
        if not shipping_addr.get(field):
            errors.append(f"Shipping address {field} is required")

    # Validate items
    items = order_data.get("items", [])
    if not items:
        errors.append("Order must contain at least one item")

    for i, item in enumerate(items):
        if not item.get("product_id"):
            errors.append(f"Item {i+1}: product_id is required")
        if not item.get("name"):
            errors.append(f"Item {i+1}: name is required")
        if not isinstance(item.get("price"), (int, float)) or item["price"] <= 0:
            errors.append(f"Item {i+1}: price must be a positive number")
        if not isinstance(item.get("quantity"), int) or item["quantity"] <= 0:
            errors.append(f"Item {i+1}: quantity must be a positive integer")

    if errors:
        return {"valid": False, "errors": errors}

    # Calculate totals
    subtotal = sum(item["price"] * item["quantity"] for item in items)
    tax_rate = order_data.get("tax_rate", 0.08)
    shipping_cost = order_data.get("shipping_cost", 9.99)
    tax = subtotal * tax_rate
    total = subtotal + tax + shipping_cost

    return {
        "valid": True,
        "order": {
            **order_data,
            "subtotal": subtotal,
            "tax": tax,
            "total": total,
            "status": "pending"
        }
    }
```

## Step 3: Inventory Management Function

Create a function to check product availability:

```python
# inventory.py
import flypy
from typing import Dict, Any, List

# Mock inventory data (in real app, this would come from a database)
INVENTORY = {
    "PROD-001": {"name": "Wireless Headphones", "stock": 50, "reserved": 0},
    "PROD-002": {"name": "Bluetooth Speaker", "stock": 25, "reserved": 2},
    "PROD-003": {"name": "USB Cable", "stock": 100, "reserved": 5},
    "PROD-004": {"name": "Phone Case", "stock": 75, "reserved": 10},
}

@flypy.function(
    name="check-inventory",
    description="Check product availability and reserve items",
    deterministic=False,  # Uses external inventory data
    capabilities=["database"]
)
def check_inventory(order_data: Dict[str, Any]) -> Dict[str, Any]:
    """
    Check inventory availability for order items.

    Args:
        order_data: Validated order data

    Returns:
        Inventory check result
    """
    items = order_data.get("items", [])
    unavailable_items = []
    total_available = True

    for item in items:
        product_id = item["product_id"]
        quantity_requested = item["quantity"]

        if product_id not in INVENTORY:
            unavailable_items.append({
                "product_id": product_id,
                "reason": "Product not found"
            })
            total_available = False
            continue

        inventory_item = INVENTORY[product_id]
        available_stock = inventory_item["stock"] - inventory_item["reserved"]

        if available_stock < quantity_requested:
            unavailable_items.append({
                "product_id": product_id,
                "requested": quantity_requested,
                "available": available_stock,
                "reason": "Insufficient stock"
            })
            total_available = False

    return {
        "available": total_available,
        "unavailable_items": unavailable_items,
        "inventory_checked": True
    }
```

## Step 4: Tax Calculation Function

Create a function to calculate taxes based on shipping location:

```python
# tax_calculation.py
import flypy
from typing import Dict, Any

# Mock tax rates by state (in real app, this would be more comprehensive)
TAX_RATES = {
    "CA": 0.0825,  # California
    "NY": 0.04,    # New York
    "TX": 0.0625,  # Texas
    "FL": 0.06,    # Florida
    "WA": 0.065,   # Washington
    # Default rate for other states
    "default": 0.08
}

@flypy.function(
    name="calculate-tax",
    description="Calculate tax based on shipping location",
    deterministic=True,
    pure=True,
    cache_ttl=86400  # Cache for 24 hours
)
def calculate_tax(order_data: Dict[str, Any]) -> Dict[str, float]:
    """
    Calculate tax for an order based on shipping address.

    Args:
        order_data: Order data with shipping address

    Returns:
        Tax calculation result
    """
    shipping_address = order_data.get("customer", {}).get("shipping_address", {})
    state = shipping_address.get("state", "")

    # Get tax rate for the state
    tax_rate = TAX_RATES.get(state, TAX_RATES["default"])

    # Calculate tax on subtotal only (shipping is usually not taxed)
    subtotal = order_data.get("subtotal", 0)
    tax_amount = subtotal * tax_rate

    return {
        "tax_rate": tax_rate,
        "tax_amount": round(tax_amount, 2),
        "subtotal": subtotal,
        "state": state
    }
```

## Step 5: Payment Processing Function

Create a simulated payment processing function:

```python
# payment.py
import flypy
from typing import Dict, Any
import hashlib
import time

@flypy.function(
    name="process-payment",
    description="Process payment for an order",
    deterministic=False,  # External payment service
    capabilities=["network", "payment"]
)
def process_payment(order_data: Dict[str, Any], payment_info: Dict[str, Any]) -> Dict[str, Any]:
    """
    Process payment for an order.

    Args:
        order_data: Validated order data
        payment_info: Payment method information

    Returns:
        Payment processing result
    """
    # Validate payment info
    if not payment_info.get("card_number"):
        return {"success": False, "error": "Card number is required"}

    if not payment_info.get("expiry_date"):
        return {"success": False, "error": "Expiry date is required"}

    if not payment_info.get("cvv"):
        return {"success": False, "error": "CVV is required"}

    # Simulate payment processing
    # In real app, this would call a payment service like Stripe
    order_id = order_data.get("id", "")
    amount = order_data.get("total", 0)

    # Simple validation - check if card number is 16 digits
    card_number = payment_info["card_number"].replace(" ", "").replace("-", "")
    if len(card_number) != 16 or not card_number.isdigit():
        return {"success": False, "error": "Invalid card number"}

    # Simulate processing delay
    time.sleep(0.1)  # Simulate network call

    # Generate transaction ID
    transaction_id = hashlib.md5(f"{order_id}-{amount}-{time.time()}".encode()).hexdigest()[:16]

    return {
        "success": True,
        "transaction_id": transaction_id,
        "amount_charged": amount,
        "currency": "USD",
        "processed_at": time.time(),
        "payment_method": "credit_card"
    }
```

## Step 6: Order Fulfillment Function

Create the main order processing function that orchestrates everything:

```python
# order_fulfillment.py
import flypy
from typing import Dict, Any, List
import time

# Import our other functions
from order_validation import validate_order
from inventory import check_inventory
from tax_calculation import calculate_tax
from payment import process_payment

@flypy.function(
    name="process-order",
    description="Complete order processing workflow",
    deterministic=False,  # Orchestrates multiple non-deterministic functions
    max_execution_time=30000  # 30 seconds
)
def process_order(raw_order: Dict[str, Any], payment_info: Dict[str, Any]) -> Dict[str, Any]:
    """
    Process a complete order from validation to fulfillment.

    Args:
        raw_order: Raw order data from client
        payment_info: Payment method information

    Returns:
        Order processing result
    """
    # Step 1: Validate order
    validation_result = validate_order(raw_order)
    if not validation_result["valid"]:
        return {
            "success": False,
            "stage": "validation",
            "errors": validation_result["errors"]
        }

    order_data = validation_result["order"]

    # Step 2: Check inventory
    inventory_result = check_inventory(order_data)
    if not inventory_result["available"]:
        return {
            "success": False,
            "stage": "inventory",
            "errors": ["Some items are not available"],
            "unavailable_items": inventory_result["unavailable_items"]
        }

    # Step 3: Calculate tax
    tax_result = calculate_tax(order_data)
    order_data["tax_rate"] = tax_result["tax_rate"]
    order_data["tax"] = tax_result["tax_amount"]
    order_data["total"] = order_data["subtotal"] + order_data["tax"] + order_data.get("shipping_cost", 0)

    # Step 4: Process payment
    payment_result = process_payment(order_data, payment_info)
    if not payment_result["success"]:
        return {
            "success": False,
            "stage": "payment",
            "errors": [payment_result["error"]]
        }

    # Step 5: Create order confirmation
    order_id = f"ORD-{int(time.time())}-{hash(str(order_data)) % 10000:04d}"

    confirmation = {
        "order_id": order_id,
        "status": "confirmed",
        "customer": order_data["customer"],
        "items": order_data["items"],
        "pricing": {
            "subtotal": order_data["subtotal"],
            "tax": order_data["tax"],
            "tax_rate": order_data["tax_rate"],
            "shipping": order_data.get("shipping_cost", 0),
            "total": order_data["total"]
        },
        "payment": {
            "transaction_id": payment_result["transaction_id"],
            "amount": payment_result["amount_charged"],
            "currency": payment_result["currency"]
        },
        "timestamps": {
            "ordered_at": time.time(),
            "estimated_delivery": time.time() + (7 * 24 * 60 * 60)  # 7 days
        }
    }

    return {
        "success": True,
        "order": confirmation,
        "message": f"Order {order_id} has been successfully processed"
    }
```

## Step 7: Testing and Deployment

Create a test script to verify our functions work correctly:

```python
# test_order_processing.py
import json
from order_fulfillment import process_order

# Test data
test_order = {
    "customer": {
        "id": "CUST-001",
        "name": "John Doe",
        "email": "john@example.com",
        "shipping_address": {
            "street": "123 Main St",
            "city": "San Francisco",
            "state": "CA",
            "zip_code": "94105",
            "country": "USA"
        }
    },
    "items": [
        {
            "product_id": "PROD-001",
            "name": "Wireless Headphones",
            "price": 99.99,
            "quantity": 1,
            "sku": "WH-001"
        },
        {
            "product_id": "PROD-003",
            "name": "USB Cable",
            "price": 12.99,
            "quantity": 2,
            "sku": "USB-001"
        }
    ]
}

test_payment = {
    "card_number": "4111111111111111",
    "expiry_date": "12/25",
    "cvv": "123"
}

if __name__ == "__main__":
    # Test the order processing
    result = process_order(test_order, test_payment)

    print("Order Processing Result:")
    print(json.dumps(result, indent=2, default=str))
```

## Step 8: Build and Deploy

Build the FlyPy functions:

```bash
# Build all functions
flypy build order_validation.py inventory.py tax_calculation.py payment.py order_fulfillment.py

# List all built functions
flypy list

# Test locally (optional)
flypy local order_fulfillment.py process-order --port 8080

# Deploy to FunctionFly
flypy deploy ./dist/process-order --token YOUR_TOKEN --app-id YOUR_APP_ID
```

## Step 9: API Usage

Once deployed, you can call your order processing API:

```python
import requests

# Example API call
order_data = {
    "customer": {
        "id": "CUST-123",
        "name": "Jane Smith",
        "email": "jane@example.com",
        "shipping_address": {
            "street": "456 Oak Ave",
            "city": "Austin",
            "state": "TX",
            "zip_code": "78701",
            "country": "USA"
        }
    },
    "items": [
        {
            "product_id": "PROD-002",
            "name": "Bluetooth Speaker",
            "price": 79.99,
            "quantity": 1,
            "sku": "SPKR-001"
        }
    ]
}

payment_info = {
    "card_number": "4111111111111111",
    "expiry_date": "12/25",
    "cvv": "123"
}

# Call the deployed function
response = requests.post(
    "https://your-functionfly-app.com/process-order",
    json={"order": order_data, "payment": payment_info}
)

result = response.json()
print(f"Order processed: {result['success']}")
if result['success']:
    print(f"Order ID: {result['order']['order_id']}")
```

## Next Steps

This tutorial covered the basics of building an e-commerce order processing system with FlyPy. You can extend this further by:

1. **Adding more validation rules** - Email format validation, address verification
2. **Implementing real inventory management** - Database integration, stock reservations
3. **Adding shipping calculations** - Weight-based shipping, carrier integration
4. **Implementing order tracking** - Status updates, shipment tracking
5. **Adding error handling and retries** - Failed payment recovery, inventory sync

The modular function design makes it easy to test, maintain, and scale individual components of your order processing system.