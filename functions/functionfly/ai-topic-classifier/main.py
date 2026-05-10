import json

def handler(event):
    if isinstance(event, dict):
        data = event.get("data", "")
    else:
        data = ""
    
    # TODO: Implement AI Topic Classifier logic
    result = {"ok": True, "result": data, "tier": "premium"}
    return result
