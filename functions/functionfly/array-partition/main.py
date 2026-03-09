def handler(event):
    items = event.get("items") if isinstance(event, dict) else None
    key = event.get("key")
    value = event.get("value")
    if items is None:
        return {"ok": False, "error": "items is required"}
    if not isinstance(items, (list, tuple)):
        return {"ok": False, "error": "items must be an array"}
    pass_list = []
    fail_list = []
    if key is not None and isinstance(key, str):
        for x in items:
            if isinstance(x, dict) and key in x:
                v = x[key]
                if value is not None:
                    if v == value:
                        pass_list.append(x)
                    else:
                        fail_list.append(x)
                else:
                    if v:
                        pass_list.append(x)
                    else:
                        fail_list.append(x)
            else:
                fail_list.append(x)
    else:
        for x in items:
            if x:
                pass_list.append(x)
            else:
                fail_list.append(x)
    return {"ok": True, "pass": pass_list, "fail": fail_list}
