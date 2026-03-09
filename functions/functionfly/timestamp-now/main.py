import time


def handler(event):
    return {"ok": True, "timestamp": int(time.time())}
