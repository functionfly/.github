import base64


def handler(event):
    if isinstance(event, dict):
        ciphertext_b64 = event.get("ciphertext", "")
        private_key_pem = event.get("private_key", "")
        padding_name = (event.get("padding") or "OAEP").upper()
    else:
        ciphertext_b64, private_key_pem, padding_name = "", "", "OAEP"

    if not private_key_pem:
        return {"ok": False, "error": "Input 'private_key' is required"}
    if not ciphertext_b64:
        return {"ok": False, "error": "Input 'ciphertext' is required"}

    try:
        from cryptography.hazmat.primitives import serialization
        from cryptography.hazmat.primitives.asymmetric import padding as rsa_padding
    except ImportError:
        return {"ok": False, "error": "cryptography required: pip install cryptography"}

    try:
        priv = serialization.load_pem_private_key(private_key_pem.encode("utf-8") if isinstance(private_key_pem, str) else private_key_pem, password=None)
    except Exception as e:
        return {"ok": False, "error": f"Invalid private key: {e}"}

    ct = base64.b64decode(ciphertext_b64)
    if padding_name == "OAEP":
        pad = rsa_padding.OAEP(mgf=rsa_padding.MGF1(algorithm=__import__("hashlib").sha256()), algorithm=__import__("hashlib").sha256(), label=None)
    else:
        pad = rsa_padding.PKCS1v15()

    try:
        plain = priv.decrypt(ct, pad)
        return {"ok": True, "data": plain.decode("utf-8", errors="replace")}
    except Exception as e:
        return {"ok": False, "error": str(e)}
