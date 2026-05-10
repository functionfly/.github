import json

def handler(event):
    if isinstance(event, dict):
        data = event.get("data", "")
    else:
        data = ""
    
    # TODO: Implement Tax Calculator logic
    result = {"ok": True, "result": data, "tier": "pro"}
    return result
