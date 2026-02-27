# Tutorial: Data Processing Pipeline

This tutorial demonstrates how to build a scalable data processing pipeline using FlyPy, processing large datasets through multiple transformation stages with proper error handling and monitoring.

## Overview

We'll build a data processing pipeline that:

1. **Data Ingestion** - Accept and validate input data
2. **Data Cleaning** - Remove duplicates, handle missing values
3. **Data Transformation** - Apply business logic transformations
4. **Data Aggregation** - Group and summarize data
5. **Data Export** - Format and output processed data

## Prerequisites

```bash
pip install flypy
```

## Step 1: Define Data Models

Let's define the data structures for our processing pipeline:

```python
# models.py
from typing import List, Dict, Any, Optional, Union
from pydantic import BaseModel
from enum import Enum
from datetime import datetime

class DataFormat(str, Enum):
    JSON = "json"
    CSV = "csv"
    XML = "xml"

class ProcessingStatus(str, Enum):
    RECEIVED = "received"
    CLEANING = "cleaning"
    TRANSFORMING = "transforming"
    AGGREGATING = "aggregating"
    COMPLETED = "completed"
    FAILED = "failed"

class DataRecord(BaseModel):
    id: str
    timestamp: datetime
    data: Dict[str, Any]
    source: str
    metadata: Optional[Dict[str, Any]] = None

class ProcessingResult(BaseModel):
    record_id: str
    status: ProcessingStatus
    processed_data: Optional[Dict[str, Any]] = None
    errors: Optional[List[str]] = None
    processing_time_ms: Optional[int] = None

class BatchProcessingRequest(BaseModel):
    batch_id: str
    records: List[DataRecord]
    processing_config: Dict[str, Any]
    output_format: DataFormat = DataFormat.JSON

class BatchProcessingResult(BaseModel):
    batch_id: str
    total_records: int
    processed_records: int
    failed_records: int
    results: List[ProcessingResult]
    processing_stats: Dict[str, Any]
```

## Step 2: Data Ingestion Function

Create a function to validate and ingest incoming data:

```python
# data_ingestion.py
import flypy
from typing import Dict, Any, List
from datetime import datetime
import uuid
from models import DataRecord, BatchProcessingRequest, DataFormat

@flypy.function(
    name="ingest-data-batch",
    description="Validate and ingest a batch of data records",
    deterministic=True,
    idempotent=True,
    max_execution_time=60000  # 60 seconds
)
def ingest_data_batch(raw_data: Dict[str, Any]) -> Dict[str, Any]:
    """
    Ingest and validate a batch of data records.

    Args:
        raw_data: Raw batch data containing records and configuration

    Returns:
        Validation and ingestion result
    """
    errors = []

    # Validate batch structure
    if not raw_data.get("batch_id"):
        errors.append("batch_id is required")
    if not raw_data.get("records"):
        errors.append("records array is required")
    if not isinstance(raw_data["records"], list):
        errors.append("records must be an array")

    if errors:
        return {"valid": False, "errors": errors}

    records = raw_data["records"]
    if len(records) == 0:
        errors.append("records array cannot be empty")
        return {"valid": False, "errors": errors}

    if len(records) > 10000:  # Limit batch size
        errors.append("batch size cannot exceed 10,000 records")
        return {"valid": False, "errors": errors}

    # Validate and transform records
    validated_records = []
    record_errors = []

    for i, record in enumerate(records):
        record_errors_i = []

        # Required fields
        if not record.get("id"):
            record_errors_i.append("id is required")
        if not record.get("data"):
            record_errors_i.append("data is required")
        if not record.get("source"):
            record_errors_i.append("source is required")

        # Validate timestamp
        timestamp_str = record.get("timestamp")
        if timestamp_str:
            try:
                # Try to parse timestamp
                if isinstance(timestamp_str, str):
                    datetime.fromisoformat(timestamp_str.replace('Z', '+00:00'))
                elif isinstance(timestamp_str, (int, float)):
                    datetime.fromtimestamp(timestamp_str)
                else:
                    record_errors_i.append("timestamp must be ISO string or unix timestamp")
            except:
                record_errors_i.append("invalid timestamp format")
        else:
            record_errors_i.append("timestamp is required")

        if record_errors_i:
            record_errors.append({"record_index": i, "errors": record_errors_i})
        else:
            # Create validated record
            validated_record = DataRecord(
                id=record["id"],
                timestamp=record.get("timestamp"),
                data=record["data"],
                source=record["source"],
                metadata=record.get("metadata", {})
            )
            validated_records.append(validated_record)

    if record_errors:
        return {
            "valid": False,
            "batch_id": raw_data["batch_id"],
            "errors": errors + record_errors,
            "valid_records": len(validated_records),
            "invalid_records": len(record_errors)
        }

    # Create processing request
    processing_request = BatchProcessingRequest(
        batch_id=raw_data["batch_id"],
        records=validated_records,
        processing_config=raw_data.get("processing_config", {}),
        output_format=raw_data.get("output_format", DataFormat.JSON)
    )

    return {
        "valid": True,
        "batch_id": raw_data["batch_id"],
        "record_count": len(validated_records),
        "processing_request": processing_request.dict(),
        "message": f"Successfully ingested {len(validated_records)} records"
    }
```

