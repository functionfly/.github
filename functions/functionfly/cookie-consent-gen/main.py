from typing import Any


COOKIE_CATEGORIES_DEFAULT = ["essential", "analytics", "advertising", "functional"]


ESSENTIAL_COOKIES = [
    {"name": "session_id", "purpose": "Required for website functionality", "duration": "Session"},
    {"name": "csrf_token", "purpose": "Security protection against cross-site attacks", "duration": "Session"},
    {"name": "auth_token", "purpose": "User authentication", "duration": "Session"},
]

CATEGORY_DESCRIPTIONS = {
    "essential": "Required for the website to function properly. Cannot be disabled.",
    "analytics": "Help us understand how visitors interact with our website.",
    "advertising": "Used to deliver personalized advertisements.",
    "functional": "Enable enhanced functionality and personalization.",
}


def generate_cookie_list(cookie_categories: list) -> list[dict]:
    cookies = []
    cookies.extend(ESSENTIAL_COOKIES)
    
    category_cookies = {
        "analytics": [
            {"name": "_ga", "purpose": "Google Analytics - distinguish users", "duration": "2 years"},
            {"name": "_gid", "purpose": "Google Analytics - distinguish users", "duration": "24 hours"},
            {"name": "_gat", "purpose": "Google Analytics - throttle request rate", "duration": "1 minute"},
        ],
        "advertising": [
            {"name": "_gcl_au", "purpose": "Google Ads conversion tracking", "duration": "90 days"},
            {"name": "fr", "purpose": "Facebook Pixel tracking", "duration": "90 days"},
        ],
        "functional": [
            {"name": "language", "purpose": "Store user language preference", "duration": "1 year"},
            {"name": "theme", "purpose": "Store user theme preference", "duration": "1 year"},
        ],
    }
    
    for category in cookie_categories:
        if category in category_cookies:
            cookies.extend(category_cookies[category])
    
    return cookies


def generate_consent_options(cookie_categories: list) -> list[dict]:
    options = []
    
    for category in cookie_categories:
        description = CATEGORY_DESCRIPTIONS.get(category, f"Manage {category} cookies")
        options.append({
            "category": category,
            "description": description,
            "default_enabled": category == "essential"
        })
    
    return options


def generate_banner_text(company_name: str, privacy_policy_url: str) -> str:
    banner_lines = [
        f"We value your privacy at {company_name}.",
        "We use cookies to enhance your browsing experience, serve personalized content, and analyze our traffic.",
        "By clicking 'Accept All', you consent to our use of cookies.",
    ]
    
    banner_text = " ".join(banner_lines)
    
    banner_text += f" Read our <a href='{privacy_policy_url}' target='_blank'>Privacy Policy</a> for more information."
    
    banner_text += " You can customize your preferences by clicking 'Customize' below."
    
    return banner_text


def generate_banner_html(company_name: str, privacy_policy_url: str) -> str:
    html = f'''<div id="cookie-consent-banner" style="position:fixed;bottom:0;left:0;right:0;background:#f8f9fa;padding:20px;border-top:1px solid #dee2e6;z-index:9999;font-family:Arial,sans-serif;">
  <div style="max-width:1200px;margin:0 auto;">
    <p style="margin:0 0 15px;font-size:14px;color:#333;">
      <strong>{company_name}</strong> uses cookies to enhance your experience. 
      By continuing to browse, you agree to our <a href="{privacy_policy_url}" target="_blank" style="color:#0066cc;">Privacy Policy</a>.
    </p>
    <div style="display:flex;gap:10px;flex-wrap:wrap;">
      <button id="cookie-accept-all" onclick="CookieConsent.acceptAll()" style="background:#0066cc;color:#fff;padding:10px 20px;border:none;border-radius:4px;cursor:pointer;font-size:14px;">Accept All</button>
      <button id="cookie-reject-all" onclick="CookieConsent.rejectAll()" style="background:#6c757d;color:#fff;padding:10px 20px;border:none;border-radius:4px;cursor:pointer;font-size:14px;">Reject All</button>
      <button id="cookie-customize" onclick="CookieConsent.showPreferences()" style="background:#fff;color:#0066cc;padding:10px 20px;border:1px solid #0066cc;border-radius:4px;cursor:pointer;font-size:14px;">Customize</button>
    </div>
  </div>
</div>'''
    
    html += '''
<script>
window.CookieConsent = {
  acceptAll: function() {
    this.setConsent({'essential': true, 'analytics': true, 'advertising': true, 'functional': true});
    document.getElementById('cookie-consent-banner').style.display = 'none';
  },
  rejectAll: function() {
    this.setConsent({'essential': true, 'analytics': false, 'advertising': false, 'functional': false});
    document.getElementById('cookie-consent-banner').style.display = 'none';
  },
  showPreferences: function() {
    console.log('Show cookie preferences panel');
  },
  setConsent: function(preferences) {
    document.cookie = 'cookie_consent=' + JSON.stringify(preferences) + ';path=/;max-age=31536000';
  }
};
</script>'''
    
    return html


def handler(event: dict[str, Any]) -> dict[str, Any]:
    try:
        cookie_categories = event.get("cookie_categories", COOKIE_CATEGORIES_DEFAULT)
        privacy_policy_url = event.get("privacy_policy_url", "/privacy-policy")
        company_name = event.get("company_name", "Our Company")
        
        if not isinstance(cookie_categories, list):
            return {"ok": False, "error": "cookie_categories must be a list"}
        
        if len(cookie_categories) == 0:
            return {"ok": False, "error": "cookie_categories cannot be empty"}
        
        if not privacy_policy_url:
            return {"ok": False, "error": "privacy_policy_url is required"}
        
        if not company_name:
            return {"ok": False, "error": "company_name is required"}
        
        valid_categories = ["essential", "analytics", "advertising", "functional", "social_media", "security"]
        for cat in cookie_categories:
            if cat not in valid_categories:
                return {"ok": False, "error": f"Invalid cookie category: {cat}. Valid categories: {', '.join(valid_categories)}"}
        
        banner_text = generate_banner_text(company_name, privacy_policy_url)
        banner_html = generate_banner_html(company_name, privacy_policy_url)
        consent_options = generate_consent_options(cookie_categories)
        cookie_list = generate_cookie_list(cookie_categories)
        
        return {
            "ok": True,
            "banner_text": banner_text,
            "banner_html": banner_html,
            "consent_options": consent_options,
            "cookie_list": cookie_list,
            "company_name": company_name
        }
        
    except Exception as e:
        return {"ok": False, "error": str(e)}
