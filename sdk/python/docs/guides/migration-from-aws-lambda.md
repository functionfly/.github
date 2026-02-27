# Migration Guide: From AWS Lambda to FlyPy

This guide helps you migrate your AWS Lambda functions to FlyPy, taking advantage of FlyPy's deterministic compilation, better performance, and simplified deployment.

## Overview

FlyPy offers several advantages over AWS Lambda:

- **Deterministic execution** - Same input always produces same output
- **WebAssembly compilation** - Faster cold starts and consistent performance
- **Simplified deployment** - Single CLI command vs complex CloudFormation/SAM
- **Better debugging** - Clear error messages and built-in validation
- **Reduced costs** - More efficient execution and better resource utilization

## Migration Steps

### Step 1: Analyze Your Lambda Functions

First, understand what your Lambda functions do:

```python
# AWS Lambda function
def lambda_handler(event, context):
    # Your code here
    return {
        'statusCode': 200,
        'body': json.dumps('Hello from Lambda!')
    }
```

Key things to identify:
- **Event sources** (API Gateway, S3, DynamoDB, etc.)
- **Environment variables** used
- **External dependencies** (DynamoDB, S3, etc.)
- **IAM permissions** required
- **Runtime** (Python version, layers, etc.)

### Step 2: Set Up FlyPy Project Structure

Create a FlyPy project structure:

```bash
# Create project directory
mkdir my-flypy-app
cd my-flypy-app

# Initialize with requirements
pip install flypy

# Create function files
touch functions.py
```

### Step 3: Convert Lambda Handler to FlyPy Function

#### Basic Handler Conversion

```python
# AWS Lambda (OLD)
def lambda_handler(event, context):
    name = event.get('name', 'World')
    return {
        'statusCode': 200,
        'body': json.dumps(f'Hello {name}!')
    }
```

```python
# FlyPy (NEW)
import flypy

@flypy.function(
    name="hello-handler",
    description="Simple greeting handler"
)
def hello_handler(event: dict) -> dict:
    name = event.get('name', 'World')
    return {
        'message': f'Hello {name}!',
        'timestamp': event.get('timestamp')
    }
```

#### Key Differences

1. **No context parameter** - FlyPy doesn't use Lambda's context object
2. **Direct return** - Return data directly, not wrapped in HTTP response
3. **Type hints** - Add type hints for better validation
4. **Decorators** - Use `@flypy.function` decorator

### Step 4: Handle Different Event Sources

#### API Gateway Events

```python
# AWS Lambda (OLD)
def lambda_handler(event, context):
    # Extract from API Gateway event
    path = event['path']
    method = event['httpMethod']
    body = json.loads(event.get('body', '{}'))
    headers = event.get('headers', {})

    # Your logic here
    return {
        'statusCode': 200,
        'body': json.dumps({'result': 'success'})
    }
```

```python
# FlyPy (NEW)
import flypy

@flypy.input_schema({
    "type": "object",
    "properties": {
        "path": {"type": "string"},
        "method": {"type": "string"},
        "body": {"type": "object"},
        "headers": {"type": "object"}
    },
    "required": ["path", "method"]
})
@flypy.output_schema({
    "type": "object",
    "properties": {
        "statusCode": {"type": "integer"},
        "body": {"type": "object"},
        "headers": {"type": "object"}
    }
})
@flypy.function(
    name="api-handler",
    description="Handle API Gateway requests"
)
def api_handler(request: dict) -> dict:
    path = request['path']
    method = request['method']
    body = request.get('body', {})
    headers = request.get('headers', {})

    # Your logic here
    return {
        'statusCode': 200,
        'body': {'result': 'success'},
        'headers': {'Content-Type': 'application/json'}
    }
```

#### DynamoDB Stream Events

```python
# AWS Lambda (OLD)
def lambda_handler(event, context):
    for record in event['Records']:
        if record['eventName'] == 'INSERT':
            new_item = record['dynamodb']['NewImage']
            # Process new item
        elif record['eventName'] == 'MODIFY':
            # Process modification
            pass
```

```python
# FlyPy (NEW)
import flypy

@flypy.function(
    name="dynamodb-processor",
    description="Process DynamoDB stream events",
    capabilities=["database"]
)
def process_dynamodb_stream(event: dict) -> dict:
    processed_count = 0

    for record in event.get('Records', []):
        if record['eventName'] == 'INSERT':
            new_item = record['dynamodb']['NewImage']
            # Process new item
            processed_count += 1
        elif record['eventName'] == 'MODIFY':
            # Process modification
            processed_count += 1

    return {
        'processed_records': processed_count,
        'status': 'completed'
    }
```

#### S3 Events

```python
# AWS Lambda (OLD)
def lambda_handler(event, context):
    for record in event['Records']:
        bucket = record['s3']['bucket']['name']
        key = record['s3']['object']['key']
        # Process S3 object
```

```python
# FlyPy (NEW)
import flypy

@flypy.function(
    name="s3-processor",
    description="Process S3 object events",
    capabilities=["storage"]
)
def process_s3_event(event: dict) -> dict:
    processed_files = []

    for record in event.get('Records', []):
        bucket = record['s3']['bucket']['name']
        key = record['s3']['object']['key']

        # Process S3 object
        processed_files.append({
            'bucket': bucket,
            'key': key,
            'processed': True
        })

    return {
        'processed_files': processed_files,
        'total_count': len(processed_files)
    }
```

### Step 5: Handle Environment Variables and Configuration

```python
# AWS Lambda (OLD)
import os

def lambda_handler(event, context):
    api_key = os.environ['API_KEY']
    database_url = os.environ['DATABASE_URL']
    # Use configuration
```

