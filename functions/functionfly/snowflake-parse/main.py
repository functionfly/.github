def parse_snowflake(snowflake_id):
    """Parse a Snowflake ID"""
    # Snowflake ID structure:
    # 1 bit: unused (always 0)
    # 41 bits: timestamp in milliseconds (custom epoch)
    # 5 bits: datacenter ID
    # 5 bits: machine ID
    # 12 bits: sequence number
    EPOCH = 1609459200000  # 2021-01-01 00:00:00 UTC
    snowflake_id = int(snowflake_id)
    timestamp = ((snowflake_id >> 22) & 0x1FFFFFFFFFF) + EPOCH
    datacenter_id = (snowflake_id >> 17) & 0x1F
    machine_id = (snowflake_id >> 12) & 0x1F
    sequence = snowflake_id & 0xFFF
    return timestamp, datacenter_id, machine_id, sequence

def handler(event):
    try:
        snowflake_id = event.get("id", "") if isinstance(event, dict) else ""
        if not snowflake_id:
            return {"ok": False, "error": "id is required"}
        timestamp, datacenter_id, machine_id, sequence = parse_snowflake(snowflake_id)
        return {"ok": True, "timestamp": timestamp, "datacenter_id": datacenter_id, "machine_id": machine_id, "sequence": sequence}
    except Exception as e:
        return {"ok": False, "error": str(e)}
