def handler(event):
    password = event.get("password") if isinstance(event, dict) else None
    hash_ = event.get("hash")

    if not password:
        return {"ok": False, "error": "password is required"}
    if not hash_:
        return {"ok": False, "error": "hash is required"}

    try:
        from argon2 import PasswordHasher
        from argon2.exceptions import VerifyMismatchError, VerificationError, InvalidHashError
        ph = PasswordHasher()
        try:
            ph.verify(str(hash_), str(password))
            return {"ok": True, "result": True}
        except VerifyMismatchError:
            return {"ok": True, "result": False}
        except (VerificationError, InvalidHashError) as e:
            return {"ok": False, "error": str(e)}
    except ImportError:
        return {"ok": False, "error": "argon2-cffi library is not installed. Install with: pip install argon2-cffi"}
    except Exception as e:
        return {"ok": False, "error": str(e)}
