def _get_known_tlds():
    """Return a set of known TLDs (simplified list)"""
    return {
        # Generic TLDs
        'com', 'org', 'net', 'info', 'biz', 'name', 'pro',
        # Country code TLDs
        'us', 'uk', 'de', 'fr', 'it', 'es', 'ca', 'au', 'jp', 'cn', 'in', 'br', 'mx', 'ru', 'nl', 'se', 'no', 'fi', 'dk', 'pl', 'cz', 'sk', 'hu', 'ro', 'bg', 'si', 'hr', 'ba', 'me', 'rs', 'mk', 'al', 'gr', 'tr', 'il', 'sa', 'ae', 'qa', 'kw', 'bh', 'om', 'jo', 'lb', 'sy', 'iq', 'ye', 'ps', 'eg', 'ly', 'tn', 'ma', 'dz', 'eh', 'ao', 'bj', 'bw', 'bf', 'bi', 'cm', 'cv', 'cf', 'td', 'cg', 'cd', 'ci', 'dj', 'gq', 'er', 'et', 'ga', 'gm', 'gh', 'gn', 'gw', 'gq', 'ke', 'ls', 'lr', 'ly', 'mg', 'mw', 'ml', 'mr', 'mu', 'mz', 'na', 'ne', 'ng', 're', 'rw', 'sn', 'sc', 'sl', 'so', 'st', 'sz', 'tg', 'tz', 'ug', 'za', 'zm', 'zw'
    }


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

        known_tlds = _get_known_tlds()

        # Try to find TLD (start from the end)
        tld = None
        remaining_parts = []

        # Check for 2-part TLDs first (like co.uk, com.au)
        if len(parts) >= 3:
            potential_tld = f"{parts[-2]}.{parts[-1]}"
            if potential_tld in ['co.uk', 'com.au', 'org.uk', 'net.au', 'com.cn', 'org.cn', 'com.tw', 'org.tw', 'co.jp', 'ne.jp']:
                tld = potential_tld
                remaining_parts = parts[:-2]
            else:
                # Check single TLD
                if parts[-1] in known_tlds:
                    tld = parts[-1]
                    remaining_parts = parts[:-1]

        # If no 2-part TLD found, check single TLD
        if not tld and parts[-1] in known_tlds:
            tld = parts[-1]
            remaining_parts = parts[:-1]

        # Fallback: assume last part is TLD
        if not tld:
            tld = parts[-1]
            remaining_parts = parts[:-1]

        result = {
            "domain": domain,
            "tld": tld,
            "domain_without_tld": '.'.join(remaining_parts) if remaining_parts else None,
            "is_known_tld": tld in known_tlds,
            "tld_parts": tld.split('.') if tld else [],
            "tld_length": len(tld.split('.')) if tld else 0
        }

        return {
            "ok": True,
            "result": result
        }

    except Exception as e:
        return {"ok": False, "error": f"failed to extract TLD: {str(e)}"}