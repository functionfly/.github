import ipaddress


def handler(event):
    cidr = event.get("cidr")
    ip = event.get("ip")

    if not cidr:
        return {"ok": False, "error": "cidr is required"}

    if not ip:
        return {"ok": False, "error": "ip is required"}

    if not isinstance(cidr, str):
        return {"ok": False, "error": "cidr must be a string"}

    if not isinstance(ip, str):
        return {"ok": False, "error": "ip must be a string"}

    try:
        # Parse CIDR
        if ':' in cidr:
            network = ipaddress.IPv6Network(cidr, strict=False)
            version = 6
        else:
            network = ipaddress.IPv4Network(cidr, strict=False)
            version = 4

        # Parse IP address
        if version == 4:
            ip_obj = ipaddress.IPv4Address(ip)
        else:
            ip_obj = ipaddress.IPv6Address(ip)

        # Check if IP is in network
        contains = ip_obj in network

        result = {
            "cidr": cidr,
            "ip": ip,
            "contains": contains,
            "version": version,
            "network_address": str(network.network_address),
            "prefix_length": network.prefixlen,
            "subnet_mask": str(network.netmask)
        }

        if contains:
            # Additional info if IP is contained
            if version == 4:
                # Calculate host number within subnet
                host_number = int(ip_obj) - int(network.network_address)
                result["host_number"] = host_number
                result["is_network_address"] = ip_obj == network.network_address
                result["is_broadcast_address"] = ip_obj == network.broadcast_address
            else:
                result["is_network_address"] = ip_obj == network.network_address

        return {
            "ok": True,
            "result": result
        }

    except (ipaddress.AddressValueError, ValueError) as e:
        return {"ok": False, "error": f"invalid CIDR or IP address: {str(e)}"}
    except Exception as e:
        return {"ok": False, "error": f"failed to check CIDR containment: {str(e)}"}