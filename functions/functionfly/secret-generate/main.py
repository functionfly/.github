import secrets
import string
import uuid
import base64


def generate_password(length, include_symbols):
    alphabet = string.ascii_letters + string.digits
    if include_symbols:
        alphabet += "!@#$%^&*()_+-=[]{}|;:,.<>?"
    # Ensure at least one of each required type
    password = [
        secrets.choice(string.ascii_uppercase),
        secrets.choice(string.ascii_lowercase),
        secrets.choice(string.digits),
    ]
    if include_symbols:
        password.append(secrets.choice("!@#$%^&*"))
    password.extend(secrets.choice(alphabet) for _ in range(length - len(password)))
    secrets.SystemRandom().shuffle(password)
    return "".join(password)


def handler(event):
    """Generate cryptographically secure secrets."""
    try:
        secret_type = event.get("type", "password")
        length = max(8, int(event.get("length", 32)))
        include_symbols = event.get("include_symbols", True)
        count = max(1, min(100, int(event.get("count", 1))))

        generated = []
        for _ in range(count):
            if secret_type == "password":
                s = generate_password(length, include_symbols)
            elif secret_type == "api_key":
                s = secrets.token_urlsafe(length)
            elif secret_type == "jwt_secret":
                s = base64.urlsafe_b64encode(secrets.token_bytes(64)).decode()
            elif secret_type == "hex":
                s = secrets.token_hex(length // 2)
            elif secret_type == "base64":
                s = base64.urlsafe_b64encode(secrets.token_bytes(length)).decode().rstrip("=")
            elif secret_type == "uuid":
                s = str(uuid.uuid4())
            else:
                s = secrets.token_urlsafe(length)
            generated.append(s)

        result = generated[0] if count == 1 else generated
        return {"ok": True, "result": result, "secrets": generated}
    except Exception as e:
        return {"ok": False, "error": str(e)}
