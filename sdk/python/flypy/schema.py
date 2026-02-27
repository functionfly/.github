"""
Schema system for FlyPy function input/output validation.

Provides automatic schema inference from type hints and manual schema definitions.
"""

import inspect
from typing import Dict, List, Any, Optional, Union, get_type_hints, get_origin, get_args
from enum import Enum
import json

from .types import FunctionMetadata


class SchemaType(str, Enum):
    """JSON Schema types."""

    STRING = "string"
    NUMBER = "number"
    INTEGER = "integer"
    BOOLEAN = "boolean"
    OBJECT = "object"
    ARRAY = "array"
    NULL = "null"


class Field:
    """Represents a field in a schema with validation rules."""

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
    ):
        self.type = type_ if isinstance(type_, str) else type_.value
        self.description = description
        self.required = required
        self.default = default
        self.minimum = minimum
        self.maximum = maximum
        self.min_length = min_length
        self.max_length = max_length
        self.pattern = pattern
        self.enum = enum
        self.items = items
        self.properties = properties
        self.additional_properties = additional_properties

    def to_dict(self) -> Dict[str, Any]:
        """Convert field to JSON schema dictionary."""
        result = {"type": self.type}

        if self.description:
            result["description"] = self.description
        if self.default is not None:
            result["default"] = self.default
        if self.minimum is not None:
            result["minimum"] = self.minimum
        if self.maximum is not None:
            result["maximum"] = self.maximum
        if self.min_length is not None:
            result["minLength"] = self.min_length
        if self.max_length is not None:
            result["maxLength"] = self.max_length
        if self.pattern:
            result["pattern"] = self.pattern
        if self.enum:
            result["enum"] = self.enum
        if self.items:
            if isinstance(self.items, Field):
                result["items"] = self.items.to_dict()
            else:
                result["items"] = self.items
        if self.properties:
            props = {}
            for key, value in self.properties.items():
                if isinstance(value, Field):
                    props[key] = value.to_dict()
                else:
                    props[key] = value
            result["properties"] = props
        if not self.additional_properties:
            result["additionalProperties"] = False

        return result


class Schema:
    """JSON Schema builder for FlyPy functions."""

    def __init__(self, title: Optional[str] = None, description: Optional[str] = None):
        self.title = title
        self.description = description
        self.properties: Dict[str, Field] = {}
        self.required: List[str] = []

    def add_field(self, name: str, field: Field) -> 'Schema':
        """Add a field to the schema."""
        self.properties[name] = field
        if field.required:
            if name not in self.required:
                self.required.append(name)
        return self

    def to_dict(self) -> Dict[str, Any]:
        """Convert schema to JSON schema dictionary."""
        result = {
            "$schema": "http://json-schema.org/draft-07/schema#",
            "type": "object",
        }

        if self.title:
            result["title"] = self.title
        if self.description:
            result["description"] = self.description

        if self.properties:
            props = {}
            for name, field in self.properties.items():
                props[name] = field.to_dict()
            result["properties"] = props

        if self.required:
            result["required"] = sorted(self.required)

        result["additionalProperties"] = False

        return result

    def to_json(self, indent: Optional[int] = 2) -> str:
        """Convert schema to JSON string."""
        return json.dumps(self.to_dict(), indent=indent)

    @classmethod
    def from_dict(cls, schema_dict: Dict[str, Any]) -> 'Schema':
        """Create schema from JSON schema dictionary."""
        schema = cls(
            title=schema_dict.get("title"),
            description=schema_dict.get("description")
        )

        properties = schema_dict.get("properties", {})
        required = schema_dict.get("required", [])

        for name, field_dict in properties.items():
            field = Field(
                type_=field_dict["type"],
                description=field_dict.get("description"),
                required=name in required,
                default=field_dict.get("default"),
                minimum=field_dict.get("minimum"),
                maximum=field_dict.get("maximum"),
                min_length=field_dict.get("minLength"),
                max_length=field_dict.get("maxLength"),
                pattern=field_dict.get("pattern"),
                enum=field_dict.get("enum"),
                items=field_dict.get("items"),
                properties=field_dict.get("properties"),
                additional_properties=field_dict.get("additionalProperties", True),
            )
            schema.add_field(name, field)

        return schema

    @classmethod
    def infer_from_function(
        cls,
        func,
        input_annotations: Optional[Dict[str, Any]] = None,
        output_annotations: Optional[Dict[str, Any]] = None
    ) -> tuple['Schema', 'Schema']:
        """
        Infer input and output schemas from function type hints.

        Args:
            func: The function to analyze
            input_annotations: Override input type hints
            output_annotations: Override output type hints

        Returns:
            Tuple of (input_schema, output_schema)
        """
        try:
            hints = get_type_hints(func)
        except (NameError, TypeError):
            # Type hints not available or invalid
            hints = {}

        # Get parameter annotations (excluding 'return')
        input_hints = {k: v for k, v in hints.items() if k != 'return'}
        if input_annotations:
            input_hints.update(input_annotations)

        # Get return annotation
        output_hint = hints.get('return')
        if output_annotations:
            output_hint = output_annotations

        input_schema = cls("Function Input", f"Input schema for {func.__name__}")
        output_schema = cls("Function Output", f"Output schema for {func.__name__}")

        # Infer input schema from parameters
        sig = inspect.signature(func)
        for param_name, param in sig.parameters.items():
            if param_name == 'self':
                continue

            param_hint = input_hints.get(param_name)
            field = _infer_field_from_type(param_hint, param_name)
            input_schema.add_field(param_name, field)

        # Infer output schema from return type
        if output_hint:
            # Check if it's a Dict type that should create multiple output fields
            if _is_dict_type(output_hint):
                # For Dict[str, SomeType], infer the value type
                if hasattr(output_hint, '__args__') and len(output_hint.__args__) >= 2:
                    value_type = output_hint.__args__[1]
                    # This is a simple case - just create a result field with the value type
                    output_field = _infer_field_from_type(value_type, "result")
                    output_schema.add_field("result", output_field)
                else:
                    # Generic dict
                    output_schema.add_field("result", Field(SchemaType.OBJECT, additional_properties=True))
            else:
                # Single return value
                output_field = _infer_field_from_type(output_hint, "result")
                output_schema.add_field("result", output_field)

        return input_schema, output_schema


