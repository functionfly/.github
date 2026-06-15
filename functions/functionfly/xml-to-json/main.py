import json
import xml.etree.ElementTree as ET
from xml.parsers.expat import ExpatError


def handler(event):
    if isinstance(event, dict):
        data = event.get("data", "")
    else:
        data = ""

    if not data:
        return {"ok": False, "error": "data is required"}

    if not isinstance(data, str):
        return {"ok": False, "error": "data must be a string"}

    try:
        root = ET.fromstring(data)

        def element_to_dict(element):
            result = {}

            if element.attrib:
                for k, v in element.attrib.items():
                    result[f"@{k}"] = v

            if element.text and element.text.strip():
                if len(element) == 0:
                    return element.text.strip()
                result["#text"] = element.text.strip()

            for child in element:
                child_dict = element_to_dict(child)
                child_tag = child.tag

                if child_tag in result:
                    if not isinstance(result[child_tag], list):
                        result[child_tag] = [result[child_tag]]
                    result[child_tag].append(child_dict)
                else:
                    result[child_tag] = child_dict

            return result

        if len(root) == 0 and not root.attrib and (not root.text or not root.text.strip()):
            result = {"#text": root.text.strip()} if root.text and root.text.strip() else {}
        else:
            result = element_to_dict(root)

            tag = root.tag
            final_result = {tag: result}

        return {"ok": True, "result": final_result}
    except ExpatError as e:
        return {"ok": False, "error": f"invalid XML: {str(e)}"}
    except ET.ParseError as e:
        return {"ok": False, "error": f"XML parsing error: {str(e)}"}
    except Exception as e:
        return {"ok": False, "error": str(e)}
