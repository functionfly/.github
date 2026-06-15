import csv
import io
import json
import xml.etree.ElementTree as ET
from typing import Any


def parse_csv(data: str) -> list[dict]:
    if not data.strip():
        return []
    
    reader = csv.DictReader(io.StringIO(data))
    return list(reader)


def parse_json(data: str) -> list[dict] | dict:
    if not data.strip():
        return []
    
    parsed = json.loads(data)
    
    if isinstance(parsed, list):
        return parsed
    elif isinstance(parsed, dict):
        if "records" in parsed and isinstance(parsed["records"], list):
            return parsed["records"]
        return [parsed]
    else:
        return []


def parse_xml(data: str) -> list[dict]:
    if not data.strip():
        return []
    
    try:
        root = ET.fromstring(data)
    except ET.ParseError:
        records = []
        for elem in ET.fromstring(f"<root>{data}</root>"):
            records.append(xml_to_dict(elem))
        return records if records else [{"error": "Could not parse XML structure"}]
    
    records = []
    for child in root:
        records.append(xml_to_dict(child))
    
    if not records:
        records.append(xml_to_dict(root))
    
    return records


def xml_to_dict(element: ET.Element) -> dict:
    result = {}
    
    if element.attrib:
        result["@attributes"] = element.attrib
    
    if element.text and element.text.strip():
        if len(element) == 0:
            return element.text.strip()
        result["#text"] = element.text.strip()
    
    for child in element:
        child_data = xml_to_dict(child)
        if child.tag in result:
            if not isinstance(result[child.tag], list):
                result[child.tag] = [result[child.tag]]
            result[child.tag].append(child_data)
        else:
            result[child.tag] = child_data
    
    return result


def to_csv(records: list[dict]) -> str:
    if not records:
        return ""
    
    if not isinstance(records[0], dict):
        records = [{"value": r} for r in records]
    
    output = io.StringIO()
    writer = csv.DictWriter(output, fieldnames=records[0].keys(), extrasaction='ignore')
    writer.writeheader()
    writer.writerows(records)
    
    return output.getvalue()


def to_json(records: list[dict]) -> str:
    if not records:
        return "[]"
    
    return json.dumps(records, indent=2)


def to_xml(records: list[dict], root_name: str = "records") -> str:
    if not records:
        return f"<{root_name}/>"
    
    root = ET.Element(root_name)
    
    for i, record in enumerate(records):
        item = ET.SubElement(root, "record")
        item.set("index", str(i))
        
        if isinstance(record, dict):
            dict_to_xml(record, item)
        else:
            item.text = str(record)
    
    return ET.tostring(root, encoding="unicode")


def dict_to_xml(data: dict, parent: ET.Element) -> None:
    for key, value in data.items():
        if key == "@attributes":
            continue
        
        elem = ET.SubElement(parent, str(key))
        
        if isinstance(value, dict):
            dict_to_xml(value, elem)
        elif isinstance(value, list):
            for item in value:
                list_item = ET.SubElement(elem, "item")
                if isinstance(item, dict):
                    dict_to_xml(item, list_item)
                else:
                    list_item.text = str(item)
        else:
            elem.text = str(value)


def transform_data(data: Any, input_format: str, output_format: str) -> tuple[str, int]:
    input_format = input_format.lower().strip()
    output_format = output_format.lower().strip()
    
    records = []
    
    if input_format == "csv":
        records = parse_csv(data)
    elif input_format == "json":
        if isinstance(data, str):
            records = parse_json(data)
        elif isinstance(data, list):
            records = data
        elif isinstance(data, dict):
            records = [data]
        else:
            records = []
    elif input_format == "xml":
        if isinstance(data, str):
            records = parse_xml(data)
        else:
            records = []
    else:
        raise ValueError(f"Unsupported input format: {input_format}")
    
    record_count = len(records)
    
    if output_format == "csv":
        return to_csv(records), record_count
    elif output_format == "json":
        return to_json(records), record_count
    elif output_format == "xml":
        return to_xml(records), record_count
    else:
        raise ValueError(f"Unsupported output format: {output_format}")


def handler(event: dict[str, Any]) -> dict[str, Any]:
    try:
        data = event.get("data")
        input_format = event.get("input_format", "").lower().strip()
        output_format = event.get("output_format", "").lower().strip()
        
        if data is None:
            return {"ok": False, "error": "data is required"}
        
        if not input_format:
            return {"ok": False, "error": "input_format is required (csv/json/xml)"}
        
        if not output_format:
            return {"ok": False, "error": "output_format is required (csv/json/xml)"}
        
        valid_formats = ["csv", "json", "xml"]
        if input_format not in valid_formats:
            return {"ok": False, "error": f"input_format must be one of: {', '.join(valid_formats)}"}
        
        if output_format not in valid_formats:
            return {"ok": False, "error": f"output_format must be one of: {', '.join(valid_formats)}"}
        
        if isinstance(data, list):
            data_input = data
        elif isinstance(data, str):
            data_input = data
        else:
            return {"ok": False, "error": "data must be a string or list"}
        
        transformed_data, record_count = transform_data(data_input, input_format, output_format)
        
        return {
            "ok": True,
            "transformed_data": transformed_data,
            "record_count": record_count,
            "input_format": input_format,
            "output_format": output_format
        }
        
    except Exception as e:
        return {"ok": False, "error": f"Transformation error: {str(e)}"}
