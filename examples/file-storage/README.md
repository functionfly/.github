# File Storage Example

This example demonstrates how to use the `storage` capability in FunctionFly functions.

## Overview

The storage capability allows functions to read and write files in a secure, sandboxed storage environment. Files are stored within the function's designated storage directory.

## Usage

### Writing to a file

```json
{
  "action": "write",
  "message": "Hello, World!"
}
```

### Reading from a file

```json
{
  "action": "read"
}
```

## Response Examples

### Write Response

```json
{
  "status": "success",
  "action": "write",
  "filename": "data.json",
  "data_size": 128,
  "data": {
    "message": "Hello, World!",
    "timestamp": "2024-01-01T12:00:00Z",
    "function": "file-storage"
  }
}
```

### Read Response

```json
{
  "status": "success",
  "action": "read",
  "filename": "data.json",
  "data": {
    "message": "Sample stored data",
    "timestamp": "2024-01-01T12:00:00",
    "function": "file-storage"
  }
}
```

## Security Notes

- File operations are restricted to the function's storage directory
- Path traversal attacks are prevented
- File size limits may apply based on function configuration