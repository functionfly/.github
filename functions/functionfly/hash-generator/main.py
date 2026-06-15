import hashlib
import json


def handler(event):
    if isinstance(event, dict):
        data = event.get("data", "")
        algorithm = event.get("algorithm", "sha256")
    else:
        data = ""
        algorithm = "sha256"

    if not data:
        return {"ok": False, "error": "data is required"}

    if not isinstance(data, str):
        return {"ok": False, "error": "data must be a string"}

    algorithm_lower = algorithm.lower()
    valid_algorithms = {"md5", "sha1", "sha256", "sha512"}
    if algorithm_lower not in valid_algorithms:
        return {"ok": False, "error": f"unsupported algorithm: {algorithm}. Supported: {', '.join(sorted(valid_algorithms))}"}

    try:
        data_bytes = data.encode("utf-8")

        if algorithm_lower == "md5":
            hasher = hashlib.md5(data_bytes)
        elif algorithm_lower == "sha1":
            hasher = hashlib.sha1(data_bytes)
        elif algorithm_lower == "sha256":
            hasher = hashlib.sha256(data_bytes)
        elif algorithm_lower == "sha512":
            hasher = hashlib.sha512(data_bytes)

        hash_hex = hasher.hexdigest()
        return {"ok": True, "result": hash_hex}
    except Exception as e:
        return {"ok": False, "error": str(e)}
