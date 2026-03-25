import base64
import os


def handler(event):
    plaintext = event.get("plaintext") if isinstance(event, dict) else None
    key = event.get("key")
    nonce = event.get("nonce")
    aad = event.get("aad")

    if plaintext is None:
        return {"ok": False, "error": "plaintext is required"}
    if not key:
        return {"ok": False, "error": "key is required"}

    try:
        from cryptography.hazmat.primitives.ciphers.aead import AESGCM

        try:
            key_bytes = bytes.fromhex(str(key))
        except ValueError:
            key_bytes = base64.b64decode(str(key))
        if len(key_bytes) not in (16, 24, 32):
            return {"ok": False, "error": f"key must be 16, 24, or 32 bytes (got {len(key_bytes)})"}

        if nonce:
            try:
                nonce_bytes = bytes.fromhex(str(nonce))
            except ValueError:
                nonce_bytes = base64.b64decode(str(nonce))
        else:
            nonce_bytes = os.urandom(12)

        if isinstance(plaintext, (dict, list)):
            import json
            pt_bytes = json.dumps(plaintext).encode("utf-8")
        else:
            pt_bytes = str(plaintext).encode("utf-8")

        aad_bytes = str(aad).encode("utf-8") if aad else None

        aesgcm = AESGCM(key_bytes)
        ct_with_tag = aesgcm.encrypt(nonce_bytes, pt_bytes, aad_bytes)
        ct = ct_with_tag[:-16]
        tag = ct_with_tag[-16:]

        return {
            "ok": True,
            "result": base64.b64encode(ct_with_tag).decode("utf-8"),
            "ciphertext": base64.b64encode(ct).decode("utf-8"),
            "tag": base64.b64encode(tag).decode("utf-8"),
            "nonce": nonce_bytes.hex(),
            "nonce_base64": base64.b64encode(nonce_bytes).decode("utf-8"),
        }
    except ImportError:
        return {"ok": False, "error": "cryptography library is not installed. Install with: pip install cryptography"}
    except Exception as e:
        return {"ok": False, "error": str(e)}
