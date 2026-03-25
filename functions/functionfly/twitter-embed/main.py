import re


def handler(event):
    url = event.get("url") if isinstance(event, dict) else None
    tweet_id = event.get("tweet_id")
    theme = event.get("theme", "light")
    width = int(event.get("width", 550))
    if not url and not tweet_id:
        return {"ok": False, "error": "url or tweet_id is required"}
    try:
        if not tweet_id and url:
            m = re.search(r'(?:twitter|x)\.com/\w+/status/(\d+)', str(url))
            tweet_id = m.group(1) if m else None
        if tweet_id:
            blockquote = f'<blockquote class="twitter-tweet" data-theme="{theme}" data-width="{width}"><a href="https://twitter.com/i/web/status/{tweet_id}"></a></blockquote><script async src="https://platform.twitter.com/widgets.js"></script>'
            return {"ok": True, "result": blockquote, "embed_html": blockquote, "tweet_id": tweet_id}
        return {"ok": True, "result": None, "embed_html": None, "note": "Could not extract tweet ID"}
    except Exception as e:
        return {"ok": False, "error": str(e)}
