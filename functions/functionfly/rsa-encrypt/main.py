import base64


def handler(event):
    if isinstance(event, dict):
        data = event.get("data", event.get("plaintext", ""))
        public_key_pem = event.get("public_key", "")
        padding_name = (event.get("padding") or "OAEP").upper()
    else:
        data, public_key_pem, padding_name = "", "", "OAEP"

    if not public_key_pem:
        return {"ok": False, "error": "Input 'public_key' is required"}
    if data is None:
        return {"ok": False, "error": "Input 'data' is required"}

    try:
        from cryptography.hazmat.primitives import serialization
        from cryptography.hazmat.primitives.asymmetric import padding as rsa_padding
    except ImportError:
        return {"ok": False, "error": "cryptography required: pip install cryptography"}

    try:
        pub = serialization.load_pem_public_key(public_key_pem.encode("utf-8") if isinstance(public_key_pem, str) else public_key_pem)
    except Exception as e:
        return {"ok": False, "error": f"Invalid public key: {e}"}

    payload = data.encode("utf-8") if isinstance(data, str) else data
    if padding_name == "OAEP":
        pad = rsa_padding.OAEP(mgf=rsa_padding.MGF1(algorithm=__import__("hashlib").sha256()), algorithm=__import__("hashlib").sha256(), label=None)
    else:
        pad = rsa_padding.PKCS1v15()

    try:
        ct = pub.encrypt(payload, pad)
        return {"ok": True, "ciphertext": base64.b64encode(ct).decode("ascii")}
    except Exception as e:
        return {"ok": False, "error": str(e)}
