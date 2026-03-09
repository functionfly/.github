def handler(event):
    ipv4 = event.get("ipv4")

    if not ipv4:
        return {"ok": False, "error": "ipv4 is required"}

    if not isinstance(ipv4, str):
        return {"ok": False, "error": "ipv4 must be a string"}

    try:
        # Split into octets
        octets = ipv4.split('.')

        if len(octets) != 4:
            return {"ok": False, "error": f"IPv4 address must have exactly 4 octets, got {len(octets)}"}

        # Convert to integers and validate
        octet_values = []
        for i, octet in enumerate(octets):
            try:
                value = int(octet)
                if value < 0 or value > 255:
                    return {"ok": False, "error": f"octet {i+1} ({octet}) must be between 0 and 255"}
                octet_values.append(value)
            except ValueError:
                return {"ok": False, "error": f"octet {i+1} ('{octet}') is not a valid integer"}

        # Calculate integer value: (octet1 * 256^3) + (octet2 * 256^2) + (octet3 * 256) + octet4
        integer_value = (
            (octet_values[0] << 24) +
            (octet_values[1] << 16) +
            (octet_values[2] << 8) +
            octet_values[3]
        )

        result = {
            "ipv4": ipv4,
            "integer_value": integer_value,
            "hex_value": f"0x{integer_value:08x}",
            "binary_value": f"0b{integer_value:032b}",
            "octets": octet_values,
            "big_endian_bytes": [
                (integer_value >> 24) & 0xFF,
                (integer_value >> 16) & 0xFF,
                (integer_value >> 8) & 0xFF,
                integer_value & 0xFF
            ]
        }

        return {
            "ok": True,
            "result": result
        }

    except Exception as e:
        return {"ok": False, "error": f"failed to convert IPv4 to integer: {str(e)}"}