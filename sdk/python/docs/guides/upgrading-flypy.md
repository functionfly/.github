# Migration Guide: Upgrading FlyPy Versions

This guide helps you upgrade your FlyPy functions when new versions are released, ensuring compatibility and taking advantage of new features.

## Version Compatibility

FlyPy follows semantic versioning (MAJOR.MINOR.PATCH):

- **MAJOR**: Breaking changes that require code updates
- **MINOR**: New features, backward compatible
- **PATCH**: Bug fixes, backward compatible

## Upgrading from 0.x to 1.0.0

### Breaking Changes

#### 1. Function Decorator Changes

**Old syntax (0.x):**
```python
import flypy

@flypy.function("my-function")  # String parameter
def handler(event):
    pass
```

**New syntax (1.0.0):**
```python
import flypy

@flypy.function(name="my-function")  # Named parameter
def handler(event):
    pass
```

#### 2. Schema Decorator Changes

**Old syntax (0.x):**
```python
@flypy.schema({
    "type": "object",
    "properties": {...}
})
```

**New syntax (1.0.0):**
```python
@flypy.input_schema({...})
@flypy.output_schema({...})
```

#### 3. Import Changes

**Old imports (0.x):**
```python
from flypy import function, schema
```

**New imports (1.0.0):**
```python
from flypy import function, input_schema, output_schema, Schema, Field
```

### Migration Steps for 1.0.0

1. **Update function decorators:**
   ```python
   # Before
   @flypy.function("calculate-total")

   # After
   @flypy.function(name="calculate-total")
   ```

2. **Split schema decorators:**
   ```python
   # Before
   @flypy.schema({
       "input": {...},
       "output": {...}
   })

   # After
   @flypy.input_schema({...})
   @flypy.output_schema({...})
   ```

3. **Update imports:**
   ```python
   # Before
   from flypy import function, schema

   # After
   from flypy import function, input_schema, output_schema
   ```

## Upgrading from 1.0.x to 1.1.0

### New Features

#### 1. Automatic Schema Inference

You can now use type hints for automatic schema generation:

```python
from typing import List, Dict, Any
import flypy

@flypy.function(name="process-data")
def process_data(items: List[Dict[str, Any]], config: Dict[str, str]) -> Dict[str, Any]:
    # Schemas are automatically inferred from type hints
    pass
```

#### 2. Enhanced Error Messages

Better error messages with more context and suggestions.

#### 3. Performance Improvements

- Faster compilation times
- Smaller WebAssembly bundle sizes
- Improved cold start performance

### Migration Steps for 1.1.0

1. **Consider using type hints** for automatic schema inference instead of manual schemas
2. **Update any code that relied on old error message formats**
3. **Rebuild and redeploy** to take advantage of performance improvements

## Upgrading from 1.1.x to 1.2.0

### New Features

#### 1. Execution Modes

New execution modes for different use cases:

```python
@flypy.function(
    name="flexible-function",
    execution_mode="compatible"  # Allows some non-deterministic operations
)
def flexible_function(event: dict) -> dict:
    # Can use operations like random, time, etc.
    pass
```

#### 2. Enhanced Capabilities System

More granular capability declarations:

```python
@flypy.function(
    name="api-function",
    capabilities=["network:http", "database:read", "cache:read"]
)
def api_function(event: dict) -> dict:
    pass
```

#### 3. Improved CLI

New CLI commands and options:

```bash
# New commands
flypy verify ./dist/function  # Verify build artifacts
flypy local functions.py my-function --port 8080  # Local testing

# New options
flypy build --mode compatible  # Build in compatible mode
flypy build --verbose  # Verbose output
```

### Migration Steps for 1.2.0

1. **Consider execution modes** for functions that need non-deterministic operations
2. **Update capability declarations** to use more specific permissions
3. **Update build scripts** to use new CLI options if beneficial

## Upgrading from 1.2.x to 2.0.0

### Breaking Changes

#### 1. WebAssembly Target Changes

Functions now compile to WASI (WebAssembly System Interface) instead of the custom runtime.

**Impact:** Existing functions may need minor adjustments for WASI compatibility.

#### 2. Dependency Management

Stricter dependency validation and different import handling.

### Migration Steps for 2.0.0

1. **Update function code** for WASI compatibility
2. **Review and update dependencies**
3. **Rebuild all functions** with new compiler
4. **Update deployment configurations**

