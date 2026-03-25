import base64, io


def handler(event):
    data = event.get("data") if isinstance(event, dict) else None
    if not data:
        return {"ok": False, "error": "data is required (Base64-encoded Avro bytes)"}
    try:
        import fastavro
        raw = base64.b64decode(str(data))
        buf = io.BytesIO(raw)
        reader = fastavro.reader(buf)
        records = list(reader)
        return {"ok": True, "result": records[0] if len(records) == 1 else records, "count": len(records)}
    except ImportError:
        return {"ok": False, "error": "fastavro is not installed. Install with: pip install fastavro"}
    except Exception as e:
        return {"ok": False, "error": str(e)}
