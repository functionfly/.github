import secrets
import string
import uuid
import base64


def handler(event):
    """Generate a new secret value."""
    try:
        secret_type = event.get("type", "random")
        length = int(event.get("length", 32))
        prefix = event.get("prefix", "")

        if secret_type == "uuid":
            secret = str(uuid.uuid4())
        elif secret_type == "hex":
            secret = secrets.token_hex(length // 2)
        elif secret_type == "base64":
            secret = base64.urlsafe_b64encode(secrets.token_bytes(length)).decode().rstrip("=")
        elif secret_type == "alphanumeric":
            alphabet = string.ascii_letters + string.digits
            secret = "".join(secrets.choice(alphabet) for _ in range(length))
        else:  # random (URL-safe)
            secret = secrets.token_urlsafe(length)

        if prefix:
            secret = f"{prefix}{secret}"

        return {"ok": True, "result": secret, "secret": secret}
    except Exception as e:
        return {"ok": False, "error": str(e)}