## General Upgrade Process

### 1. Check Release Notes

Always read the release notes for the version you're upgrading to:

```bash
# Check current version
flypy --version

# Read changelog (hypothetical - check actual documentation)
curl https://docs.functionfly.com/changelog
```

### 2. Update Dependencies

```bash
# Update FlyPy
pip install --upgrade flypy

# Update other dependencies as needed
pip install --upgrade -r requirements.txt
```

### 3. Update Code

Apply the version-specific changes mentioned above.

### 4. Test Locally

```bash
# Test functions locally
flypy local functions.py my-function --port 8080

# Build and verify
flypy build functions.py
flypy verify ./dist/my-function
```

### 5. Deploy Gradually

```bash
# Deploy to staging first
flypy deploy ./dist/my-function --app-id staging-app --token STAGING_TOKEN

# Test in staging
# ...

# Deploy to production
flypy deploy ./dist/my-function --app-id prod-app --token PROD_TOKEN
```

## Common Upgrade Issues

### Compilation Failures

**Problem:** Functions fail to compile after upgrade.

**Solutions:**
1. Check for deprecated APIs in your code
2. Update function decorators as described above
3. Review error messages for specific guidance
4. Use `flypy build --verbose` for detailed error information

### Runtime Errors

**Problem:** Functions compile but fail at runtime.

**Solutions:**
1. Check for changes in execution environment
2. Update capability declarations
3. Review function logic for compatibility
4. Use local testing to debug

### Performance Regressions

**Problem:** Upgraded functions perform worse.

**Solutions:**
1. Check if execution mode settings are appropriate
2. Review bundle sizes with new compiler
3. Consider optimization flags
4. Monitor cold start times

## Version-Specific Migration Scripts

### Automated Migration for 1.0.0

```python
#!/usr/bin/env python3
"""
Automated migration script for FlyPy 1.0.0
"""

import ast
import re

def migrate_file(filepath: str) -> None:
    """Migrate a Python file to FlyPy 1.0.0 syntax."""

    with open(filepath, 'r') as f:
        content = f.read()

    # Migrate function decorators
    content = re.sub(
        r'@flypy\.function\("([^"]+)"\)',
        r'@flypy.function(name="\1")',
        content
    )

    # Migrate schema decorators
    def migrate_schema(match):
        schema_content = match.group(1)
        # This is a simplified migration - real implementation would parse JSON
        return f'@flypy.input_schema({schema_content})'

    content = re.sub(
        r'@flypy\.schema\((.*?)\)',
        migrate_schema,
        content,
        flags=re.DOTALL
    )

    with open(filepath, 'w') as f:
        f.write(content)

if __name__ == "__main__":
    import sys
    for filepath in sys.argv[1:]:
        migrate_file(filepath)
        print(f"Migrated {filepath})")
```

## Best Practices for Upgrades

### 1. Version Pinning

Use specific versions in production:

```python
# requirements.txt
flypy==1.2.3  # Pin to specific version
```

### 2. Gradual Rollouts

```bash
# Deploy to 10% of traffic first
flypy deploy ./dist/my-function --traffic-split 0.1

# Gradually increase traffic
flypy deploy ./dist/my-function --traffic-split 0.5
flypy deploy ./dist/my-function --traffic-split 1.0
```

### 3. Rollback Plan

Always have a rollback plan:

```bash
# Keep previous version deployed
flypy deploy ./dist/old-version --app-id rollback-app

# Quick rollback if needed
flypy deploy ./dist/old-version --app-id prod-app
```

### 4. Monitoring

Monitor key metrics during upgrades:

- Function execution time
- Error rates
- Cold start performance
- Bundle sizes

### 5. Testing

Comprehensive testing before and after upgrades:

```python
# Unit tests
pytest tests/

# Integration tests
# Test against deployed functions

# Load testing
# Ensure performance meets requirements
```

## Getting Help

If you encounter issues during upgrade:

1. Check the [FlyPy documentation](https://docs.functionfly.com)
2. Review [GitHub issues](https://github.com/functionfly/functionfly/issues)
3. Join the [FunctionFly community](https://discord.gg/functionfly)
4. Contact [support@functionfly.com](mailto:support@functionfly.com)

Remember: Always backup your code and test thoroughly before upgrading in production!