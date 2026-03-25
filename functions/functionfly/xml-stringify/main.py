import xml.etree.ElementTree as ET


def _dict_to_elem(tag, data):
    elem = ET.Element(tag)
    if isinstance(data, dict):
        if "@attributes" in data:
            for k, v in data["@attributes"].items():
                elem.set(str(k), str(v))
        if "#text" in data:
            elem.text = str(data["#text"])
        for k, v in data.items():
            if k.startswith("@") or k == "#text":
                continue
            if isinstance(v, list):
                for item in v:
                    elem.append(_dict_to_elem(k, item))
            else:
                elem.append(_dict_to_elem(k, v))
    else:
        elem.text = str(data)
    return elem


def handler(event):
    data = event.get("data") if isinstance(event, dict) else None
    indent = event.get("indent", True)
    if not data or not isinstance(data, dict):
        return {"ok": False, "error": "data must be an object with a root element"}
    try:
        keys = list(data.keys())
        if len(keys) != 1:
            return {"ok": False, "error": "data must have exactly one root key"}
        root_tag = keys[0]
        root = _dict_to_elem(root_tag, data[root_tag])
        if indent:
            ET.indent(root)
        result = ET.tostring(root, encoding="unicode", xml_declaration=False)
        return {"ok": True, "result": result}
    except Exception as e:
        return {"ok": False, "error": str(e)}
