import xmlrpc.client


def handler(event):
    method = event.get("method") if isinstance(event, dict) else None
    params = event.get("params", [])
    if not method:
        return {"ok": False, "error": "method is required"}
    try:
        result = xmlrpc.client.dumps(tuple(params), methodname=str(method), encoding="utf-8")
        return {"ok": True, "result": result}
    except Exception as e:
        return {"ok": False, "error": str(e)}
