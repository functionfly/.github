import base64
import os


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
        data = event.get("data", event.get("plaintext", ""))
        key = event.get("key", "")
        iv_b64 = event.get("iv")
        mode = (event.get("mode") or "GCM").upper()
    else:
        data, key, iv_b64, mode = "", "", None, "GCM"

    if not key:
        return {"ok": False, "error": "Input 'key' is required"}
    if data is None:
        return {"ok": False, "error": "Input 'data' is required"}

    try:
        from cryptography.hazmat.primitives.ciphers.aead import AESGCM
        from cryptography.hazmat.primitives.ciphers import Cipher, algorithms, modes
    except ImportError:
        return {"ok": False, "error": "cryptography required: pip install cryptography"}

    key_bytes = _decode_key(key)
    if len(key_bytes) not in (16, 24, 32):
        key_bytes = key_bytes[:32].ljust(32, b"\0")
    payload = data.encode("utf-8") if isinstance(data, str) else data
    if isinstance(payload, str):
        try:
            payload = base64.b64decode(payload)
        except Exception:
            payload = payload.encode("utf-8")

    try:
        if mode == "GCM":
            aes = AESGCM(key_bytes)
            nonce = os.urandom(12)
            ct = aes.encrypt(nonce, payload, None)
            ciphertext = ct
            tag = ct[-16:]
            ciphertext_b64 = base64.b64encode(ct).decode("ascii")
            iv_b64_out = base64.b64encode(nonce).decode("ascii")
            tag_b64 = base64.b64encode(tag).decode("ascii")
            return {"ok": True, "ciphertext": ciphertext_b64, "iv": iv_b64_out, "tag": tag_b64}
        else:
            iv = base64.b64decode(iv_b64) if iv_b64 else os.urandom(16)
            cipher = Cipher(algorithms.AES(key_bytes), modes.CBC(iv))
            enc = cipher.encryptor()
            from cryptography.hazmat.primitives import padding
            pad = padding.PKCS7(128).padder()
            padded = pad.update(payload) + pad.finalize()
            ct = enc.update(padded) + enc.finalize()
            return {"ok": True, "ciphertext": base64.b64encode(ct).decode("ascii"), "iv": base64.b64encode(iv).decode("ascii")}
    except Exception as e:
        return {"ok": False, "error": str(e)}
