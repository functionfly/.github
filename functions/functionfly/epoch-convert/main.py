from datetime import datetime

def handler(event):
    try:
        epoch_seconds = event.get("epoch_seconds") if isinstance(event, dict) else None
        epoch_millis = event.get("epoch_millis") if isinstance(event, dict) else None
        date_str = event.get("date") if isinstance(event, dict) else None
        if epoch_seconds is not None:
            try:
                dt = datetime.fromtimestamp(epoch_seconds)
                return {"ok": True, "epoch_seconds": epoch_seconds, "epoch_millis": int(epoch_seconds * 1000), "date": dt.isoformat() + "Z"}
            except Exception as e:
                return {"ok": False, "error": f"invalid epoch_seconds: {str(e)}"}
        elif epoch_millis is not None:
            try:
                dt = datetime.fromtimestamp(epoch_millis / 1000)
                return {"ok": True, "epoch_seconds": epoch_millis / 1000, "epoch_millis": epoch_millis, "date": dt.isoformat() + "Z"}
            except Exception as e:
                return {"ok": False, "error": f"invalid epoch_millis: {str(e)}"}
        elif date_str:
            try:
                dt = datetime.fromisoformat(date_str.replace("Z", "+00:00"))
                epoch_sec = dt.timestamp()
                return {"ok": True, "epoch_seconds": epoch_sec, "epoch_millis": int(epoch_sec * 1000), "date": dt.isoformat() + "Z"}
            except ValueError:
                return {"ok": False, "error": "invalid date format"}
        else:
            dt = datetime.now()
            epoch_sec = dt.timestamp()
            return {"ok": True, "epoch_seconds": epoch_sec, "epoch_millis": int(epoch_sec * 1000), "date": dt.isoformat() + "Z"}
    except Exception as e:
        return {"ok": False, "error": str(e)}
