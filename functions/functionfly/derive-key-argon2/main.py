import base64


def handler(event):
    password = event.get("password") if isinstance(event, dict) else None
    salt = event.get("salt")
    time_cost = event.get("time_cost", 3)
    memory_cost = event.get("memory_cost", 65536)
    parallelism = event.get("parallelism", 4)
    length = event.get("length", 32)
    output = event.get("output", "hex")

    if not password:
        return {"ok": False, "error": "password is required"}
    if not salt:
        return {"ok": False, "error": "salt is required"}

    try:
        from argon2.low_level import hash_secret_raw, Type
        try:
            salt_bytes = bytes.fromhex(str(salt))
        except ValueError:
            salt_bytes = base64.b64decode(str(salt))
        dk = hash_secret_raw(
            str(password).encode("utf-8"), salt_bytes,
            time_cost=int(time_cost), memory_cost=int(memory_cost),
            parallelism=int(parallelism), hash_len=int(length),
            type=Type.ID
        )
        result = base64.b64encode(dk).decode("utf-8") if output == "base64" else dk.hex()
        return {"ok": True, "result": result, "time_cost": int(time_cost), "memory_cost": int(memory_cost), "length": int(length)}
    except ImportError:
        return {"ok": False, "error": "argon2-cffi library is not installed. Install with: pip install argon2-cffi"}
    except Exception as e:
        return {"ok": False, "error": str(e)}
