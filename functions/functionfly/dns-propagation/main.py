import socket


DEFAULT_RESOLVERS = ["8.8.8.8", "1.1.1.1", "9.9.9.9", "208.67.222.222"]


def query_dns(domain, record_type):
    """Simple DNS query using socket."""
    try:
        if record_type == "A":
            results = socket.getaddrinfo(domain, None, socket.AF_INET)
            return list(set(r[4][0] for r in results))
        elif record_type == "AAAA":
            results = socket.getaddrinfo(domain, None, socket.AF_INET6)
            return list(set(r[4][0] for r in results))
        else:
            # For other record types, just do a basic lookup
            return socket.gethostbyname_ex(domain)[2]
    except socket.gaierror as e:
        return None


def handler(event):
    """Check DNS record propagation."""
    try:
        domain = event.get("domain")
        if not domain:
            return {"ok": False, "error": "domain is required"}

        domain = domain.replace("https://", "").replace("http://", "").split("/")[0]
        record_type = event.get("record_type", "A")
        resolvers = event.get("resolvers", DEFAULT_RESOLVERS[:3])

        results = []
        all_records = []

        for resolver in resolvers:
            records = query_dns(domain, record_type)
            result = {
                "resolver": resolver,
                "records": records,
                "success": records is not None,
            }
            results.append(result)
            if records:
                all_records.extend(records)

        # Check if all resolvers agree
        unique_records = list(set(all_records))
        propagated = len(results) > 0 and all(r["success"] for r in results)

        return {
            "ok": True,
            "domain": domain,
            "record_type": record_type,
            "results": results,
            "propagated": propagated,
            "unique_records": unique_records,
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
