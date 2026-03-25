import xml.etree.ElementTree as ET


def _elem_to_dict(elem):
    result = {}
    if elem.attrib:
        result["@attributes"] = dict(elem.attrib)
    if elem.text and elem.text.strip():
        result["#text"] = elem.text.strip()
    for child in elem:
        child_data = _elem_to_dict(child)
        if child.tag in result:
            if not isinstance(result[child.tag], list):
                result[child.tag] = [result[child.tag]]
            result[child.tag].append(child_data)
        else:
            result[child.tag] = child_data
    return result


def handler(event):
    data = event.get("data") if isinstance(event, dict) else None
    if not data:
        return {"ok": False, "error": "data is required (XML string)"}
    try:
        root = ET.fromstring(str(data))
        result = {root.tag: _elem_to_dict(root)}
        return {"ok": True, "result": result, "root": root.tag}
    except Exception as e:
        return {"ok": False, "error": str(e)}
