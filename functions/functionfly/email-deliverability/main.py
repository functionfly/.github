import re
import socket
from typing import Any


EMAIL_REGEX = re.compile(
    r"^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$"
)

DISPOSABLE_DOMAINS = {
    "tempmail.com", "guerrillamail.com", "mailinator.com", "10minutemail.com",
    "throwaway.email", "fakeinbox.com", "trashmail.com", "yopmail.com",
    "sharklasers.com", "guerrillamail.info", "grr.la", "maildrop.cc"
}


def validate_email_syntax(email: str) -> bool:
    if not email or not isinstance(email, str):
        return False
    
    email = email.strip().lower()
    
    if len(email) > 254:
        return False
    
    if not EMAIL_REGEX.match(email):
        return False
    
    local, domain = email.rsplit('@', 1)
    
    if len(local) > 64:
        return False
    
    if '..' in local or local.startswith('.') or local.endswith('.'):
        return False
    
    if domain.startswith('-') or domain.endswith('-'):
        return False
    
    if domain.startswith('.') or domain.endswith('.'):
        return False
    
    if not re.match(r'^[a-zA-Z0-9.-]+$', domain):
        return False
    
    return True


def check_domain_mx(domain: str) -> tuple[bool, list]:
    try:
        if not domain or not isinstance(domain, str):
            return False, []
        
        domain = domain.strip().lower()
        
        if domain.startswith('@'):
            domain = domain[1:]
        
        if not re.match(r'^[a-zA-Z0-9.-]+$', domain):
            return False, []
        
        mx_records = socket.getaddrinfo(domain, 25)
        
        mx_hosts = []
        for r in mx_records:
            if r[0] == socket.AF_INET or r[0] == socket.AF_INET6:
                mx_hosts.append(r[3][0] if isinstance(r[3], tuple) else r[3])
        
        mx_hosts = list(set(mx_hosts))
        
        return len(mx_hosts) > 0, mx_hosts[:3]
        
    except socket.gaierror:
        return False, []
    except Exception:
        return False, []


def get_suggestions(email: str, issues: list) -> list:
    suggestions = []
    
    if "invalid_syntax" in issues:
        if '@' not in email:
            suggestions.append("Add an @ symbol to separate local and domain parts")
        else:
            local, domain = email.rsplit('@', 1)
            if not domain:
                suggestions.append("Add a domain after the @ symbol")
            if not local:
                suggestions.append("Add a username before the @ symbol")
            elif len(local) > 64:
                suggestions.append("Shorten the username part (max 64 characters)")
    
    if "disposable_domain" in issues:
        suggestions.append("Avoid using disposable email services for important communications")
    
    if "domain_no_mx" in issues:
        suggestions.append("Verify the recipient's email domain is correct")
        suggestions.append("The recipient's email server may be temporarily unavailable")
    
    if "missing_recipient" in issues:
        suggestions.append("Enter a valid email address")
    
    return suggestions


def handler(event: dict[str, Any]) -> dict[str, Any]:
    try:
        email_address = event.get("email_address", "")
        
        if not email_address:
            return {"ok": False, "error": "email_address is required"}
        
        if not isinstance(email_address, str):
            return {"ok": False, "error": "email_address must be a string"}
        
        email_address = email_address.strip()
        
        if len(email_address) == 0:
            return {"ok": False, "error": "email_address cannot be empty"}
        
        is_valid_format = validate_email_syntax(email_address)
        
        issues = []
        
        if not is_valid_format:
            issues.append("invalid_syntax")
        
        domain = ""
        if "@" in email_address:
            domain = email_address.rsplit("@", 1)[1].lower()
        
        if domain:
            if domain in DISPOSABLE_DOMAINS:
                issues.append("disposable_domain")
            
            if is_valid_format:
                domain_has_mx, mx_hosts = check_domain_mx(domain)
                
                if not domain_has_mx:
                    issues.append("domain_no_mx")
            else:
                domain_has_mx = False
                mx_hosts = []
        else:
            domain_has_mx = False
            mx_hosts = []
            issues.append("missing_recipient")
        
        suggestions = get_suggestions(email_address, issues)
        
        result = {
            "ok": True,
            "is_valid_format": is_valid_format,
            "domain_has_mx": domain_has_mx,
            "suggestions": suggestions,
            "email_address": email_address
        }
        
        if mx_hosts:
            result["mx_hosts"] = mx_hosts
        
        return result
        
    except Exception as e:
        return {"ok": False, "error": str(e)}