## Step 3: Data Cleaning Function

Create a function to clean and normalize data:

```python
# data_cleaning.py
import flypy
from typing import Dict, Any, List, Optional
from models import DataRecord, ProcessingResult, ProcessingStatus
import re

@flypy.function(
    name="clean-data-records",
    description="Clean and normalize data records",
    deterministic=True,
    pure=True,
    cache_ttl=3600  # Cache for 1 hour
)
def clean_data_records(records: List[Dict[str, Any]], config: Dict[str, Any]) -> Dict[str, Any]:
    """
    Clean and normalize a batch of data records.

    Args:
        records: List of data records to clean
        config: Cleaning configuration options

    Returns:
        Cleaning results with cleaned records
    """
    cleaned_records = []
    errors = []

    # Cleaning rules from config
    remove_duplicates = config.get("remove_duplicates", True)
    normalize_text = config.get("normalize_text", True)
    fill_missing = config.get("fill_missing_values", {})
    validate_schema = config.get("validate_schema", {})

    # Remove duplicates if requested
    if remove_duplicates:
        seen_ids = set()
        unique_records = []
        for record in records:
            record_id = record.get("id")
            if record_id not in seen_ids:
                seen_ids.add(record_id)
                unique_records.append(record)
            else:
                errors.append(f"Duplicate record removed: {record_id}")
        records = unique_records

    for record in records:
        try:
            cleaned_record = clean_single_record(record, config)
            cleaned_records.append(cleaned_record)
        except Exception as e:
            errors.append(f"Failed to clean record {record.get('id')}: {str(e)}")

    return {
        "original_count": len(records),
        "cleaned_count": len(cleaned_records),
        "errors": errors,
        "cleaned_records": cleaned_records
    }

def clean_single_record(record: Dict[str, Any], config: Dict[str, Any]) -> Dict[str, Any]:
    """Clean a single data record."""
    data = record.get("data", {}).copy()

    # Normalize text fields
    if config.get("normalize_text", True):
        for key, value in data.items():
            if isinstance(value, str):
                # Trim whitespace and normalize case
                data[key] = value.strip()
                if config.get("lowercase_text", False):
                    data[key] = data[key].lower()

    # Fill missing values
    fill_missing = config.get("fill_missing_values", {})
    for field, default_value in fill_missing.items():
        if field not in data or data[field] is None:
            data[field] = default_value

    # Validate email format
    if config.get("validate_emails", False):
        email_fields = config.get("email_fields", ["email"])
        for field in email_fields:
            if field in data and isinstance(data[field], str):
                if not is_valid_email(data[field]):
                    raise ValueError(f"Invalid email format: {data[field]}")

    # Remove null/empty fields if requested
    if config.get("remove_empty_fields", False):
        data = {k: v for k, v in data.items() if v is not None and v != ""}

    return {
        **record,
        "data": data,
        "cleaned": True
    }

def is_valid_email(email: str) -> bool:
    """Simple email validation."""
    pattern = r'^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$'
    return re.match(pattern, email) is not None
```

