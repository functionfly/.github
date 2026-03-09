import re


def handler(event):
    mac = event.get("mac")
    output_format = event.get("output_format", "colon")  # "colon", "dash", "dot", "none"

    if not mac:
        return {"ok": False, "error": "mac is required"}

    if not isinstance(mac, str):
        return {"ok": False, "error": "mac must be a string"}

    try:
        # Normalize input: remove all separators and convert to lowercase
        normalized = re.sub(r'[^0-9a-fA-F]', '', mac).lower()

        if len(normalized) != 12:
            return {"ok": False, "error": f"MAC address must have exactly 12 hexadecimal digits, got {len(normalized)}"}

        # Validate hexadecimal
        try:
            int(normalized, 16)
        except ValueError:
            return {"ok": False, "error": "MAC address contains invalid hexadecimal characters"}

        # Split into octets
        octets = [normalized[i:i+2] for i in range(0, 12, 2)]

        # Format output based on requested format
        if output_format == "colon":
            formatted = ':'.join(octets)
        elif output_format == "dash":
            formatted = '-'.join(octets)
        elif output_format == "dot":
            # Cisco-style: xxxx.xxxx.xxxx
            formatted = '.'.join([''.join(octets[i:i+2]) for i in range(0, 6, 2)])
        elif output_format == "none":
            formatted = normalized
        else:
            return {"ok": False, "error": f"unsupported output_format: {output_format}"}

        # Check for special MAC addresses
        first_octet = int(octets[0], 16)
        is_multicast = bool(first_octet & 0x01)
        is_locally_administered = bool(first_octet & 0x02)
        is_broadcast = normalized == "ffffffffffff"

        # Extract OUI (first 3 octets)
        oui = ':'.join(octets[:3])

        result = {
            "original_mac": mac,
            "normalized_mac": normalized.upper(),
            "formatted_mac": formatted,
            "output_format": output_format,
            "octets": octets,
            "oui": oui,
            "is_multicast": is_multicast,
            "is_locally_administered": is_locally_administered,
            "is_broadcast": is_broadcast,
            "is_unicast": not is_multicast
        }

        return {
            "ok": True,
            "result": result
        }

    except Exception as e:
        return {"ok": False, "error": f"failed to format MAC address: {str(e)}"}