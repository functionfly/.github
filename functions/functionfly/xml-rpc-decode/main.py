import xmlrpc.client


def handler(event):
    data = event.get("data") if isinstance(event, dict) else None
    if not data:
        return {"ok": False, "error": "data is required (XML-RPC request/response string)"}
    try:
        params, method = xmlrpc.client.loads(str(data))
        return {"ok": True, "result": {"params": list(params), "method": method}}
    except Exception as e:
        return {"ok": False, "error": str(e)}
