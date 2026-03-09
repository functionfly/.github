import base64


def _decode_key(key):
    if isinstance(key, bytes):
        return key
    s = key.strip()
    if len(s) in (32, 48, 64) and all(c in "0123456789abcdefABCDEF" for c in s):
        return bytes.fromhex(s)
    try:
        return base64.b64decode(s)
    except Exception:
        return s.encode("utf-8")[:32].ljust(32, b"\0")


def handler(event):
    if isinstance(event, dict):
        ciphertext = event.get("ciphertext", "")
        key = event.get("key", "")
        iv = event.get("iv", "")
        tag = event.get("tag")
        mode = (event.get("mode") or "GCM").upper()
    else:
        ciphertext, key, iv, tag, mode = "", "", "", None, "GCM"

    if not key:
        return {"ok": False, "error": "Input 'key' is required"}
    if not ciphertext:
        return {"ok": False, "error": "Input 'ciphertext' is required"}

    try:
        from cryptography.hazmat.primitives.ciphers.aead import AESGCM
        from cryptography.hazmat.primitives.ciphers import Cipher, algorithms, modes
        from cryptography.hazmat.primitives import padding
    except ImportError:
        return {"ok": False, "error": "cryptography required: pip install cryptography"}

    key_bytes = _decode_key(key)
    if len(key_bytes) not in (16, 24, 32):
        key_bytes = key_bytes[:32].ljust(32, b"\0")
    try:
        ct = base64.b64decode(ciphertext, validate=True)
    except Exception as e:
        return {"ok": False, "error": f"Invalid base64 ciphertext: {e}"}

    try:
        if mode == "GCM":
            if not iv:
                return {"ok": False, "error": "IV is required for GCM decryption"}
            nonce = base64.b64decode(iv)
            aes = AESGCM(key_bytes)
            plain = aes.decrypt(nonce, ct, None)
            return {"ok": True, "data": plain.decode("utf-8", errors="replace")}
        else:
            if not iv:
                return {"ok": False, "error": "IV is required for CBC decryption"}
            iv_bytes = base64.b64decode(iv)
            cipher = Cipher(algorithms.AES(key_bytes), modes.CBC(iv_bytes))
            dec = cipher.decryptor()
            padded = dec.update(ct) + dec.finalize()
            unpadder = padding.PKCS7(128).unpadder()
            plain = unpadder.update(padded) + unpadder.finalize()
            return {"ok": True, "data": plain.decode("utf-8", errors="replace")}
    except Exception as e:
        return {"ok": False, "error": str(e)}
