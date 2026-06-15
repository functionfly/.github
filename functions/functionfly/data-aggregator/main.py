import statistics
from typing import Any


def aggregate_dataset(data: list, agg_type: str) -> float | None:
    if not data:
        return None
    
    clean_data = [x for x in data if isinstance(x, (int, float)) and x is not None]
    
    if not clean_data:
        return None
    
    agg_type = agg_type.lower().strip()
    
    if agg_type == "sum":
        return sum(clean_data)
    elif agg_type == "average" or agg_type == "avg" or agg_type == "mean":
        return round(statistics.mean(clean_data), 4)
    elif agg_type == "count":
        return float(len(clean_data))
    elif agg_type == "min":
        return min(clean_data)
    elif agg_type == "max":
        return max(clean_data)
    elif agg_type == "median":
        return statistics.median(clean_data)
    elif agg_type == "stdev" or agg_type == "std":
        if len(clean_data) < 2:
            return 0.0
        return round(statistics.stdev(clean_data), 4)
    elif agg_type == "variance":
        if len(clean_data) < 2:
            return 0.0
        return round(statistics.variance(clean_data), 4)
    elif agg_type == "product":
        result = 1
        for x in clean_data:
            result *= x
        return result
    else:
        return None


def generate_dataset_summary(datasets: list, results: dict) -> dict:
    total_records = 0
    total_values = 0
    datasets_with_data = 0
    
    for ds in datasets:
        if "data_array" in ds and isinstance(ds["data_array"], list):
            valid_count = len([x for x in ds["data_array"] if isinstance(x, (int, float))])
            total_records += valid_count
            datasets_with_data += 1 if valid_count > 0 else 0
    
    for result in results.values():
        if isinstance(result, (int, float)):
            total_values += result
    
    return {
        "total_datasets": len(datasets),
        "datasets_with_data": datasets_with_data,
        "total_records": total_records,
        "aggregations_computed": len(results),
        "aggregate_total": round(total_values, 4) if isinstance(total_values, float) else total_values
    }


def handler(event: dict[str, Any]) -> dict[str, Any]:
    try:
        datasets = event.get("datasets", [])
        aggregation_type = event.get("aggregation_type", "sum").lower().strip()
        
        if not isinstance(datasets, list):
            return {"ok": False, "error": "datasets must be a list"}
        
        if len(datasets) == 0:
            return {"ok": False, "error": "datasets cannot be empty"}
        
        valid_agg_types = ["sum", "average", "avg", "mean", "count", "min", "max", "median", "stdev", "std", "variance", "product"]
        if aggregation_type not in valid_agg_types:
            return {"ok": False, "error": f"aggregation_type must be one of: {', '.join(valid_agg_types)}"}
        
        results = {}
        
        for i, ds in enumerate(datasets):
            if not isinstance(ds, dict):
                return {"ok": False, "error": f"Each dataset must be an object with name and data_array, got type {type(ds).__name__} at index {i}"}
            
            name = ds.get("name", f"dataset_{i}")
            data_array = ds.get("data_array", [])
            
            if not isinstance(data_array, list):
                return {"ok": False, "error": f"data_array for '{name}' must be a list"}
            
            result = aggregate_dataset(data_array, aggregation_type)
            
            if result is None and len(data_array) > 0:
                return {"ok": False, "error": f"Could not compute {aggregation_type} for '{name}' - no numeric values found"}
            
            results[name] = result
        
        dataset_summary = generate_dataset_summary(datasets, results)
        
        return {
            "ok": True,
            "results": results,
            "dataset_summary": dataset_summary,
            "aggregation_type": aggregation_type
        }
        
    except Exception as e:
        return {"ok": False, "error": str(e)}