```python
# FlyPy (NEW)
import flypy
import os

@flypy.function(
    name="config-handler",
    description="Handler that uses environment configuration"
)
def config_handler(event: dict) -> dict:
    # Environment variables work the same way
    api_key = os.environ.get('API_KEY')
    database_url = os.environ.get('DATABASE_URL')

    if not api_key:
        raise ValueError("API_KEY environment variable is required")

    # Your logic here
    return {
        'configured': True,
        'has_api_key': bool(api_key),
        'has_database_url': bool(database_url)
    }
```

### Step 6: Handle External Dependencies

#### DynamoDB Operations

```python
# AWS Lambda (OLD)
import boto3

dynamodb = boto3.resource('dynamodb')
table = dynamodb.Table('MyTable')

def lambda_handler(event, context):
    response = table.get_item(Key={'id': event['id']})
    return response.get('Item')
```

```python
# FlyPy (NEW)
import flypy
# Note: External dependencies like boto3 are not allowed in deterministic mode
# Use FunctionFly's built-in capabilities or HTTP APIs instead

@flypy.function(
    name="get-item",
    description="Get item from database",
    capabilities=["database", "network"]  # Declare required capabilities
)
def get_item(request: dict) -> dict:
    item_id = request.get('id')
    if not item_id:
        raise ValueError("id is required")

    # Use HTTP API or FunctionFly database capabilities
    # This is a simplified example - actual implementation depends on your setup

    return {
        'id': item_id,
        'found': True,
        'data': {'placeholder': 'data'}
    }
```

### Step 7: Error Handling

```python
# AWS Lambda (OLD)
def lambda_handler(event, context):
    try:
        # Your logic
        return {'statusCode': 200, 'body': 'success'}
    except ValueError as e:
        return {'statusCode': 400, 'body': str(e)}
    except Exception as e:
        return {'statusCode': 500, 'body': 'Internal server error'}
```

```python
# FlyPy (NEW)
import flypy

@flypy.function(
    name="error-handling-example",
    description="Example of proper error handling"
)
def error_handling_example(request: dict) -> dict:
    try:
        # Validate input
        if not request.get('required_field'):
            raise ValueError("required_field is missing")

        # Your logic here
        result = process_data(request)

        return {
            'success': True,
            'data': result
        }

    except ValueError as e:
        # Input validation errors
        raise ValueError(f"Invalid input: {str(e)}")

    except Exception as e:
        # Unexpected errors
        raise RuntimeError(f"Processing failed: {str(e)}")
```

### Step 8: Update IAM Permissions

FlyPy uses a different permission model. Instead of IAM roles, you declare capabilities:

```python
# FlyPy capabilities (in function decorators)
@flypy.function(
    name="my-function",
    capabilities=["database", "network", "storage"]  # Declare what you need
)
def my_function(request: dict) -> dict:
    # FunctionFly will ensure these capabilities are available
    pass
```

### Step 9: Build and Deploy

```bash
# Build functions
flypy build functions.py

# Deploy (much simpler than AWS SAM/CloudFormation)
flypy deploy ./dist/my-function --token YOUR_FF_TOKEN --app-id YOUR_APP_ID
```

### Step 10: Update API Gateway/Load Balancer

Instead of API Gateway, FlyPy functions are deployed directly to FunctionFly endpoints:

```python
# Old: API Gateway URL
# https://your-api-id.execute-api.region.amazonaws.com/stage/function

# New: FunctionFly URL
# https://your-app.functionfly.com/function-name
```

## Migration Checklist

- [ ] Identify all Lambda functions to migrate
- [ ] Analyze event sources and dependencies
- [ ] Convert function signatures and remove context parameter
- [ ] Add FlyPy decorators and type hints
- [ ] Handle external dependencies (may need to switch to HTTP APIs)
- [ ] Update error handling patterns
- [ ] Declare required capabilities
- [ ] Test functions locally with `flypy local`
- [ ] Build and deploy functions
- [ ] Update client applications to use new URLs
- [ ] Monitor and optimize performance

## Common Migration Issues

### 1. Non-Deterministic Operations

**Problem**: Lambda allows non-deterministic operations that FlyPy doesn't.

**Solution**: Use `execution_mode="compatible"` for transitional code:

```python
@flypy.function(
    name="transitional-function",
    execution_mode="compatible"  # Allows some non-deterministic operations
)
def transitional_function(event: dict) -> dict:
    # Can use random, time, etc.
    pass
```

### 2. Large Dependencies

**Problem**: Lambda layers vs FlyPy's minimal runtime.

**Solution**: FlyPy compiles to WebAssembly with minimal runtime. Large dependencies may need to be handled differently.

### 3. Cold Start Performance

**Problem**: Lambda cold starts can be slow.

**Solution**: FlyPy's WebAssembly compilation provides faster, more consistent cold starts.

### 4. Vendor Lock-in

**Problem**: Moving from AWS-specific services.

**Solution**: FlyPy is cloud-agnostic and can be deployed to multiple providers.

## Performance Comparison

| Metric | AWS Lambda | FlyPy |
|--------|------------|-------|
| Cold Start | 100-5000ms | 50-200ms (WebAssembly) |
| Runtime | Python interpreter | WebAssembly |
| Bundle Size | Up to 250MB | Optimized WASM |
| Determinism | No | Yes |
| Deployment | Complex (SAM/CF) | Simple CLI |

## Next Steps

1. **Start small** - Migrate one function at a time
2. **Test thoroughly** - Use `flypy local` for testing
3. **Monitor performance** - Compare latency and costs
4. **Gradual migration** - Run both systems in parallel initially
5. **Full migration** - Once confident, complete the migration

FlyPy provides better performance, simpler deployment, and more predictable execution than AWS Lambda, making it an excellent choice for modern serverless applications.