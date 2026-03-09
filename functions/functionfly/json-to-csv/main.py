import csv
import io
import json

def handler(event):
    """
    Convert a JSON array of objects to CSV format.

    Args:
        event: A list of objects (dicts) to convert to CSV, or a JSON string

    Returns:
        dict with 'csv' containing the CSV string and 'rows' count
    """
    # If input is a string, try to parse it as JSON
    # If input is a dict with "data", use that (manifest/API shape)
    data = event
    if isinstance(event, str):
        data = json.loads(event)
    elif isinstance(event, dict) and "data" in event:
        data = event["data"]

    # Input must be a list
    if not isinstance(data, list):
        return {"error": "Input must be a list of objects"}

    # Handle empty list
    if not data:
        return {"csv": "", "rows": 0}

    # Get field names from first object
    if not isinstance(data[0], dict):
        return {"error": "Array items must be objects"}

    fieldnames = list(data[0].keys())

    # Create CSV
    output = io.StringIO()
    writer = csv.DictWriter(output, fieldnames=fieldnames)
    writer.writeheader()
    for row in data:
        writer.writerow(row)

    return {"csv": output.getvalue(), "rows": len(data)}
