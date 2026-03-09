import json
import csv
import io


def handler(event):
    """
    Convert CSV string to JSON array.

    Input:
        - csv: CSV string (required)
        - delimiter: Optional delimiter character (default: comma)
        - has_header: Whether first row is header (default: true)
        - skip_blank_lines: Skip blank lines (default: true)
        - trim_fields: Trim whitespace from fields (default: true)
        - null_value: Value to interpret as null (default: empty string)

    Returns:
        - data: JSON array of objects
        - row_count: Number of data rows
        - column_count: Number of columns
    """
    # Parse input
    if isinstance(event, dict):
        csv_string = event.get("csv", "")
        delimiter = event.get("delimiter", ",")
        has_header = event.get("has_header", True)
        skip_blank_lines = event.get("skip_blank_lines", True)
        trim_fields = event.get("trim_fields", True)
        null_value = event.get("null_value", "")
    else:
        # Try parsing as JSON if string
        try:
            parsed = json.loads(str(event))
            if isinstance(parsed, dict):
                csv_string = parsed.get("csv", "")
                delimiter = parsed.get("delimiter", ",")
                has_header = parsed.get("has_header", True)
                skip_blank_lines = parsed.get("skip_blank_lines", True)
                trim_fields = parsed.get("trim_fields", True)
                null_value = parsed.get("null_value", "")
            else:
                csv_string = str(parsed)
                delimiter = ","
                has_header = True
                skip_blank_lines = True
                trim_fields = True
                null_value = ""
        except:
            csv_string = str(event)
            delimiter = ","
            has_header = True
            skip_blank_lines = True
            trim_fields = True
            null_value = ""

    # Validate input
    if not csv_string or not csv_string.strip():
        return {
            "ok": False,
            "error": "Input 'csv' is required and cannot be empty"
        }

    try:
        input_stream = io.StringIO(csv_string)
        reader = csv.reader(input_stream, delimiter=delimiter)

        rows = list(reader)

        if len(rows) == 0:
            return {
                "ok": True,
                "data": [],
                "row_count": 0,
                "column_count": 0
            }

        # Handle blank lines
        if skip_blank_lines:
            rows = [row for row in rows if any(cell.strip() for cell in row)]

        if len(rows) == 0:
            return {
                "ok": True,
                "data": [],
                "row_count": 0,
                "column_count": 0
            }

        # Determine columns
        if has_header:
            columns = rows[0]
            data_rows = rows[1:]
        else:
            # Use column names like col_0, col_1, etc.
            columns = [f"col_{i}" for i in range(len(rows[0]))]
            data_rows = rows

        # Trim fields if requested
        if trim_fields:
            columns = [col.strip() for col in columns]
            data_rows = [[cell.strip() for cell in row] for row in data_rows]

        # Convert to JSON objects
        data = []
        for row in data_rows:
            obj = {}
            for i, col in enumerate(columns):
                if i < len(row):
                    value = row[i]
                    # Convert null values
                    if value == null_value or (not value and null_value == ""):
                        obj[col] = None
                    else:
                        # Try to convert numbers
                        if value:
                            # Try integer
                            try:
                                obj[col] = int(value)
                                continue
                            except ValueError:
                                pass
                            # Try float
                            try:
                                obj[col] = float(value)
                                continue
                            except ValueError:
                                pass
                            # Try boolean
                            if value.lower() in ('true', 'false'):
                                obj[col] = value.lower() == 'true'
                                continue
                            # Keep as string
                            obj[col] = value
                        else:
                            obj[col] = None
                else:
                    obj[col] = None
            data.append(obj)

        return {
            "ok": True,
            "data": data,
            "row_count": len(data),
            "column_count": len(columns)
        }

    except Exception as e:
        return {
            "ok": False,
            "error": f"Failed to parse CSV: {str(e)}"
        }