## Step 4: Data Transformation Function

Create a function to apply business logic transformations:

```python
# data_transformation.py
import flypy
from typing import Dict, Any, List, Callable
from models import ProcessingResult, ProcessingStatus
import json

@flypy.function(
    name="transform-data-records",
    description="Apply business logic transformations to data records",
    deterministic=True,
    pure=True
)
def transform_data_records(records: List[Dict[str, Any]], transformations: List[Dict[str, Any]]) -> Dict[str, Any]:
    """
    Apply transformations to data records.

    Args:
        records: List of cleaned data records
        transformations: List of transformation rules

    Returns:
        Transformation results
    """
    transformed_records = []
    errors = []

    for record in records:
        try:
            transformed_record = apply_transformations(record, transformations)
            transformed_records.append(transformed_record)
        except Exception as e:
            errors.append(f"Failed to transform record {record.get('id')}: {str(e)}")

    return {
        "original_count": len(records),
        "transformed_count": len(transformed_records),
        "errors": errors,
        "transformed_records": transformed_records
    }

def apply_transformations(record: Dict[str, Any], transformations: List[Dict[str, Any]]) -> Dict[str, Any]:
    """Apply a list of transformations to a single record."""
    data = record.get("data", {}).copy()

    for transformation in transformations:
        transform_type = transformation.get("type")

        if transform_type == "add_field":
            field_name = transformation["field"]
            value = transformation["value"]
            data[field_name] = value

        elif transform_type == "rename_field":
            old_name = transformation["old_name"]
            new_name = transformation["new_name"]
            if old_name in data:
                data[new_name] = data.pop(old_name)

        elif transform_type == "calculate_field":
            field_name = transformation["field"]
            expression = transformation["expression"]
            try:
                # Simple expression evaluation (in real app, use a safer method)
                result = eval(expression, {"data": data, "__builtins__": {}})
                data[field_name] = result
            except Exception as e:
                raise ValueError(f"Failed to calculate {field_name}: {str(e)}")

        elif transform_type == "map_values":
            field_name = transformation["field"]
            mapping = transformation["mapping"]
            if field_name in data:
                current_value = data[field_name]
                data[field_name] = mapping.get(str(current_value), current_value)

        elif transform_type == "filter_fields":
            keep_fields = set(transformation["fields"])
            data = {k: v for k, v in data.items() if k in keep_fields}

        elif transform_type == "split_field":
            source_field = transformation["source_field"]
            target_fields = transformation["target_fields"]
            separator = transformation.get("separator", ",")

            if source_field in data and isinstance(data[source_field], str):
                parts = data[source_field].split(separator, len(target_fields) - 1)
                for i, target_field in enumerate(target_fields):
                    data[target_field] = parts[i] if i < len(parts) else ""

    return {
        **record,
        "data": data,
        "transformed": True
    }
```

## Step 5: Data Aggregation Function

Create a function to aggregate and summarize data:

