# FlyPy Python SDK

A Python SDK for compiling deterministic Python functions to WebAssembly for execution on the FunctionFly platform.

## Installation

```bash
pip install flypy
```

## Quick Start

### Basic Usage

```python
import flypy

@flypy.function(
    name="calculate-total",
    deterministic=True,
    idempotent=True,
    cache_ttl=3600
)
def handler(event):
    """Calculate order total with tax."""
    items = event.get("items", [])
    tax_rate = event.get("tax_rate", 0.08)

    subtotal = sum(item["price"] * item["quantity"] for item in items)
    tax = subtotal * tax_rate
    total = subtotal + tax

    return {
        "subtotal": subtotal,
        "tax": tax,
        "total": total
    }
```

### With Explicit Schemas

```python
import flypy

@flypy.input_schema({
    "type": "object",
    "properties": {
        "items": {
            "type": "array",
            "items": {
                "type": "object",
                "properties": {
                    "price": {"type": "number"},
                    "quantity": {"type": "integer"}
                },
                "required": ["price", "quantity"]
            }
        },
        "tax_rate": {"type": "number", "default": 0.08}
    },
    "required": ["items"]
})
@flypy.output_schema({
    "type": "object",
    "properties": {
        "subtotal": {"type": "number"},
        "tax": {"type": "number"},
        "total": {"type": "number"}
    },
    "required": ["subtotal", "tax", "total"]
})
@flypy.function(name="calculate-total")
def handler(event):
    # Function implementation...
    pass
```

### With Type Hints (Automatic Schema Inference)

```python
from typing import List, Dict, Any
import flypy

@flypy.function(name="process-data")
def process_data(items: List[Dict[str, Any]], config: Dict[str, str]) -> Dict[str, Any]:
    """Process a list of items with configuration."""
    processed = []
    for item in items:
        # Process each item
        processed_item = {
            "id": item["id"],
            "processed_at": config.get("timestamp", "now"),
            "value": item.get("value", 0) * 2
        }
        processed.append(processed_item)

    return {
        "processed_items": processed,
        "total_count": len(processed)
    }
```

## CLI Usage

### Build Functions

```bash
# Build all functions in a file
flypy build handler.py

# Build with verbose output
flypy build handler.py --verbose

# Build in compatible mode (allows some non-deterministic operations)
flypy build handler.py --mode compatible

# Specify output directory
flypy build handler.py --output ./build
```

### List Functions

```bash
# List all functions in Python files
flypy list handler.py utils.py

# List functions in current directory
flypy list
```

### Local Development

```bash
# Run functions locally for testing
flypy local handler.py calculate-total --port 8080
```

### Deploy Functions

```bash
# Deploy built artifacts
flypy deploy ./dist/calculate-total
```

## Function Decorators

### @flypy.function

The main decorator that marks a function as a FlyPy function.

**Parameters:**
- `name` (str, optional): Function name (defaults to function.__name__)
- `version` (str): Semantic version string (default: "1.0.0")
- `description` (str, optional): Function description
- `deterministic` (bool): Whether function is deterministic (default: True)
- `idempotent` (bool): Whether function is idempotent (default: False)
- `pure` (bool): Whether function is pure (default: False)
- `cache_ttl` (int, optional): Cache TTL in seconds
- `capabilities` (List[str]): Required capabilities
- `max_execution_time` (int, optional): Max execution time in milliseconds
- `execution_mode` (ExecutionMode): Execution mode (default: DETERMINISTIC)

### @flypy.input_schema

Specify the input schema for a function.

**Parameters:**
- `schema` (dict or Schema): JSON schema for function input

### @flypy.output_schema

Specify the output schema for a function.

**Parameters:**
- `schema` (dict or Schema): JSON schema for function output

## Schema System

FlyPy includes a powerful schema system for automatic type inference and validation.

### Manual Schema Creation

```python
from flypy import Schema, Field

# Create input schema
input_schema = Schema("Order Input")
input_schema.add_field("customer_id", Field("string", required=True))
input_schema.add_field("items", Field("array", required=True, items={
    "type": "object",
    "properties": {
        "product_id": {"type": "string"},
        "quantity": {"type": "integer", "minimum": 1}
    }
}))

# Use with decorator
@flypy.input_schema(input_schema)
@flypy.function(name="process-order")
def handler(event):
    pass
```

### Type Hint Inference

FlyPy can automatically infer schemas from Python type hints:

```python
from typing import List, Dict, Optional
from pydantic import BaseModel

class Item(BaseModel):
    product_id: str
    quantity: int
    price: float

@flypy.function(name="calculate-total")
def handler(items: List[Item], discount: Optional[float] = None) -> Dict[str, float]:
    # FlyPy will automatically infer schemas from type hints
    pass
```

