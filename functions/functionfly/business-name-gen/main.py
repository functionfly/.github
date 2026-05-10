import json

def handler(event):
    if isinstance(event, dict):
        data = event.get("data", "")
    else:
        data = ""
    
    # TODO: Implement Business Name Generator logic
    result = {"ok": True, "result": data, "tier": "pro"}
    return result
