def handler(event):
    domain = event.get("domain")

    if not domain:
        return {"ok": False, "error": "domain is required"}

    if not isinstance(domain, str):
        return {"ok": False, "error": "domain must be a string"}

    try:
        # Clean the domain
        domain = domain.lower().strip()

        # Remove protocol if present
        if '://' in domain:
            domain = domain.split('://', 1)[1]

        # Remove path, query, etc.
        domain = domain.split('/')[0].split('?')[0].split('#')[0]

        # Remove port if present
        domain = domain.split(':')[0]

        # Split into parts
        parts = domain.split('.')

        if len(parts) < 2:
            return {
                "ok": False,
                "error": f"domain '{domain}' does not appear to be a valid domain"
            }

        # For most domains, the last part is TLD, second to last is main domain
        # Anything before that is subdomain
        tld = parts[-1]
        main_domain = parts[-2]
        subdomain_parts = parts[:-2]

        result = {
            "domain": domain,
            "subdomain": '.'.join(subdomain_parts) if subdomain_parts else None,
            "main_domain": main_domain,
            "tld": tld,
            "full_domain": f"{main_domain}.{tld}",
            "subdomain_parts": subdomain_parts,
            "has_subdomain": len(subdomain_parts) > 0,
            "subdomain_count": len(subdomain_parts)
        }

        # Special handling for common TLDs that might have multiple parts
        common_multi_part_tlds = ['co.uk', 'com.au', 'org.uk', 'net.au', 'com.cn', 'org.cn']

        for multi_tld in common_multi_part_tlds:
            if domain.endswith('.' + multi_tld):
                tld = multi_tld
                main_domain = parts[-3] if len(parts) >= 3 else parts[-2]
                subdomain_parts = parts[:-3] if len(parts) >= 3 else parts[:-2]

                result.update({
                    "subdomain": '.'.join(subdomain_parts) if subdomain_parts else None,
                    "main_domain": main_domain,
                    "tld": tld,
                    "full_domain": f"{main_domain}.{tld}",
                    "subdomain_parts": subdomain_parts,
                    "has_subdomain": len(subdomain_parts) > 0,
                    "subdomain_count": len(subdomain_parts)
                })
                break

        return {
            "ok": True,
            "result": result
        }

    except Exception as e:
        return {"ok": False, "error": f"failed to extract subdomain: {str(e)}"}