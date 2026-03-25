import copy


def _pointer_get(doc, pointer):
    if pointer == "": return doc
    tokens = pointer.lstrip("/").split("/")
    cur = doc
    for t in tokens:
        t = t.replace("~1", "/").replace("~0", "~")
        if isinstance(cur, dict): cur = cur[t]
        elif isinstance(cur, list): cur = cur[int(t)]
        else: raise ValueError(f"cannot traverse {type(cur)}")
    return cur


def _pointer_set(doc, pointer, value):
    tokens = pointer.lstrip("/").split("/")
    tokens = [t.replace("~1", "/").replace("~0", "~") for t in tokens]
    cur = doc
    for t in tokens[:-1]:
        if isinstance(cur, dict): cur = cur[t]
        elif isinstance(cur, list): cur = cur[int(t)]
    last = tokens[-1]
    if isinstance(cur, dict): cur[last] = value
    elif isinstance(cur, list):
        if last == "-": cur.append(value)
        else: cur[int(last)] = value


def _pointer_del(doc, pointer):
    tokens = pointer.lstrip("/").split("/")
    tokens = [t.replace("~1", "/").replace("~0", "~") for t in tokens]
    cur = doc
    for t in tokens[:-1]:
        if isinstance(cur, dict): cur = cur[t]
        elif isinstance(cur, list): cur = cur[int(t)]
    last = tokens[-1]
    if isinstance(cur, dict): del cur[last]
    elif isinstance(cur, list): del cur[int(last)]


def handler(event):
    document = event.get("document") if isinstance(event, dict) else None
    operations = event.get("operations")
    if document is None:
        return {"ok": False, "error": "document is required"}
    if not operations or not isinstance(operations, list):
        return {"ok": False, "error": "operations must be an array of RFC 6902 operations"}
    try:
        try:
            import jsonpatch
            patch = jsonpatch.JsonPatch(operations)
            result = patch.apply(document)
            return {"ok": True, "result": result}
        except ImportError:
            doc = copy.deepcopy(document)
            for op in operations:
                o = op.get("op")
                path = op.get("path", "")
                value = op.get("value")
                if o == "add":
                    _pointer_set(doc, path, copy.deepcopy(value))
                elif o == "remove":
                    _pointer_del(doc, path)
                elif o == "replace":
                    _pointer_set(doc, path, copy.deepcopy(value))
                elif o == "move":
                    from_ = op.get("from")
                    val = copy.deepcopy(_pointer_get(doc, from_))
                    _pointer_del(doc, from_)
                    _pointer_set(doc, path, val)
                elif o == "copy":
                    from_ = op.get("from")
                    val = copy.deepcopy(_pointer_get(doc, from_))
                    _pointer_set(doc, path, val)
                elif o == "test":
                    actual = _pointer_get(doc, path)
                    if actual != value:
                        return {"ok": False, "error": f"test failed at {path}: expected {value!r}, got {actual!r}"}
                else:
                    return {"ok": False, "error": f"unknown op: {o!r}"}
            return {"ok": True, "result": doc, "note": "jsonpatch not installed; using basic implementation"}
    except Exception as e:
        return {"ok": False, "error": str(e)}
