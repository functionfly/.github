import configparser, io


def handler(event):
    data = event.get("data") if isinstance(event, dict) else None
    if not data or not isinstance(data, dict):
        return {"ok": False, "error": "data must be an object"}
    try:
        parser = configparser.ConfigParser()
        for section, values in data.items():
            if section == "DEFAULT":
                for k, v in values.items():
                    parser.defaults()[k] = str(v)
            elif isinstance(values, dict):
                parser[section] = {str(k): str(v) for k,v in values.items()}
            else:
                return {"ok": False, "error": f"section '{section}' must be a dict"}
        buf = io.StringIO()
        parser.write(buf)
        return {"ok": True, "result": buf.getvalue()}
    except Exception as e:
        return {"ok": False, "error": str(e)}
