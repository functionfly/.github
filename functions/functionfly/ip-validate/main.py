import ipaddress


def handler(event):
    ip = event.get("ip")

    if not ip:
        return {"ok": False, "error": "ip is required"}

    if not isinstance(ip, str):
        return {"ok": False, "error": "ip must be a string"}

    try:
        # Try to parse as IPv4 first
        try:
            ip_obj = ipaddress.IPv4Address(ip)
            version = 4
            is_private = ip_obj.is_private
            is_loopback = ip_obj.is_loopback
            is_multicast = ip_obj.is_multicast
            is_link_local = ip_obj.is_link_local
            is_reserved = ip_obj.is_reserved
            packed = ip_obj.packed
            integer_value = int(ip_obj)

        except ipaddress.AddressValueError:
            # Try as IPv6
            try:
                ip_obj = ipaddress.IPv6Address(ip)
                version = 6
                is_private = ip_obj.is_private
                is_loopback = ip_obj.is_loopback
                is_multicast = ip_obj.is_multicast
                is_link_local = ip_obj.is_link_local
                is_reserved = ip_obj.is_reserved
                packed = ip_obj.packed
                integer_value = int(ip_obj)

            except ipaddress.AddressValueError:
                return {
                    "ok": False,
                    "error": f"'{ip}' is not a valid IPv4 or IPv6 address",
                    "result": {
                        "ip": ip,
                        "is_valid": False,
                        "version": None
                    }
                }

        result = {
            "ip": ip,
            "is_valid": True,
            "version": version,
            "is_ipv4": version == 4,
            "is_ipv6": version == 6,
            "is_private": is_private,
            "is_loopback": is_loopback,
            "is_multicast": is_multicast,
            "is_link_local": is_link_local,
            "is_reserved": is_reserved,
            "integer_value": integer_value,
            "hex_value": hex(integer_value),
            "binary_value": bin(integer_value),
            "packed_bytes": packed.hex()
        }

        # IPv4 specific info
        if version == 4:
            octets = ip.split('.')
            result["octets"] = [int(octet) for octet in octets]
            result["class"] = _get_ipv4_class(ip_obj)

        # IPv6 specific info
        elif version == 6:
            result["compressed"] = ip_obj.compressed
            result["exploded"] = ip_obj.exploded
            result["is_6to4"] = ip_obj.is_6to4
            result["is_site_local"] = hasattr(ip_obj, 'is_site_local') and ip_obj.is_site_local

        return {
            "ok": True,
            "result": result
        }

    except Exception as e:
        return {"ok": False, "error": f"failed to validate IP: {str(e)}"}


def _get_ipv4_class(ip_obj):
    """Determine IPv4 class"""
    first_octet = ip_obj.packed[0]
    if first_octet <= 127:
        return "A"
    elif first_octet <= 191:
        return "B"
    elif first_octet <= 223:
        return "C"
    elif first_octet <= 239:
        return "D"
    else:
        return "E"