## Capabilities

Functions can declare required capabilities for security and access control:

```python
@flypy.function(
    name="fetch-user-data",
    capabilities=["network", "database"]
)
def handler(event):
    # This function can make network requests and access databases
    pass
```

## Execution Modes

### Deterministic Mode (Default)

- Full determinism verification
- Restricted Python subset
- Maximum security and replayability
- Best for business logic and calculations

### Compatible Mode

- Allows some non-deterministic operations
- More flexible Python usage
- Still provides some guarantees
- Good for transitional code

## Building and Deployment

### Development Workflow

1. **Write your function** with FlyPy decorators
2. **Test locally** using `flypy local`
3. **Build artifacts** using `flypy build`
4. **Verify artifacts** using `flypy verify`
5. **Deploy to FunctionFly** using `flypy deploy`

### CI/CD Integration

```yaml
# .github/workflows/deploy.yml
name: Deploy FlyPy Functions
on: [push]

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - uses: actions/setup-python@v2
        with:
          python-version: '3.9'
      - run: pip install flypy
      - run: flypy build functions/
      - run: flypy deploy ./dist/ --token ${{ secrets.FF_TOKEN }}
```

## Best Practices

### Deterministic Functions

1. **Avoid non-deterministic operations:**
   - No `random.*`, `time.time()`, `uuid.uuid4()`
   - No file system access
   - No network requests (unless capability declared)

2. **Use pure functions when possible:**
   ```python
   @flypy.function(pure=True, idempotent=True)
   def add_tax(amount: float, rate: float) -> float:
       return amount * (1 + rate)
   ```

3. **Declare capabilities explicitly:**
   ```python
   @flypy.function(capabilities=["network:read"])
   def fetch_data(url: str):
       # This function can make HTTP requests
       pass
   ```

### Schema Design

1. **Use specific types:**
   ```python
   # Good
   def process(items: List[Dict[str, Union[str, int]]]) -> Dict[str, Any]:
       pass

   # Better
   from typing import TypedDict
   class Item(TypedDict):
       name: str
       quantity: int

   def process(items: List[Item]) -> Dict[str, Any]:
       pass
   ```

2. **Validate input thoroughly:**
   ```python
   @flypy.function(name="validate-order")
   def handler(order: Dict[str, Any]) -> Dict[str, Any]:
       if not order.get("customer_id"):
           raise ValueError("customer_id is required")
       # ... more validation
       return {"valid": True}
   ```

## Error Handling

FlyPy provides clear error messages for common issues:

```bash
# Compilation errors
flypy build handler.py
# Error: Function uses forbidden builtin 'open'

# Schema validation errors
flypy build handler.py
# Error: Input schema validation failed: 'price' is required
```

## Advanced Usage

### Custom Schema Validation

```python
from flypy import Schema, validate_schema

@flypy.function(name="custom-validation")
def handler(data: Dict[str, Any]) -> Dict[str, Any]:
    # Custom validation logic
    errors = validate_schema(data, my_custom_schema)
    if errors:
        raise ValueError(f"Validation failed: {errors}")
    return {"processed": True}
```

### Function Composition

```python
@flypy.function(name="calculate-subtotal", pure=True)
def calculate_subtotal(items: List[Dict[str, Any]]) -> float:
    return sum(item["price"] * item["quantity"] for item in items)

@flypy.function(name="calculate-tax", pure=True)
def calculate_tax(subtotal: float, rate: float) -> float:
    return subtotal * rate

@flypy.function(name="calculate-total")
def calculate_total(items: List[Dict[str, Any]], tax_rate: float) -> Dict[str, float]:
    subtotal = calculate_subtotal(items)
    tax = calculate_tax(subtotal, tax_rate)
    return {
        "subtotal": subtotal,
        "tax": tax,
        "total": subtotal + tax
    }
```

## Troubleshooting

### Common Issues

1. **"Function uses forbidden builtin"**
   - Solution: Remove use of restricted operations like `open()`, `eval()`, etc.

2. **"Schema validation failed"**
   - Solution: Ensure your function input/output matches the declared schema

3. **"Determinism verification failed"**
   - Solution: Remove non-deterministic operations or switch to compatible mode

4. **"Go binary not found"**
   - Solution: Ensure FlyPy Go compiler is installed and in PATH

### Getting Help

- Check the [FunctionFly documentation](https://docs.functionfly.com)
- Report issues on [GitHub](https://github.com/functionfly/functionfly)
- Join our [Discord community](https://discord.gg/functionfly)

## Contributing

We welcome contributions! Please see our [contributing guide](../CONTRIBUTING.md) for details.

## License

MIT License - see [LICENSE](../LICENSE) file for details.