from urllib.parse import urlparse
import re


def _extract_domain_from_url(url):
    """Extract domain from a full URL"""
    try:
        parsed = urlparse(url)
        return parsed.hostname or parsed.netloc
    except:
        return None


def _clean_domain(domain):
    """Clean and normalize domain name"""
    if not domain:
        return None

    # Remove port if present
    domain = domain.split(':')[0]

    # Convert to lowercase
    domain = domain.lower()

    # Remove www. prefix if present
    if domain.startswith('www.'):
        domain = domain[4:]

    return domain


def handler(event):
    input_text = event.get("url") or event.get("domain") or event.get("text")

    if not input_text:
        return {"ok": False, "error": "url, domain, or text is required"}

    if not isinstance(input_text, str):
        return {"ok": False, "error": "input must be a string"}

    try:
        # Try to extract from URL first
        domain = _extract_domain_from_url(input_text)

        if not domain:
            # Try to find domain-like patterns in text
            # Simple regex for domain-like strings
            domain_pattern = r'\b([a-zA-Z0-9-]+\.)*[a-zA-Z0-9-]+\.[a-zA-Z]{2,}\b'
            matches = re.findall(domain_pattern, input_text)

            if matches:
                # Take the first match and clean it
                domain = _clean_domain(matches[0])

        if not domain:
            return {"ok": False, "error": "no valid domain found in input"}

        # Further cleaning and validation
        cleaned_domain = _clean_domain(domain)

        if not cleaned_domain:
            return {"ok": False, "error": "extracted domain is invalid"}

        # Basic domain validation
        if not re.match(r'^[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$', cleaned_domain):
            return {"ok": False, "error": f"extracted domain '{cleaned_domain}' does not appear to be valid"}

        result = {
            "domain": cleaned_domain,
            "original_input": input_text,
            "is_from_url": bool(_extract_domain_from_url(input_text)),
            "has_www": cleaned_domain.startswith('www.')
        }

        # Split domain parts
        parts = cleaned_domain.split('.')
        if len(parts) >= 2:
            result["tld"] = parts[-1]
            result["domain_name"] = '.'.join(parts[:-1])
            if len(parts) >= 3:
                result["subdomain"] = '.'.join(parts[:-2])

        return {
            "ok": True,
            "result": result
        }

    except Exception as e:
        return {"ok": False, "error": f"failed to extract domain: {str(e)}"}