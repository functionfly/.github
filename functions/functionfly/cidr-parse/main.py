import ipaddress


def handler(event):
    cidr = event.get("cidr")

    if not cidr:
        return {"ok": False, "error": "cidr is required"}

    if not isinstance(cidr, str):
        return {"ok": False, "error": "cidr must be a string"}

    try:
        # Determine if it's IPv4 or IPv6 CIDR
        if ':' in cidr:
            # IPv6 CIDR
            network = ipaddress.IPv6Network(cidr, strict=False)
            version = 6
        else:
            # IPv4 CIDR
            network = ipaddress.IPv4Network(cidr, strict=False)
            version = 4

        # Calculate subnet information
        prefix_length = network.prefixlen
        subnet_mask = str(network.netmask)
        network_address = str(network.network_address)
        broadcast_address = str(network.broadcast_address) if version == 4 else None
        first_usable = str(network.network_address + 1) if network.network_address != network.broadcast_address else None
        last_usable = str(network.broadcast_address - 1) if network.network_address != network.broadcast_address else None

        # Calculate number of addresses
        num_addresses = network.num_addresses
        num_hosts = num_addresses - 2 if version == 4 and num_addresses > 2 else num_addresses

        result = {
            "cidr": cidr,
            "version": version,
            "is_ipv4": version == 4,
            "is_ipv6": version == 6,
            "network_address": network_address,
            "prefix_length": prefix_length,
            "subnet_mask": subnet_mask,
            "broadcast_address": broadcast_address,
            "first_usable_address": first_usable,
            "last_usable_address": last_usable,
            "total_addresses": num_addresses,
            "usable_addresses": num_hosts,
            "is_private": network.is_private,
            "is_loopback": network.is_loopback,
            "is_multicast": network.is_multicast,
            "is_link_local": network.is_link_local,
            "is_reserved": network.is_reserved
        }

        # IPv4 specific info
        if version == 4:
            # Calculate subnet class
            first_octet = network.network_address.packed[0]
            if first_octet <= 127:
                subnet_class = "A"
            elif first_octet <= 191:
                subnet_class = "B"
            elif first_octet <= 223:
                subnet_class = "C"
            else:
                subnet_class = "Other"

            result["subnet_class"] = subnet_class

            # Binary representation of subnet mask
            mask_int = int(network.netmask)
            result["subnet_mask_binary"] = f"0b{mask_int:032b}"
            result["subnet_mask_hex"] = f"0x{mask_int:08x}"

        return {
            "ok": True,
            "result": result
        }

    except (ipaddress.AddressValueError, ValueError) as e:
        return {"ok": False, "error": f"invalid CIDR notation: {str(e)}"}
    except Exception as e:
        return {"ok": False, "error": f"failed to parse CIDR: {str(e)}"}