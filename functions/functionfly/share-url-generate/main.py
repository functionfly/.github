import urllib.parse


SHARE_TEMPLATES = {
    "twitter": "https://twitter.com/intent/tweet?url={url}&text={text}&hashtags={hashtags}",
    "facebook": "https://www.facebook.com/sharer/sharer.php?u={url}",
    "linkedin": "https://www.linkedin.com/shareArticle?mini=true&url={url}&title={title}",
    "reddit": "https://www.reddit.com/submit?url={url}&title={title}",
    "whatsapp": "https://api.whatsapp.com/send?text={text}%20{url}",
    "telegram": "https://t.me/share/url?url={url}&text={text}",
    "pinterest": "https://pinterest.com/pin/create/button/?url={url}&description={text}",
    "email": "mailto:?subject={title}&body={text}%0A{url}",
    "hacker-news": "https://news.ycombinator.com/submitlink?u={url}&t={title}",
}


def handler(event):
    url = event.get("url") if isinstance(event, dict) else None
    platform = event.get("platform", "twitter").lower()
    text = event.get("text", "")
    title = event.get("title", "")
    hashtags = event.get("hashtags", "")
    if not url:
        return {"ok": False, "error": "url is required"}
    if platform not in SHARE_TEMPLATES:
        return {"ok": False, "error": f"platform must be one of: {', '.join(SHARE_TEMPLATES.keys())}"}
    try:
        tpl = SHARE_TEMPLATES[platform]
        share_url = tpl.format(
            url=urllib.parse.quote(str(url), safe=''),
            text=urllib.parse.quote(str(text), safe=''),
            title=urllib.parse.quote(str(title), safe=''),
            hashtags=urllib.parse.quote(str(hashtags).strip("#"), safe=''),
        )
        return {"ok": True, "result": share_url, "share_url": share_url, "platform": platform}
    except Exception as e:
        return {"ok": False, "error": str(e)}
