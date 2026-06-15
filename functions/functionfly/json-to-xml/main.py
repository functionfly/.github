import json
import xml.etree.ElementTree as ET
from xml.sax.saxutils import escape, quoteattr


def handler(event):
    if isinstance(event, dict):
        data = event.get("data", "")
        root_element = event.get("root_element", "root")
    else:
        data = ""
        root_element = "root"

    if not data:
        return {"ok": False, "error": "data is required"}

    if not isinstance(root_element, str) or not root_element:
        return {"ok": False, "error": "root_element must be a non-empty string"}

    if root_element.isdigit():
        return {"ok": False, "error": "root_element cannot be purely numeric"}

    import re
    if not re.match(r'^[a-zA-Z_][a-zA-Z0-9_\-\.]*$', root_element):
        return {"ok": False, "error": "root_element must be a valid XML element name"}

    try:
        if isinstance(data, str):
            parsed = json.loads(data)
        else:
            parsed = data

        def json_to_xml_value(value, element_name):
            elem = ET.Element(element_name)
            if isinstance(value, dict):
                for k, v in value.items():
                    if isinstance(k, str):
                        child = json_to_xml_value(v, k)
                        elem.append(child)
                    else:
                        child = json_to_xml_value(v, str(k))
                        elem.append(child)
            elif isinstance(value, list):
                for i, item in enumerate(value):
                    child = json_to_xml_value(item, "item")
                    elem.append(child)
            elif value is None or value == "":
                pass
            elif isinstance(value, bool):
                elem.text = "true" if value else "false"
            elif isinstance(value, (int, float)):
                elem.text = str(value)
            else:
                elem.text = escape(str(value))
            return elem

        if isinstance(parsed, dict):
            root = json_to_xml_value(parsed, root_element)
        elif isinstance(parsed, list):
            root = ET.Element(root_element)
            for i, item in enumerate(parsed):
                child = json_to_xml_value(item, "item")
                root.append(child)
        elif isinstance(parsed, (int, float, bool)):
            root = ET.Element(root_element)
            if isinstance(parsed, bool):
                root.text = "true" if parsed else "false"
            else:
                root.text = str(parsed)
        else:
            root = ET.Element(root_element)
            root.text = escape(str(parsed))

        xml_str = ET.tostring(root, encoding="unicode")
        return {"ok": True, "result": xml_str}
    except json.JSONDecodeError as e:
        return {"ok": False, "error": f"invalid JSON: {str(e)}"}
    except ET.ParseError as e:
        return {"ok": False, "error": f"XML parsing error: {str(e)}"}
    except Exception as e:
        return {"ok": False, "error": str(e)}
