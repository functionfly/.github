# FlyPy Complex Mode

FlyPy Complex Mode is an execution mode that extends the standard deterministic mode to support additional Python modules while maintaining deterministic execution guarantees. This mode bridges the gap between the strict minimal stdlib of deterministic mode and the full Python compatibility mode.

## Overview

### Execution Modes Comparison

| Feature | Deterministic | Complex | Compatible |
|---------|--------------|---------|------------|
| **Execution** | Pure WASM | Pure WASM | MicroPython VM |
| **Startup Time** | Fast (~1ms) | Fast (~1ms) | Slower (~50ms) |
| **Determinism** | Guaranteed | Guaranteed | Best effort |
| **Module Support** | Minimal | Extended | Full Python |
| **Use Case** | Simple transforms | Data processing | Full Python apps |

### When to Use Complex Mode

Use complex mode when you need:
- CSV parsing and generation
- StringIO/BytesIO for in-memory I/O
- Regular expressions
- Date/time operations
- Iterator utilities (itertools)
- Functional programming utilities (functools)
- Hashing (hashlib)
- Base64 encoding/decoding
- UUID generation (deterministic UUID5 only)
- Decimal arithmetic
- Rational numbers (fractions)

## Supported Modules

### Core Modules (Also in Deterministic Mode)

| Module | Description | Key Functions |
|--------|-------------|---------------|
| `json` | JSON encoding/decoding | `loads()`, `dumps()` |
| `math` | Mathematical functions | `sqrt()`, `sin()`, `cos()`, `log()` |
| `typing` | Type hints | `List`, `Dict`, `Optional`, etc. |
| `collections` | Collection utilities | `defaultdict`, `Counter`, `deque` |

### Complex Mode Extensions

| Module | Description | Key Functions |
|--------|-------------|---------------|
| `csv` | CSV parsing/writing | `reader()`, `writer()`, `DictReader()`, `DictWriter()` |
| `io` | In-memory I/O | `StringIO`, `BytesIO` |
| `re` | Regular expressions | `match()`, `search()`, `findall()`, `sub()` |
| `datetime` | Date/time operations | `datetime`, `date`, `time`, `timedelta` |
| `itertools` | Iterator utilities | `chain()`, `cycle()`, `islice()`, `groupby()` |
| `functools` | Functional utilities | `partial()`, `reduce()`, `lru_cache()` |
| `operator` | Operator functions | `itemgetter()`, `attrgetter()` |
| `string` | String constants | `ascii_letters`, `digits`, `punctuation` |
| `textwrap` | Text wrapping | `wrap()`, `fill()`, `dedent()` |
| `hashlib` | Hashing algorithms | `md5()`, `sha1()`, `sha256()` |
| `base64` | Base64 encoding | `b64encode()`, `b64decode()` |
| `uuid` | UUID generation | `uuid5()` only (deterministic) |
| `decimal` | Decimal arithmetic | `Decimal` |
| `fractions` | Rational numbers | `Fraction` |

## Usage

### CLI

```bash
# Compile with complex mode
flypy compile --mode complex main.py -o function.wasm

# Run with complex mode support
functionfly-local --wasm function.wasm --mode complex
```

### Example: CSV Processing

```python
import csv
import io
import json

def handler(event):
    # Parse input JSON
    data = event.get("data", [])
    
    # Create CSV in memory
    output = io.StringIO()
    if data:
        writer = csv.DictWriter(output, fieldnames=data[0].keys())
        writer.writeheader()
        writer.writerows(data)
    
    return {
        "csv": output.getvalue(),
        "rows": len(data)
    }
```

### Example: Regex Processing

```python
import re
import json

def handler(event):
    text = event.get("text", "")
    pattern = event.get("pattern", r"\w+")
    
    # Find all matches
    matches = re.findall(pattern, text)
    
    return {
        "matches": matches,
        "count": len(matches)
    }
```

