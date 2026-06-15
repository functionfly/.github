import uuid
from typing import Any


def generate_uuidv4() -> str:
    return str(uuid.uuid4())


def generate_uuidv7() -> str:
    import time
    
    nanoseconds = int(time.time_ns())
    uuid_time = (nanoseconds // 1000) + 0x01B21DD213814000
    
    time_low = uuid_time & 0xFFFFFFFF
    time_mid = (uuid_time >> 32) & 0xFFFF
    time_hi_and_version = ((uuid_time >> 48) & 0x0FFF) | (7 << 12)
    
    random_bits = uuid.uuid4().int >> 80
    
    clock_seq = random_bits & 0x3FFF
    clock_seq_low = clock_seq & 0xFF
    clock_seq_hi_and_reserved = (clock_seq >> 8) | 0x80
    
    node = uuid.uuid4().int >> 48
    
    uuid_int = (
        (time_low << 96) |
        ((time_mid << 80) & 0xFFFF000000000000) |
        ((time_hi_and_version << 64) & 0xFFFFFFFF00000000) |
        ((clock_seq_hi_and_reserved << 56) & 0xFF000000000000) |
        ((clock_seq_low << 48) & 0xFF0000000000) |
        (node & 0xFFFFFFFFFFFF)
    )
    
    return str(uuid.UUID(int=uuid_int))


def handler(event: dict[str, Any]) -> dict[str, Any]:
    try:
        num = event.get("num", 1)
        version = event.get("version", 4)
        
        if not isinstance(num, int):
            try:
                num = int(num)
            except (ValueError, TypeError):
                return {"ok": False, "error": "num must be an integer"}
        
        if not isinstance(version, int):
            try:
                version = int(version)
            except (ValueError, TypeError):
                return {"ok": False, "error": "version must be an integer"}
        
        if num < 1:
            return {"ok": False, "error": "num must be at least 1"}
        
        if num > 1000:
            return {"ok": False, "error": "num cannot exceed 1000 (to prevent abuse)"}
        
        if version not in [4, 7]:
            return {"ok": False, "error": "version must be 4 or 7"}
        
        if version == 4:
            uuids = [generate_uuidv4() for _ in range(num)]
        else:
            uuids = [generate_uuidv7() for _ in range(num)]
        
        return {
            "ok": True,
            "uuids": uuids,
            "version": version,
            "count": num
        }
        
    except Exception as e:
        return {"ok": False, "error": str(e)}
