import base64
import os


def handler(event):
    plaintext = event.get("plaintext") if isinstance(event, dict) else None
    key = event.get("key")
    iv = event.get("iv")

    if plaintext is None:
        return {"ok": False, "error": "plaintext is required"}
    if not key:
        return {"ok": False, "error": "key is required (hex or base64 encoded AES key)"}

    try:
        from cryptography.hazmat.primitives.ciphers import Cipher, algorithms, modes
        from cryptography.hazmat.primitives import padding

        # Decode key
        try:
            key_bytes = bytes.fromhex(str(key))
        except ValueError:
            key_bytes = base64.b64decode(str(key))
        if len(key_bytes) not in (16, 24, 32):
            return {"ok": False, "error": f"key must be 16, 24, or 32 bytes (got {len(key_bytes)})"}

        # Decode or generate IV
        if iv:
            try:
                iv_bytes = bytes.fromhex(str(iv))
            except ValueError:
                iv_bytes = base64.b64decode(str(iv))
        else:
            iv_bytes = os.urandom(16)
        if len(iv_bytes) != 16:
            return {"ok": False, "error": "IV must be 16 bytes for AES-CBC"}

        # Encode plaintext
        if isinstance(plaintext, (dict, list)):
            import json
            pt_bytes = json.dumps(plaintext).encode("utf-8")
        else:
            pt_bytes = str(plaintext).encode("utf-8")

        # PKCS7 padding
        padder = padding.PKCS7(128).padder()
        padded = padder.update(pt_bytes) + padder.finalize()

        cipher = Cipher(algorithms.AES(key_bytes), modes.CBC(iv_bytes))
        encryptor = cipher.encryptor()
        ciphertext = encryptor.update(padded) + encryptor.finalize()

        return {
            "ok": True,
            "result": base64.b64encode(ciphertext).decode("utf-8"),
            "iv": iv_bytes.hex(),
            "iv_base64": base64.b64encode(iv_bytes).decode("utf-8"),
        }
    except ImportError:
        return {"ok": False, "error": "cryptography library is not installed. Install with: pip install cryptography"}
    except Exception as e:
        return {"ok": False, "error": str(e)}
