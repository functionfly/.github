import base64


def handler(event):
    ciphertext = event.get("ciphertext") if isinstance(event, dict) else None
    key = event.get("key")
    nonce = event.get("nonce")
    aad = event.get("aad")
    encoding = event.get("encoding", "utf-8")

    if not ciphertext:
        return {"ok": False, "error": "ciphertext is required"}
    if not key:
        return {"ok": False, "error": "key is required"}
    if not nonce:
        return {"ok": False, "error": "nonce is required"}

    try:
        from cryptography.hazmat.primitives.ciphers.aead import AESGCM

        try:
            key_bytes = bytes.fromhex(str(key))
        except ValueError:
            key_bytes = base64.b64decode(str(key))
        try:
            nonce_bytes = bytes.fromhex(str(nonce))
        except ValueError:
            nonce_bytes = base64.b64decode(str(nonce))

        ct_bytes = base64.b64decode(str(ciphertext))
        aad_bytes = str(aad).encode("utf-8") if aad else None

        aesgcm = AESGCM(key_bytes)
        pt_bytes = aesgcm.decrypt(nonce_bytes, ct_bytes, aad_bytes)

        try:
            result = pt_bytes.decode(encoding)
            is_text = True
        except UnicodeDecodeError:
            result = base64.b64encode(pt_bytes).decode("utf-8")
            is_text = False

        return {"ok": True, "result": result, "is_text": is_text}
    except ImportError:
        return {"ok": False, "error": "cryptography library is not installed. Install with: pip install cryptography"}
    except Exception as e:
        return {"ok": False, "error": str(e)}
