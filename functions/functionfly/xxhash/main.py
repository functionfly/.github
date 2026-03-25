def handler(event):
    data = event.get("data") if isinstance(event, dict) else None
    seed = event.get("seed", 0)
    variant = event.get("variant", "xxh64")

    if data is None:
        return {"ok": False, "error": "data is required"}

    try:
        import xxhash as _xxhash
        if isinstance(data, (dict, list)):
            import json
            raw = json.dumps(data).encode("utf-8")
        else:
            raw = str(data).encode("utf-8")
        if variant == "xxh32":
            h = _xxhash.xxh32(raw, seed=int(seed))
        elif variant == "xxh3_64":
            h = _xxhash.xxh3_64(raw, seed=int(seed))
        elif variant == "xxh3_128":
            h = _xxhash.xxh3_128(raw, seed=int(seed))
        else:
            h = _xxhash.xxh64(raw, seed=int(seed))
        return {"ok": True, "result": h.hexdigest(), "decimal": h.intdigest(), "variant": variant}
    except ImportError:
        # Pure Python FNV-1a fallback (same spirit — fast hash)
        raw = str(data).encode("utf-8")
        val = 14695981039346656037
        for b in raw:
            val ^= b
            val = (val * 1099511628211) & 0xFFFFFFFFFFFFFFFF
        return {"ok": True, "result": f"{val:016x}", "decimal": val, "variant": "fnv1a_64_fallback", "note": "xxhash library not installed; using FNV-1a fallback"}
    except Exception as e:
        return {"ok": False, "error": str(e)}
