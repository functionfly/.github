def handler(event):
    integer_value = event.get("integer_value")

    if integer_value is None:
        return {"ok": False, "error": "integer_value is required"}

    try:
        # Convert to integer if it's a string
        if isinstance(integer_value, str):
            if integer_value.startswith('0x'):
                integer_value = int(integer_value, 16)
            elif integer_value.startswith('0b'):
                integer_value = int(integer_value, 2)
            else:
                integer_value = int(integer_value)
        elif not isinstance(integer_value, int):
            return {"ok": False, "error": "integer_value must be an integer or string representation"}

        # Check range for 32-bit unsigned integer
        if integer_value < 0 or integer_value > 4294967295:
            return {"ok": False, "error": f"integer_value {integer_value} must be between 0 and 4294967295"}

        # Extract octets from integer
        octet1 = (integer_value >> 24) & 0xFF
        octet2 = (integer_value >> 16) & 0xFF
        octet3 = (integer_value >> 8) & 0xFF
        octet4 = integer_value & 0xFF

        ipv4 = f"{octet1}.{octet2}.{octet3}.{octet4}"

        result = {
            "integer_value": integer_value,
            "ipv4": ipv4,
            "hex_value": f"0x{integer_value:08x}",
            "binary_value": f"0b{integer_value:032b}",
            "octets": [octet1, octet2, octet3, octet4],
            "big_endian_bytes": [octet1, octet2, octet3, octet4]
        }

        return {
            "ok": True,
            "result": result
        }

    except ValueError as e:
        return {"ok": False, "error": f"invalid integer format: {str(e)}"}
    except Exception as e:
        return {"ok": False, "error": f"failed to convert integer to IPv4: {str(e)}"}