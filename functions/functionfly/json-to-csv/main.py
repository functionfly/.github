import csv
import io
import json


def handler(event):
    if isinstance(event, dict):
        data = event.get("data", "")
        include_header = event.get("include_header", True)
        delimiter = event.get("delimiter", ",")
    else:
        data = ""
        include_header = True
        delimiter = ","

    if not data:
        return {"ok": False, "error": "data is required"}

    if not isinstance(delimiter, str) or len(delimiter) != 1:
        return {"ok": False, "error": "delimiter must be a single character"}

    if delimiter in "\r\n":
        return {"ok": False, "error": "delimiter cannot be a newline character"}

    try:
        if isinstance(data, str):
            parsed = json.loads(data)
        else:
            parsed = data

        if not isinstance(parsed, list):
            return {"ok": False, "error": "data must be a JSON array"}

        if not parsed:
            return {"ok": True, "result": ""}

        output = io.StringIO()
        writer = csv.writer(output, delimiter=delimiter)

        if include_header:
            first_item = parsed[0]
            if isinstance(first_item, dict):
                headers = list(first_item.keys())
                writer.writerow(headers)

        for item in parsed:
            if isinstance(item, dict):
                if include_header:
                    row = [item.get(h, "") for h in headers]
                else:
                    row = list(item.values())
            elif isinstance(item, (list, tuple)):
                row = list(item)
            else:
                row = [str(item)]
            writer.writerow(row)

        result = output.getvalue()
        if result.endswith("\n"):
            result = result[:-1]
        return {"ok": True, "result": result}
    except json.JSONDecodeError as e:
        return {"ok": False, "error": f"invalid JSON: {str(e)}"}
    except csv.Error as e:
        return {"ok": False, "error": f"CSV writing error: {str(e)}"}
    except NameError:
        return {"ok": False, "error": "data is required and must be a JSON array"}
    except Exception as e:
        return {"ok": False, "error": str(e)}
