import json

def handler(event):
    if isinstance(event, dict):
        data = event.get("data", "")
    else:
        data = ""
    
    # TODO: Implement Discount Calculator logic
    result = {"ok": True, "result": data, "tier": "pro"}
    return result
