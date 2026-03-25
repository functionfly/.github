import base64, io


def handler(event):
    data = event.get("data") if isinstance(event, dict) else None
    schema = event.get("schema")
    if data is None:
        return {"ok": False, "error": "data is required"}
    if not schema:
        return {"ok": False, "error": "schema is required (Avro schema dict)"}
    try:
        import fastavro
        buf = io.BytesIO()
        fastavro.writer(buf, fastavro.parse_schema(schema), [data])
        encoded = buf.getvalue()
        return {"ok": True, "result": base64.b64encode(encoded).decode("utf-8"), "bytes": len(encoded)}
    except ImportError:
        return {"ok": False, "error": "fastavro is not installed. Install with: pip install fastavro"}
    except Exception as e:
        return {"ok": False, "error": str(e)}