```python
# data_aggregation.py
import flypy
from typing import Dict, Any, List
from collections import defaultdict
from models import ProcessingResult, ProcessingStatus

@flypy.function(
    name="aggregate-data-records",
    description="Aggregate and summarize data records",
    deterministic=True,
    pure=True,
    cache_ttl=1800  # Cache for 30 minutes
)
def aggregate_data_records(records: List[Dict[str, Any]], aggregation_config: Dict[str, Any]) -> Dict[str, Any]:
    """
    Aggregate data records based on specified rules.

    Args:
        records: List of transformed data records
        aggregation_config: Aggregation configuration

    Returns:
        Aggregation results
    """
    group_by = aggregation_config.get("group_by", [])
    aggregations = aggregation_config.get("aggregations", [])

    if not group_by:
        # No grouping - aggregate all records
        return aggregate_records(records, aggregations, "all")

    # Group records
    groups = defaultdict(list)
    for record in records:
        data = record.get("data", {})

        # Create group key
        group_key = tuple(data.get(field) for field in group_by)
        groups[group_key].append(record)

    # Aggregate each group
    results = []
    for group_key, group_records in groups.items():
        result = aggregate_records(group_records, aggregations, group_key)
        results.append(result)

    return {
        "group_count": len(results),
        "results": results,
        "total_records": len(records)
    }

def aggregate_records(records: List[Dict[str, Any]], aggregations: List[Dict[str, Any]], group_key: Any) -> Dict[str, Any]:
    """Aggregate a group of records."""
    result = {
        "group_key": group_key,
        "record_count": len(records)
    }

    # Apply aggregations
    for agg in aggregations:
        agg_type = agg["type"]
        field = agg["field"]
        output_field = agg.get("output_field", f"{field}_{agg_type}")

        values = [r.get("data", {}).get(field) for r in records if r.get("data", {}).get(field) is not None]

        if not values:
            result[output_field] = None
            continue

        if agg_type == "count":
            result[output_field] = len(values)
        elif agg_type == "sum":
            result[output_field] = sum(values)
        elif agg_type == "avg":
            result[output_field] = sum(values) / len(values)
        elif agg_type == "min":
            result[output_field] = min(values)
        elif agg_type == "max":
            result[output_field] = max(values)
        elif agg_type == "distinct_count":
            result[output_field] = len(set(values))

    return result
```

## Step 6: Data Export Function

Create a function to format and export processed data:

```python
# data_export.py
import flypy
from typing import Dict, Any, List
from models import DataFormat
import json
import csv
from io import StringIO

@flypy.function(
    name="export-processed-data",
    description="Format and export processed data",
    deterministic=True,
    pure=True
)
def export_processed_data(data: Dict[str, Any], format_config: Dict[str, Any]) -> Dict[str, Any]:
    """
    Export processed data in the requested format.

    Args:
        data: Processed data to export
        format_config: Export format configuration

    Returns:
        Formatted export data
    """
    output_format = format_config.get("format", DataFormat.JSON)
    include_metadata = format_config.get("include_metadata", True)

    if output_format == DataFormat.JSON:
        return export_json(data, format_config)
    elif output_format == DataFormat.CSV:
        return export_csv(data, format_config)
    else:
        raise ValueError(f"Unsupported export format: {output_format}")

def export_json(data: Dict[str, Any], config: Dict[str, Any]) -> Dict[str, Any]:
    """Export data as JSON."""
    # Flatten nested structures if requested
    if config.get("flatten", False):
        if "results" in data:
            data = {"results": [flatten_record(r) for r in data["results"]]}

    return {
        "format": "json",
        "data": json.dumps(data, indent=2, default=str),
        "content_type": "application/json"
    }

def export_csv(data: Dict[str, Any], config: Dict[str, Any]) -> Dict[str, Any]:
    """Export data as CSV."""
    if "results" not in data:
        raise ValueError("CSV export requires 'results' array in data")

    results = data["results"]
    if not results:
        return {
            "format": "csv",
            "data": "",
            "content_type": "text/csv"
        }

    # Determine CSV columns from first record
    first_record = results[0]
    if isinstance(first_record, dict):
        columns = list(first_record.keys())
    else:
        columns = ["value"]

    # Create CSV output
    output = StringIO()
    writer = csv.DictWriter(output, fieldnames=columns)
    writer.writeheader()

    for record in results:
        if isinstance(record, dict):
            # Flatten nested dicts for CSV
            flat_record = flatten_record(record)
            writer.writerow(flat_record)
        else:
            writer.writerow({"value": record})

    return {
        "format": "csv",
        "data": output.getvalue(),
        "content_type": "text/csv",
        "columns": columns
    }

def flatten_record(record: Dict[str, Any], prefix: str = "") -> Dict[str, Any]:
    """Flatten nested dictionary structures."""
    flattened = {}

    for key, value in record.items():
        new_key = f"{prefix}{key}" if prefix else key

        if isinstance(value, dict):
            flattened.update(flatten_record(value, f"{new_key}_"))
        elif isinstance(value, list):
            # Convert lists to comma-separated strings
            flattened[new_key] = ", ".join(str(v) for v in value)
        else:
            flattened[new_key] = value

    return flattened
```

