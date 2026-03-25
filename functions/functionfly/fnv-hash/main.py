def handler(event):
    data = event.get("data") if isinstance(event, dict) else None
    variant = event.get("variant", "fnv1a_64")

    if data is None:
        return {"ok": False, "error": "data is required"}

    try:
        if isinstance(data, (dict, list)):
            import json
            raw = json.dumps(data).encode("utf-8")
        else:
            raw = str(data).encode("utf-8")

        if variant in ("fnv1_32", "fnv1a_32"):
            offset = 2166136261
            prime = 16777619
            mask = 0xFFFFFFFF
            width = 32
        else:
            offset = 14695981039346656037
            prime = 1099511628211
            mask = 0xFFFFFFFFFFFFFFFF
            width = 64

        h = offset
        for byte in raw:
            if "1a" in variant:
                h ^= byte
                h = (h * prime) & mask
            else:
                h = (h * prime) & mask
                h ^= byte

        return {"ok": True, "result": f"{h:0{width // 4}x}", "decimal": h, "variant": variant}
    except Exception as e:
        return {"ok": False, "error": str(e)}
