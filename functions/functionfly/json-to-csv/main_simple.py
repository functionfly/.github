import json

def handler(event):
    """
    Simple test function that just returns the input data.
    """
    return {"received": event, "type": str(type(event))}
