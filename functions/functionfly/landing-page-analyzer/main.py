"""Landing Page Analyzer - Analyze landing page for SEO and performance."""
import re
from urllib.request import urlopen
from urllib.error import URLError, HTTPError
import ssl


def extract_title_and_meta(html_content):
    title_match = re.search(r'<title[^>]*>([^<]+)</title>', html_content, re.IGNORECASE)
    title = title_match.group(1).strip() if title_match else ""

    desc_match = re.search(r'<meta[^>]*name=["\']description["\'][^>]*content=["\']([^"\']+)["\']', html_content, re.IGNORECASE)
    if not desc_match:
        desc_match = re.search(r'<meta[^>]*content=["\']([^"\']+)["\'][^>]*name=["\']description["\']', html_content, re.IGNORECASE)
    meta_description = desc_match.group(1).strip() if desc_match else ""

    return title, meta_description


def extract_headings(html_content):
    h1_matches = re.findall(r'<h1[^>]*>([^<]+)</h1>', html_content, re.IGNORECASE)
    h2_matches = re.findall(r'<h2[^>]*>([^<]+)</h2>', html_content, re.IGNORECASE)
    h3_matches = re.findall(r'<h3[^>]*>([^<]+)</h3>', html_content, re.IGNORECASE)

    headings = []
    for h1 in h1_matches:
        headings.append({"level": "h1", "text": h1.strip()})
    for h2 in h2_matches[:5]:
        headings.append({"level": "h2", "text": h2.strip()})
    for h3 in h3_matches[:5]:
        headings.append({"level": "h3", "text": h3.strip()})

    return headings


def calculate_keyword_density(html_content, keyword):
    clean_html = re.sub(r'<[^>]+>', ' ', html_content)
    clean_text = re.sub(r'[^\w\s]', ' ', clean_html.lower())
    words = clean_text.split()
    words = [w for w in words if len(w) > 2]

    if not words:
        return 0

    keyword_lower = keyword.lower()
    keyword_words = keyword_lower.split()

    if len(keyword_words) == 1:
        count = sum(1 for w in words if keyword_lower in w)
    else:
        count = clean_text.count(keyword_lower)

    density = count / len(words)
    return round(density, 4)


def check_cta_presence(html_content):
    cta_patterns = [
        r'<button[^>]*>',
        r'<a[^>]*class=["\'][^"\']*(?:cta|button|signup|submit|register|get started|sign up)[^"\']*["\']',
        r'<input[^>]*type=["\']submit["\']',
        r'call\s*to\s*action',
        r'sign\s*up',
        r'get\s*started',
        r'subscribe',
        r'register',
        r'buy\s*now',
        r'shop\s*now'
    ]

    for pattern in cta_patterns:
        if re.search(pattern, html_content, re.IGNORECASE):
            return True

    return False


def estimate_mobile_friendly(html_content):
    viewport_match = re.search(r'<meta[^>]*viewport[^>]*>', html_content, re.IGNORECASE)
    media_queries = re.findall(r'@media', html_content, re.IGNORECASE)
    mobile_links = re.findall(r'<link[^>]*media=["\'][^"\']*(?:mobile|tablet|small)[^"\']*["\']', html_content, re.IGNORECASE)

    score = 0
    if viewport_match:
        score += 40
    if media_queries:
        score += 30
    if mobile_links:
        score += 30

    return score >= 50


def estimate_load_speed(html_content):
    size_bytes = len(html_content.encode('utf-8'))
    size_kb = size_bytes / 1024

    scripts = re.findall(r'<script[^>]*src=["\'][^"\']+["\']', html_content, re.IGNORECASE)
    images = re.findall(r'<img[^>]*>', html_content, re.IGNORECASE)
    stylesheets = re.findall(r'<link[^>]*rel=["\']stylesheet["\'][^>]*>', html_content, re.IGNORECASE)

    resource_count = len(scripts) + len(images) + len(stylesheets)

    if size_kb < 100 and resource_count < 10:
        return "fast"
    elif size_kb < 500 and resource_count < 30:
        return "medium"
    else:
        return "slow"


def generate_recommendations(title, meta_description, headings, keyword_density, has_cta, mobile_friendly, load_speed, word_count):
    recommendations = []

    if not title:
        recommendations.append("Add a unique, keyword-rich title tag")
    elif len(title) > 60:
        recommendations.append("Shorten title tag to 60 characters or less")
    else:
        recommendations.append("Title tag is well-optimized")

    if not meta_description:
        recommendations.append("Add a meta description with target keywords")
    elif len(meta_description) > 160:
        recommendations.append("Shorten meta description to 160 characters or less")
    else:
        recommendations.append("Meta description is well-optimized")

    if not headings:
        recommendations.append("Add heading tags (H1, H2, H3) to structure content")
    else:
        recommendations.append("Heading structure is present")

    if keyword_density < 0.01:
        recommendations.append("Increase keyword density (current: " + str(keyword_density) + ")")
    elif keyword_density > 0.05:
        recommendations.append("Keyword density is high - consider reducing to avoid over-optimization")
    else:
        recommendations.append("Keyword density is in optimal range")

    if not has_cta:
        recommendations.append("Add clear call-to-action buttons")
    else:
        recommendations.append("Call-to-action is present")

    if not mobile_friendly:
        recommendations.append("Improve mobile friendliness with responsive design")
    else:
        recommendations.append("Mobile-friendly design detected")

    if load_speed == "slow":
        recommendations.append("Improve page load speed - optimize images and reduce scripts")
    elif load_speed == "medium":
        recommendations.append("Consider further load speed optimizations")
    else:
        recommendations.append("Page load speed is good")

    if word_count < 300:
        recommendations.append("Consider adding more content (current: " + str(word_count) + " words)")
    elif word_count < 600:
        recommendations.append("Content length is moderate - consider expanding for better SEO")
    else:
        recommendations.append("Content length is good for SEO")

    return recommendations


def handler(event):
    try:
        url = event.get("url", "")
        target_keyword = event.get("target_keyword", "")

        if not url:
            return {"ok": False, "error": "url is required"}

        context = ssl.create_default_context()
        context.check_hostname = False
        context.verify_mode = ssl.CERT_NONE

        try:
            with urlopen(url, timeout=10, context=context) as response:
                html_content = response.read().decode('utf-8', errors='ignore')
        except HTTPError as e:
            return {"ok": False, "error": f"HTTP error: {e.code}"}
        except URLError as e:
            return {"ok": False, "error": f"URL error: {e.reason}"}
        except Exception as e:
            return {"ok": False, "error": f"Failed to fetch page: {str(e)}"}

        title, meta_description = extract_title_and_meta(html_content)
        headings = extract_headings(html_content)

        clean_html = re.sub(r'<[^>]+>', ' ', html_content)
        word_count = len([w for w in clean_html.split() if len(w) > 2])

        keyword_density = calculate_keyword_density(html_content, target_keyword)
        has_cta = check_cta_presence(html_content)
        mobile_friendly = estimate_mobile_friendly(html_content)
        load_speed = estimate_load_speed(html_content)

        recommendations = generate_recommendations(
            title, meta_description, headings, keyword_density,
            has_cta, mobile_friendly, load_speed, word_count
        )

        return {
            "ok": True,
            "title": title,
            "meta_description": meta_description,
            "headings": headings,
            "word_count": word_count,
            "keyword_density": keyword_density,
            "has_cta": has_cta,
            "mobile_friendly": mobile_friendly,
            "load_speed_estimate": load_speed,
            "recommendations": recommendations,
            "url": url,
            "target_keyword": target_keyword
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
