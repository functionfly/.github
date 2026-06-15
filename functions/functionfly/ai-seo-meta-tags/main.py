"""AI SEO Meta Tags Generator - Generate SEO meta tags and schema markup."""
import re
import json
import random


def truncate(text, max_length):
    if len(text) <= max_length:
        return text
    return text[:max_length - 3].rsplit(' ', 1)[0] + "..."


def generate_schema(page_title, focus_keyword, page_url):
    return {
        "@context": "https://schema.org",
        "@type": "WebPage",
        "name": page_title,
        "description": f"Learn more about {focus_keyword} at our website.",
        "url": page_url,
        "mainEntity": {
            "@type": "WebSite",
            "name": page_title.split()[0] if page_title.split() else "Website",
            "url": page_url
        }
    }


def handler(event):
    try:
        page_title = event.get("page_title", "")
        page_content = event.get("page_content", "")
        focus_keyword = event.get("focus_keyword", "")
        target_audience = event.get("target_audience", "general audience")

        if not page_title:
            return {"ok": False, "error": "page_title is required"}

        if not focus_keyword:
            words = re.sub(r'[^\w\s]', ' ', page_title.lower()).split()
            focus_keyword = " ".join(words[:3])

        page_url = event.get("page_url", "https://example.com/page")
        canonical_url = page_url.rstrip('/')

        meta_title = page_title
        if len(meta_title) > 60:
            meta_title = truncate(page_title, 60)

        if page_content:
            sentences = page_content.split('.')
            description_text = sentences[0] if sentences else page_content[:160]
        else:
            description_text = f"Comprehensive information about {focus_keyword} for {target_audience}."

        meta_description = description_text
        if len(meta_description) > 160:
            meta_description = truncate(description_text, 160)

        og_title = meta_title
        if len(og_title) > 60:
            og_title = truncate(meta_title, 60)

        og_description = meta_description
        if len(og_description) > 160:
            og_description = truncate(meta_description, 160)

        schema_markup = generate_schema(page_title, focus_keyword, canonical_url)

        return {
            "ok": True,
            "meta_title": meta_title,
            "meta_description": meta_description,
            "og_title": og_title,
            "og_description": og_description,
            "canonical_url": canonical_url,
            "schema_markup": json.dumps(schema_markup, indent=2),
            "focus_keyword": focus_keyword
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
