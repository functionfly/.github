# FlyPy API Reference

This document provides comprehensive API reference documentation for the FlyPy Python SDK.

## Table of Contents

- [Core API](#core-api)
- [Agent Integrations API](#agent-integrations-api)
- [Decorators](#decorators)
- [Schema System](#schema-system)
- [Build System](#build-system)
- [CLI Commands](#cli-commands)
- [Types](#types)

## Core API

### flypy.function

```python
def function(
    name: Optional[str] = None,
    version: str = "1.0.0",
    description: Optional[str] = None,
    deterministic: bool = True,
    idempotent: bool = False,
    pure: bool = False,
    cache_ttl: Optional[int] = None,
    capabilities: Optional[List[str]] = None,
    max_execution_time: Optional[int] = None,
    execution_mode: ExecutionMode = ExecutionMode.DETERMINISTIC,
) -> Callable
```

## Agent Integrations API

### TrustPolicy

```python
from flypy import TrustPolicy

policy = TrustPolicy(
    min_trust_score=80,          # default
    require_verified=True,       # default
    required_trust_levels=["high"],
    capabilities_allow=["http_get"],
    capabilities_deny=["secrets_read"],
)
```

`TrustPolicy.policy_hash()` returns a deterministic hash used in execution metadata.

### AgentClient

```python
from flypy import AgentClient

client = AgentClient(
    api_base="https://api.functionfly.com",
    api_key="...",
    timeout_seconds=10.0,
    max_retries=2,
)
```

Key methods:
- `search_registry(query, category=None, min_rating=None, limit=20, offset=0)`
- `get_function_profile(author, name, expand_manifest=True)`
- `get_ai_schema(author, name)`
- `discover_trusted_functions(policy, query, category=None, limit=20)`
- `execute_trusted_tool(trusted_function, policy, tool_input)`

### Framework Adapters

```python
from flypy import LangChainAdapter, AutoGenAdapter, CrewAIAdapter
```

Each adapter exposes:
- `build_tools(policy, query, category=None, limit=20)`
- `execute_tool(trusted_function, policy, tool_input)` (LangChain adapter)

All adapters route execution through FunctionFly and include audit metadata:
- `tool_id`
- `author`, `name`, `version`
- `policy_hash`

The main decorator that marks a function as a FlyPy function.

**Parameters:**
- `name` (str, optional): Function name (defaults to function.__name__)
- `version` (str): Semantic version string (default: "1.0.0")
- `description` (str, optional): Function description
- `deterministic` (bool): Whether the function is deterministic (default: True)
- `idempotent` (bool): Whether the function is idempotent (default: False)
- `pure` (bool): Whether the function is pure (default: False)
- `cache_ttl` (int, optional): Cache TTL in seconds
- `capabilities` (List[str], optional): Required capabilities (e.g., ["network", "filesystem"])
- `max_execution_time` (int, optional): Maximum execution time in milliseconds
- `execution_mode` (ExecutionMode): Execution mode (default: DETERMINISTIC)

**Returns:**
- Decorated function

**Example:**
```python
@flypy.function(
    name="calculate-total",
    deterministic=True,
    idempotent=True,
    cache_ttl=3600,
    capabilities=["math"]
)
def handler(event):
    return {"total": sum(event["items"])}
```

### flypy.input_schema

```python
def input_schema(schema: Union[Dict[str, Any], Schema]) -> Callable
```

Decorator to specify input schema for a FlyPy function.

**Parameters:**
- `schema` (dict or Schema): Input schema as dict or Schema object

**Returns:**
- Decorator function

**Example:**
```python
@flypy.input_schema({
    "type": "object",
    "properties": {
        "name": {"type": "string"},
        "age": {"type": "integer"}
    },
    "required": ["name"]
})
@flypy.function(name="process-user")
def handler(event):
    return {"message": f"Hello {event['name']}!"}
```

### flypy.output_schema

```python
def output_schema(schema: Union[Dict[str, Any], Schema]) -> Callable
```

Decorator to specify output schema for a FlyPy function.

**Parameters:**
- `schema` (dict or Schema): Output schema as dict or Schema object

**Returns:**
- Decorator function

**Example:**
```python
@flypy.output_schema({
    "type": "object",
    "properties": {
        "result": {"type": "number"},
        "status": {"type": "string"}
    },
    "required": ["result"]
})
@flypy.function(name="calculate")
def handler(event):
    return {"result": event["a"] + event["b"], "status": "success"}
```

## Schema System

### Schema

```python
class Schema:
    def __init__(self, title: Optional[str] = None, description: Optional[str] = None)
```

JSON Schema builder for FlyPy functions.

**Methods:**

#### add_field

```python
def add_field(self, name: str, field: Field) -> 'Schema'
```

Add a field to the schema.

**Parameters:**
- `name` (str): Field name
- `field` (Field): Field definition

**Returns:**
- Schema instance (for chaining)

#### to_dict

```python
def to_dict(self) -> Dict[str, Any]
```

Convert schema to JSON schema dictionary.

**Returns:**
- JSON schema dictionary

#### to_json

```python
def to_json(self, indent: Optional[int] = 2) -> str
```

Convert schema to JSON string.

**Parameters:**
- `indent` (int, optional): JSON indentation (default: 2)

**Returns:**
- JSON schema string

#### from_dict

```python
@classmethod
def from_dict(cls, schema_dict: Dict[str, Any]) -> 'Schema'
```

Create schema from JSON schema dictionary.

**Parameters:**
- `schema_dict` (dict): JSON schema dictionary

**Returns:**
- Schema instance

#### infer_from_function

```python
@classmethod
def infer_from_function(
    cls,
    func,
    input_annotations: Optional[Dict[str, Any]] = None,
    output_annotations: Optional[Dict[str, Any]] = None
) -> tuple['Schema', 'Schema']
```

Infer input and output schemas from function type hints.

**Parameters:**
- `func`: The function to analyze
- `input_annotations` (dict, optional): Override input type hints
- `output_annotations` (dict, optional): Override output type hints

**Returns:**
- Tuple of (input_schema, output_schema)

### Field

```python
class Field:
    def __init__(
        self,
        type_: Union[str, SchemaType],
        description: Optional[str] = None,
        required: bool = True,
        default: Any = None,
        minimum: Optional[Union[int, float]] = None,
        maximum: Optional[Union[int, float]] = None,
        min_length: Optional[int] = None,
        max_length: Optional[int] = None,
        pattern: Optional[str] = None,
        enum: Optional[List[Any]] = None,
        items: Optional[Union['Field', Dict[str, Any]]] = None,
        properties: Optional[Dict[str, Union['Field', Dict[str, Any]]]] = None,
        additional_properties: bool = False,
    )
```

Represents a field in a schema with validation rules.

**Parameters:**
- `type_` (str or SchemaType): JSON schema type
- `description` (str, optional): Field description
- `required` (bool): Whether field is required (default: True)
- `default` (Any, optional): Default value
- `minimum` (int/float, optional): Minimum value for numbers
- `maximum` (int/float, optional): Maximum value for numbers
- `min_length` (int, optional): Minimum length for strings
- `max_length` (int, optional): Maximum length for strings
- `pattern` (str, optional): Regex pattern for strings
- `enum` (List[Any], optional): Allowed values
- `items` (Field or dict, optional): Schema for array items
- `properties` (dict, optional): Properties for object fields
- `additional_properties` (bool): Allow additional properties in objects (default: False)

**Methods:**

#### to_dict

```python
def to_dict(self) -> Dict[str, Any]
```

Convert field to JSON schema dictionary.

**Returns:**
- JSON schema dictionary

### SchemaType

```python
class SchemaType(str, Enum):
    STRING = "string"
    NUMBER = "number"
    INTEGER = "integer"
    BOOLEAN = "boolean"
    OBJECT = "object"
    ARRAY = "array"
    NULL = "null"
```

JSON Schema types enumeration.

## Build System

### FlyPyBuilder

```python
class FlyPyBuilder:
    def __init__(self, go_binary_path: Optional[str] = None)
```

Builder for FlyPy functions.

**Parameters:**
- `go_binary_path` (str, optional): Path to the FlyPy Go binary

**Methods:**

#### build_function

```python
def build_function(
    self,
    func_name: str,
    output_dir: str = "./dist",
    mode: str = "deterministic",
    verbose: bool = False
) -> BuildResult
```

Build a single FlyPy function.

**Parameters:**
- `func_name` (str): Name of the function to build
- `output_dir` (str): Output directory for artifacts (default: "./dist")
- `mode` (str): Execution mode ("deterministic" or "compatible") (default: "deterministic")
- `verbose` (bool): Enable verbose output (default: False)

**Returns:**
- BuildResult with build information

#### build_all_functions

```python
def build_all_functions(
    self,
    output_dir: str = "./dist",
    mode: str = "deterministic",
    verbose: bool = False
) -> List[BuildResult]
```

Build all registered FlyPy functions.

**Parameters:**
- `output_dir` (str): Output directory for artifacts (default: "./dist")
- `mode` (str): Execution mode ("deterministic" or "compatible") (default: "deterministic")
- `verbose` (bool): Enable verbose output (default: False)

**Returns:**
- List of BuildResult objects

### Convenience Functions

#### build_function

```python
def build_function(
    func_name: str,
    output_dir: str = "./dist",
    mode: str = "deterministic",
    verbose: bool = False,
    go_binary: Optional[str] = None
) -> BuildResult
```

Convenience function to build a single FlyPy function.

**Parameters:**
- `func_name` (str): Name of the function to build
- `output_dir` (str): Output directory for artifacts (default: "./dist")
- `mode` (str): Execution mode ("deterministic" or "compatible") (default: "deterministic")
- `verbose` (bool): Enable verbose output (default: False)
- `go_binary` (str, optional): Path to FlyPy Go binary

**Returns:**
- BuildResult with build information

#### build_all

```python
def build_all(
    output_dir: str = "./dist",
    mode: str = "deterministic",
    verbose: bool = False,
    go_binary: Optional[str] = None
) -> List[BuildResult]
```

Convenience function to build all registered FlyPy functions.

**Parameters:**
- `output_dir` (str): Output directory for artifacts (default: "./dist")
- `mode` (str): Execution mode ("deterministic" or "compatible") (default: "deterministic")
- `verbose` (bool): Enable verbose output (default: False)
- `go_binary` (str, optional): Path to FlyPy Go binary

**Returns:**
- List of BuildResult objects

## CLI Commands

### flypy build

Build FlyPy functions to WebAssembly.

```bash
flypy build [OPTIONS] FILES...

Options:
  -o, --output TEXT          Output directory for artifacts (default: ./dist)
  --mode [deterministic|compatible]
                             Execution mode (default: deterministic)
  -v, --verbose              Verbose output
  --go-binary TEXT           Path to FlyPy Go binary (auto-detected if not specified)
```

### flypy deploy

Deploy FlyPy functions to FunctionFly.

```bash
flypy deploy [OPTIONS] ARTIFACT_DIR

Options:
  --registry TEXT            FunctionFly registry URL (default: https://api.functionfly.com)
  --token TEXT               Authentication token (required)
  --app-id TEXT              FunctionFly app ID (required)
  --provider [cloudflare|vercel|fly|deno]
                             Cloud provider (default: cloudflare)
  --region TEXT              Deployment region (default: us-east-1)
```

### flypy list

List registered FlyPy functions.

```bash
flypy list [FILES...]
```

### flypy local

Run FlyPy functions locally for testing.

```bash
flypy local [OPTIONS] FILE FUNCTION

Options:
  -p, --port INTEGER         Port to run local server on (default: 8080)
```

### flypy verify

Verify determinism of built artifacts.

```bash
flypy verify ARTIFACT_DIR
```

## Types

### FunctionMetadata

```python
class FunctionMetadata(BaseModel):
    name: str
    version: str = "1.0.0"
    description: Optional[str] = None

    # Execution properties
    deterministic: bool = True
    idempotent: bool = False
    pure: bool = False

    # Caching
    cache_ttl: Optional[int] = None

    # Capabilities and permissions
    capabilities: List[str] = Field(default_factory=list)
    max_execution_time: Optional[int] = None  # in milliseconds

    # Schema information
    input_schema: Optional[Dict[str, Any]] = None
    output_schema: Optional[Dict[str, Any]] = None

    # Build information
    source_file: Optional[str] = None
    source_hash: Optional[str] = None

    # Runtime information
    execution_mode: ExecutionMode = ExecutionMode.DETERMINISTIC
```

Metadata for a FlyPy function.

### FunctionDefinition

```python
class FunctionDefinition(BaseModel):
    metadata: FunctionMetadata
    source_code: str
    ast_json: Optional[str] = None  # Serialized AST for the Go compiler

    # Derived information
    dependencies: List[str] = Field(default_factory=list)
    imports: List[str] = Field(default_factory=list)
```

Complete function definition including code and metadata.

### BuildResult

```python
class BuildResult(BaseModel):
    success: bool
    function_name: str
    output_dir: str
    warnings: List[str] = Field(default_factory=list)
    errors: List[str] = Field(default_factory=list)

    # Artifact information
    wasm_file: Optional[str] = None
    manifest_file: Optional[str] = None
    determinism_hash: Optional[str] = None

    # Performance metrics
    build_time_ms: Optional[int] = None
    wasm_size_bytes: Optional[int] = None
```

Result of building a FlyPy function.

### ExecutionMode

```python
class ExecutionMode(str, Enum):
    DETERMINISTIC = "deterministic"
    COMPATIBLE = "compatible"
```

Execution modes for FlyPy functions.