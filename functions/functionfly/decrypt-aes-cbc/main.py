import base64


def handler(event):
    ciphertext = event.get("ciphertext") if isinstance(event, dict) else None
    key = event.get("key")
    iv = event.get("iv")
    encoding = event.get("encoding", "utf-8")

    if not ciphertext:
        return {"ok": False, "error": "ciphertext is required"}
    if not key:
        return {"ok": False, "error": "key is required"}
    if not iv:
        return {"ok": False, "error": "iv is required"}

    try:
        from cryptography.hazmat.primitives.ciphers import Cipher, algorithms, modes
        from cryptography.hazmat.primitives import padding

        try:
            key_bytes = bytes.fromhex(str(key))
        except ValueError:
            key_bytes = base64.b64decode(str(key))
        try:
            iv_bytes = bytes.fromhex(str(iv))
        except ValueError:
            iv_bytes = base64.b64decode(str(iv))

        ct_bytes = base64.b64decode(str(ciphertext))

        cipher = Cipher(algorithms.AES(key_bytes), modes.CBC(iv_bytes))
        decryptor = cipher.decryptor()
        padded = decryptor.update(ct_bytes) + decryptor.finalize()

        unpadder = padding.PKCS7(128).unpadder()
        pt_bytes = unpadder.update(padded) + unpadder.finalize()

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
