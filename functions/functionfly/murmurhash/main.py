def _murmurhash3_x86_32(data: bytes, seed: int = 0) -> int:
    """Pure-Python MurmurHash3 (x86_32)."""
    length = len(data)
    h = seed & 0xFFFFFFFF
    c1, c2 = 0xcc9e2d51, 0x1b873593
    for block_start in range(0, length - length % 4, 4):
        k = (data[block_start + 3] << 24 | data[block_start + 2] << 16 |
             data[block_start + 1] << 8 | data[block_start])
        k = (k * c1) & 0xFFFFFFFF
        k = ((k << 15) | (k >> 17)) & 0xFFFFFFFF
        k = (k * c2) & 0xFFFFFFFF
        h ^= k
        h = ((h << 13) | (h >> 19)) & 0xFFFFFFFF
        h = (h * 5 + 0xe6546b64) & 0xFFFFFFFF
    tail = data[length - length % 4:]
    k = 0
    if len(tail) >= 3:
        k ^= tail[2] << 16
    if len(tail) >= 2:
        k ^= tail[1] << 8
    if len(tail) >= 1:
        k ^= tail[0]
        k = (k * c1) & 0xFFFFFFFF
        k = ((k << 15) | (k >> 17)) & 0xFFFFFFFF
        k = (k * c2) & 0xFFFFFFFF
        h ^= k
    h ^= length
    h ^= (h >> 16)
    h = (h * 0x85ebca6b) & 0xFFFFFFFF
    h ^= (h >> 13)
    h = (h * 0xc2b2ae35) & 0xFFFFFFFF
    h ^= (h >> 16)
    return h


def handler(event):
    data = event.get("data") if isinstance(event, dict) else None
    seed = event.get("seed", 0)

    if data is None:
        return {"ok": False, "error": "data is required"}

    try:
        import mmh3
        if isinstance(data, (dict, list)):
            import json
            raw = json.dumps(data).encode("utf-8")
        else:
            raw = str(data).encode("utf-8")
        value = mmh3.hash(raw, seed=int(seed), signed=False)
        return {"ok": True, "result": f"{value:08x}", "decimal": value, "library": "mmh3"}
    except ImportError:
        pass

    try:
        if isinstance(data, (dict, list)):
            import json
            raw = json.dumps(data).encode("utf-8")
        else:
            raw = str(data).encode("utf-8")
        value = _murmurhash3_x86_32(raw, seed=int(seed))
        return {"ok": True, "result": f"{value:08x}", "decimal": value, "library": "pure-python-x86_32"}
    except Exception as e:
        return {"ok": False, "error": str(e)}