## Step 7: Main Processing Pipeline

Create the main pipeline function that orchestrates all steps:

```python
# data_pipeline.py
import flypy
from typing import Dict, Any
import time

# Import processing functions
from data_ingestion import ingest_data_batch
from data_cleaning import clean_data_records
from data_transformation import transform_data_records
from data_aggregation import aggregate_data_records
from data_export import export_processed_data

@flypy.function(
    name="process-data-pipeline",
    description="Complete data processing pipeline",
    deterministic=False,  # Orchestrates multiple functions
    max_execution_time=300000  # 5 minutes
)
def process_data_pipeline(raw_batch: Dict[str, Any]) -> Dict[str, Any]:
    """
    Process a batch of data through the complete pipeline.

    Args:
        raw_batch: Raw batch data to process

    Returns:
        Complete processing results
    """
    start_time = time.time()
    pipeline_stats = {
        "stages": [],
        "total_time_ms": 0
    }

    try:
        # Stage 1: Data Ingestion
        stage_start = time.time()
        ingestion_result = ingest_data_batch(raw_batch)
        stage_time = int((time.time() - stage_start) * 1000)

        pipeline_stats["stages"].append({
            "stage": "ingestion",
            "duration_ms": stage_time,
            "success": ingestion_result["valid"]
        })

        if not ingestion_result["valid"]:
            return {
                "success": False,
                "failed_stage": "ingestion",
                "errors": ingestion_result["errors"],
                "pipeline_stats": pipeline_stats
            }

        # Extract processing request
        processing_request = ingestion_result["processing_request"]

        # Stage 2: Data Cleaning
        stage_start = time.time()
        cleaning_config = processing_request.get("processing_config", {}).get("cleaning", {})
        records = [{"id": r["id"], "data": r["data"]} for r in processing_request["records"]]

        cleaning_result = clean_data_records(records, cleaning_config)
        stage_time = int((time.time() - stage_start) * 1000)

        pipeline_stats["stages"].append({
            "stage": "cleaning",
            "duration_ms": stage_time,
            "records_processed": cleaning_result["cleaned_count"]
        })

        # Stage 3: Data Transformation
        stage_start = time.time()
        transformation_config = processing_request.get("processing_config", {}).get("transformations", [])
        cleaned_records = cleaning_result["cleaned_records"]

        transformation_result = transform_data_records(cleaned_records, transformation_config)
        stage_time = int((time.time() - stage_start) * 1000)

        pipeline_stats["stages"].append({
            "stage": "transformation",
            "duration_ms": stage_time,
            "records_processed": transformation_result["transformed_count"]
        })

        # Stage 4: Data Aggregation (optional)
        aggregation_config = processing_request.get("processing_config", {}).get("aggregation")
        aggregated_data = None

        if aggregation_config:
            stage_start = time.time()
            transformed_records = transformation_result["transformed_records"]

            aggregation_result = aggregate_data_records(transformed_records, aggregation_config)
            stage_time = int((time.time() - stage_start) * 1000)

            pipeline_stats["stages"].append({
                "stage": "aggregation",
                "duration_ms": stage_time,
                "groups_created": aggregation_result["group_count"]
            })

            aggregated_data = aggregation_result

        # Stage 5: Data Export
        stage_start = time.time()
        export_config = {
            "format": processing_request["output_format"],
            "include_metadata": True
        }

        export_data = aggregated_data if aggregated_data else transformation_result
        export_result = export_processed_data(export_data, export_config)
        stage_time = int((time.time() - stage_start) * 1000)

        pipeline_stats["stages"].append({
            "stage": "export",
            "duration_ms": stage_time,
            "format": export_result["format"]
        })

        # Calculate total time
        total_time = int((time.time() - start_time) * 1000)
        pipeline_stats["total_time_ms"] = total_time

        return {
            "success": True,
            "batch_id": processing_request["batch_id"],
            "pipeline_stats": pipeline_stats,
            "export_result": export_result,
            "record_counts": {
                "ingested": ingestion_result["record_count"],
                "cleaned": cleaning_result["cleaned_count"],
                "transformed": transformation_result["transformed_count"],
                "aggregated_groups": aggregated_data["group_count"] if aggregated_data else 0
            }
        }

    except Exception as e:
        total_time = int((time.time() - start_time) * 1000)
        pipeline_stats["total_time_ms"] = total_time

        return {
            "success": False,
            "error": str(e),
            "pipeline_stats": pipeline_stats
        }
```