def _is_dict_type(type_hint) -> bool:
    """Check if a type hint represents a dictionary."""
    origin = get_origin(type_hint)
    return (type_hint == Dict or
            origin == dict or
            (hasattr(type_hint, '__origin__') and type_hint.__origin__ == dict))


def _infer_field_from_type(type_hint, field_name: str) -> Field:
    """Infer a Field from a Python type hint."""
    if type_hint is None:
        return Field(SchemaType.STRING, description=f"Parameter {field_name}")

    origin = get_origin(type_hint)
    args = get_args(type_hint)

    # Handle basic types
    if type_hint == str:
        return Field(SchemaType.STRING, description=f"String parameter {field_name}")
    elif type_hint == int:
        return Field(SchemaType.INTEGER, description=f"Integer parameter {field_name}")
    elif type_hint == float:
        return Field(SchemaType.NUMBER, description=f"Number parameter {field_name}")
    elif type_hint == bool:
        return Field(SchemaType.BOOLEAN, description=f"Boolean parameter {field_name}")

    # Handle Optional types
    elif origin is Union:
        # Check if it's Optional[T] (Union[T, None])
        non_none_args = [arg for arg in args if arg is not type(None)]
        if len(non_none_args) == 1:
            field = _infer_field_from_type(non_none_args[0], field_name)
            field.required = False
            return field

    # Handle List types
    elif origin is list or type_hint == List:
        item_type = args[0] if args else Any
        item_field = _infer_field_from_type(item_type, f"{field_name} item")
        return Field(
            SchemaType.ARRAY,
            description=f"Array of {item_type}",
            items=item_field
        )

    # Handle Dict types
    elif origin is dict or type_hint == Dict:
        # Try to infer more specific structure if possible
        if len(args) >= 2 and args[0] == str:
            # Dict[str, SomeType] - can infer value type
            value_type = args[1]
            value_field = _infer_field_from_type(value_type, f"{field_name} value")
            return Field(
                SchemaType.OBJECT,
                description=f"Dictionary with string keys and {value_type} values",
                additional_properties=value_field.to_dict() if hasattr(value_field, 'to_dict') else True
            )
        else:
            return Field(
                SchemaType.OBJECT,
                description=f"Object parameter {field_name}",
                additional_properties=True
            )

    # Handle Tuple types (fixed-length arrays)
    elif origin is tuple or type_hint == tuple:
        if args:
            # Create items schema for tuple elements
            items = []
            for i, arg in enumerate(args):
                item_field = _infer_field_from_type(arg, f"{field_name}[{i}]")
                items.append(item_field.to_dict())
            return Field(
                SchemaType.ARRAY,
                description=f"Tuple of {len(args)} elements",
                items=items
            )
        else:
            return Field(
                SchemaType.ARRAY,
                description=f"Tuple parameter {field_name}"
            )

    # Handle Literal types (Python 3.8+)
    try:
        from typing import Literal
        if origin is Literal:
            # Create enum from literal values
            return Field(
                SchemaType.STRING,  # Could be other types, but default to string
                description=f"Literal value {field_name}",
                enum=list(args)
            )
    except ImportError:
        pass

    # Handle TypedDict (if available)
    try:
        from typing import TypedDict
        if hasattr(type_hint, '__annotations__'):
            # This looks like a TypedDict or similar
            properties = {}
            required = []
            for attr_name, attr_type in type_hint.__annotations__.items():
                prop_field = _infer_field_from_type(attr_type, f"{field_name}.{attr_name}")
                properties[attr_name] = prop_field
                if prop_field.required:
                    required.append(attr_name)

            return Field(
                SchemaType.OBJECT,
                description=f"Typed object {field_name}",
                properties=properties
            )
    except:
        pass

    # Handle dataclasses
    try:
        import dataclasses
        if dataclasses.is_dataclass(type_hint):
            properties = {}
            required = []
            for field_info in dataclasses.fields(type_hint):
                field_type = field_info.type
                prop_field = _infer_field_from_type(field_type, f"{field_name}.{field_info.name}")
                properties[field_info.name] = prop_field
                if prop_field.required and field_info.default == dataclasses.MISSING:
                    required.append(field_info.name)

            return Field(
                SchemaType.OBJECT,
                description=f"Dataclass {field_name}",
                properties=properties
            )
    except:
        pass

    # Handle Pydantic models
    try:
        from pydantic import BaseModel
        if issubclass(type_hint, BaseModel):
            schema = type_hint.model_json_schema()
            return Field(
                SchemaType.OBJECT,
                description=f"Pydantic model {field_name}",
                properties=schema.get("properties", {}),
                required=schema.get("required", [])
            )
    except:
        pass

    # Default to string for unknown types
    return Field(SchemaType.STRING, description=f"Parameter {field_name} ({type_hint})")


def validate_schema(data: Any, schema: Schema) -> List[str]:
    """
    Validate data against a schema.

    Returns a list of validation errors (empty if valid).
    """
    import jsonschema

    try:
        jsonschema.validate(data, schema.to_dict())
        return []
    except jsonschema.ValidationError as e:
        return [str(e.message)]
    except Exception as e:
        return [f"Schema validation error: {str(e)}"]