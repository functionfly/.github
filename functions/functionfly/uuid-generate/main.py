import uuid


def handler(event):
    """
    Generate UUID v4 (random) identifiers.

    Input:
        - count: Number of UUIDs to generate (default: 1, max typically 100)

    Returns:
        - ok: True on success
        - uuid: Single UUID string (if count=1)
        - uuids: List of UUID strings (if count>1)
        - error: Message if ok is False
    """
    if isinstance(event, dict):
        count = event.get("count", 1)
    else:
        count = 1

    try:
        count = int(count)
        if count < 1:
            count = 1
        if count > 100:
            count = 100
    except (TypeError, ValueError):
        count = 1

    try:
        if count == 1:
            return {"ok": True, "uuid": str(uuid.uuid4())}
        return {"ok": True, "uuids": [str(uuid.uuid4()) for _ in range(count)]}
    except Exception as e:
        return {"ok": False, "error": str(e)}