## Step 8: Testing and Example Usage

Create a test script and example usage:

```python
# test_pipeline.py
import json
from data_pipeline import process_data_pipeline

# Sample data for testing
test_batch = {
    "batch_id": "batch-001",
    "records": [
        {
            "id": "rec-001",
            "timestamp": "2024-01-01T10:00:00Z",
            "source": "api",
            "data": {
                "user_id": "user-123",
                "name": "John Doe",
                "email": "john@example.com",
                "age": 30,
                "city": "New York",
                "revenue": 1500.50,
                "category": "premium"
            }
        },
        {
            "id": "rec-002",
            "timestamp": "2024-01-01T11:00:00Z",
            "source": "api",
            "data": {
                "user_id": "user-456",
                "name": "Jane Smith",
                "email": "jane@example.com",
                "age": 25,
                "city": "Los Angeles",
                "revenue": 899.99,
                "category": "standard"
            }
        },
        {
            "id": "rec-003",
            "timestamp": "2024-01-01T12:00:00Z",
            "source": "api",
            "data": {
                "user_id": "user-789",
                "name": "Bob Johnson",
                "email": "bob@example.com",
                "age": 35,
                "city": "New York",
                "revenue": 2200.00,
                "category": "premium"
            }
        }
    ],
    "processing_config": {
        "cleaning": {
            "remove_duplicates": True,
            "normalize_text": True,
            "validate_emails": True,
            "email_fields": ["email"]
        },
        "transformations": [
            {
                "type": "add_field",
                "field": "processed_at",
                "value": "2024-01-01T13:00:00Z"
            },
            {
                "type": "calculate_field",
                "field": "revenue_tier",
                "expression": "'high' if data['revenue'] > 1500 else 'medium' if data['revenue'] > 800 else 'low'"
            }
        ],
        "aggregation": {
            "group_by": ["city", "category"],
            "aggregations": [
                {"type": "count", "field": "user_id", "output_field": "user_count"},
                {"type": "sum", "field": "revenue", "output_field": "total_revenue"},
                {"type": "avg", "field": "age", "output_field": "avg_age"}
            ]
        }
    },
    "output_format": "json"
}

if __name__ == "__main__":
    # Process the batch
    result = process_data_pipeline(test_batch)

    print("Pipeline Processing Result:")
    print(json.dumps(result, indent=2, default=str))
```

## Step 9: Build and Deploy

Build and deploy the pipeline:

```bash
# Build all pipeline functions
flypy build data_*.py

# Test locally
flypy local data_pipeline.py process-data-pipeline --port 8080

# Deploy to FunctionFly
flypy deploy ./dist/process-data-pipeline --token YOUR_TOKEN --app-id YOUR_APP_ID
```

## Summary

This tutorial demonstrated how to build a complete data processing pipeline using FlyPy with:

- **Modular functions** for each processing stage
- **Proper error handling** and validation
- **Configurable transformations** and aggregations
- **Multiple output formats** (JSON, CSV)
- **Performance monitoring** with timing metrics

The pipeline architecture makes it easy to:
- Test individual components
- Scale processing stages independently
- Add new transformation types
- Monitor and debug processing issues

This approach is ideal for ETL processes, data analytics pipelines, and any scenario requiring reliable, deterministic data processing.