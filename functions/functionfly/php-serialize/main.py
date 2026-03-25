def _serialize(val):
    if val is None:
        return "N;"
    if isinstance(val, bool):
        return f"b:{1 if val else 0};"
    if isinstance(val, int):
        return f"i:{val};"
    if isinstance(val, float):
        return f"d:{val};"
    if isinstance(val, str):
        b = val.encode("utf-8")
        return f's:{len(b)}:"{val}";'
    if isinstance(val, (list, tuple)):
        parts = "".join(f"{_serialize(i)}{_serialize(v)}" for i, v in enumerate(val))
        return f"a:{len(val)}:{{{parts}}}"
    if isinstance(val, dict):
        parts = "".join(f"{_serialize(k)}{_serialize(v)}" for k, v in val.items())
        return f"a:{len(val)}:{{{parts}}}"
    raise ValueError(f"Unsupported type: {type(val).__name__}")


def handler(event):
    data = event.get("data") if isinstance(event, dict) else None
    if data is None:
        return {"ok": False, "error": "data is required"}
    try:
        result = _serialize(data)
        return {"ok": True, "result": result}
    except Exception as e:
        return {"ok": False, "error": str(e)}