### Example: Date/Time Processing

```python
from datetime import datetime, timedelta
import json

def handler(event):
    # Parse input date
    date_str = event.get("date", "2024-01-01")
    input_date = datetime.strptime(date_str, "%Y-%m-%d")
    
    # Add days
    days_to_add = event.get("days", 7)
    result_date = input_date + timedelta(days=days_to_add)
    
    return {
        "original": date_str,
        "result": result_date.strftime("%Y-%m-%d"),
        "weekday": result_date.strftime("%A")
    }
```

### Example: Hashing

```python
import hashlib
import json

def handler(event):
    data = event.get("data", "")
    
    return {
        "md5": hashlib.md5(data.encode()).hexdigest(),
        "sha256": hashlib.sha256(data.encode()).hexdigest()
    }
```

## Restrictions

Even in complex mode, certain operations remain forbidden to ensure deterministic execution:

### Forbidden Modules
- `os` - Operating system interface
- `sys` - System-specific parameters
- `random` - Non-deterministic random (use `hashlib` for deterministic randomness)
- `subprocess` - Process creation
- `socket` - Network operations
- `threading` - Thread operations
- `multiprocessing` - Process operations

### Forbidden Operations in `io` Module
- `open()` - File I/O
- `FileIO` - Direct file access
- `BufferedReader`/`BufferedWriter` - File buffering

### Forbidden in `uuid` Module
- `uuid1()` - Time-based UUID (non-deterministic)
- `uuid4()` - Random UUID (non-deterministic)
- Use `uuid5()` for deterministic UUIDs

## Implementation Details

### Backend Support

Complex mode generates Rust code that compiles to WebAssembly. The Rust backend includes helper functions for:

1. **CSV Operations**: Uses the `csv` crate for parsing and writing
2. **StringIO/BytesIO**: Implemented using `Vec<u8>` and `String` with seek/tell support
3. **Regex**: Uses the `regex` crate for pattern matching
4. **Datetime**: Uses the `chrono` crate for date/time operations
5. **Hashing**: Uses Rust's native crypto implementations

### Performance Characteristics

| Operation | Deterministic Mode | Complex Mode | Overhead |
|-----------|-------------------|--------------|----------|
| JSON parse | ~10µs | ~10µs | None |
| CSV parse (100 rows) | N/A | ~50µs | N/A |
| Regex match | N/A | ~5µs | N/A |
| SHA256 hash (1KB) | N/A | ~2µs | N/A |

## Migration Guide

### From Deterministic Mode

If your code uses only deterministic modules, no changes are needed. Simply add `--mode complex` to enable additional modules.

### From Compatible Mode

When migrating from compatible mode:

1. **Check imports**: Ensure all imports are in the complex mode allowlist
2. **Replace file I/O**: Use `StringIO`/`BytesIO` instead of file operations
3. **Replace random**: Use `hashlib`-based deterministic randomness
4. **Replace time**: Use `datetime` for date operations (no `time.time()`)

### Example Migration

**Before (Compatible Mode):**
```python
import random
import time

def handler(event):
    return {
        "random": random.randint(1, 100),
        "timestamp": time.time()
    }
```

**After (Complex Mode):**
```python
import hashlib
from datetime import datetime

def handler(event):
    # Deterministic "random" based on input
    data = event.get("seed", "default")
    hash_val = int(hashlib.md5(data.encode()).hexdigest()[:8], 16)
    
    return {
        "random": (hash_val % 100) + 1,
        "timestamp": datetime.utcnow().isoformat()
    }
```

## Testing

Run the complex mode tests:

```bash
go test ./internal/flypy/restrictions/... -run Complex -v
```

## Future Enhancements

Planned additions to complex mode:
- `xml.etree.ElementTree` - XML parsing
- `html.parser` - HTML parsing
- `configparser` - Configuration file parsing
- `zipfile` - Archive handling (read-only, in-memory)
