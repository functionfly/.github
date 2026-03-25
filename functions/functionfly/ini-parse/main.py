import configparser, io


def handler(event):
    data = event.get("data") if isinstance(event, dict) else None
    if not data:
        return {"ok": False, "error": "data is required (INI string)"}
    try:
        parser = configparser.ConfigParser()
        parser.read_string(str(data))
        result = {}
        if parser.defaults():
            result["DEFAULT"] = dict(parser.defaults())
        for section in parser.sections():
            result[section] = dict(parser[section])
        return {"ok": True, "result": result}
    except Exception as e:
        return {"ok": False, "error": str(e)}
