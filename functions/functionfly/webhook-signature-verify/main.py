import hmac
import hashlib


def handler(event):
    payload = event.get("payload") if isinstance(event, dict) else None
    secret = event.get("secret")
    signature = event.get("signature")
    algorithm = event.get("algorithm", "sha256")
    prefix = event.get("prefix", "")

    if payload is None:
        return {"ok": False, "error": "payload is required"}
    if not secret:
        return {"ok": False, "error": "secret is required"}
    if not signature:
        return {"ok": False, "error": "signature is required"}
    if not isinstance(secret, str):
        return {"ok": False, "error": "secret must be a string"}
    if not isinstance(signature, str):
        return {"ok": False, "error": "signature must be a string"}

    algorithm = algorithm.lower()

    # Normalize payload to bytes
    if isinstance(payload, (dict, list)):
        import json
        payload_bytes = json.dumps(payload, separators=(',', ':')).encode('utf-8')
    elif isinstance(payload, str):
        payload_bytes = payload.encode('utf-8')
    elif isinstance(payload, bytes):
        payload_bytes = payload
    else:
        payload_bytes = str(payload).encode('utf-8')

    # Strip prefix from provided signature (e.g. "sha256=abc123" → "abc123")
    sig_to_check = signature
    if prefix and sig_to_check.lower().startswith(prefix.lower()):
        sig_to_check = sig_to_check[len(prefix):]

    # Map algorithm
    algo_map = {
        "sha256": hashlib.sha256,
        "sha1": hashlib.sha1,
        "sha512": hashlib.sha512,
        "md5": hashlib.md5,
    }
    hash_fn = algo_map.get(algorithm)
    if not hash_fn:
        return {"ok": False, "error": f"Unsupported algorithm: {algorithm}. Supported: {list(algo_map.keys())}"}

    expected_digest = hmac.new(
        secret.encode('utf-8'),
        payload_bytes,
        hash_fn
    ).hexdigest()

    expected_sig = f"{prefix}{expected_digest}" if prefix else expected_digest
    valid = hmac.compare_digest(sig_to_check.lower(), expected_digest.lower())

    return {
        "ok": True,
        "valid": valid,
        "algorithm": algorithm,
        "expected": expected_sig,
        "provided": signature,
    }
