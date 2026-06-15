import csv
import io
import json


def handler(event):
    if isinstance(event, dict):
        data = event.get("data", "")
        has_header = event.get("has_header", True)
        delimiter = event.get("delimiter", ",")
    else:
        data = ""
        has_header = True
        delimiter = ","

    if not data:
        return {"ok": False, "error": "data is required"}

    if not isinstance(data, str):
        return {"ok": False, "error": "data must be a string"}

    if not isinstance(delimiter, str) or len(delimiter) != 1:
        return {"ok": False, "error": "delimiter must be a single character"}

    if delimiter in "\r\n":
        return {"ok": False, "error": "delimiter cannot be a newline character"}

    try:
        reader = csv.reader(io.StringIO(data), delimiter=delimiter)
        rows = list(reader)

        if not rows:
            return {"ok": True, "result": []}

        if has_header:
            if len(rows) == 1:
                return {"ok": True, "result": []}
            header = rows[0]
            result = []
            for row in rows[1:]:
                if len(row) != len(header):
                    if len(row) < len(header):
                        row = row + [""] * (len(header) - len(row))
                    else:
                        row = row[:len(header)]
                    row = list(row)
                obj = {header[i]: row[i] for i in range(len(header))}
                result.append(obj)
        else:
            result = [list(row) for row in rows]

        return {"ok": True, "result": result}
    except csv.Error as e:
        return {"ok": False, "error": f"CSV parsing error: {str(e)}"}
    except Exception as e:
        return {"ok": False, "error": str(e)}
