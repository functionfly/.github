import re

TWITTER_LIMIT = 280
URL_LENGTH = 23  # Twitter wraps all URLs to t.co (23 chars)


def _count_twitter_chars(text):
    url_pattern = re.compile(r'https?://\S+')
    urls = url_pattern.findall(text)
    text_no_urls = url_pattern.sub('', text)
    # Twitter counts most Unicode chars as 2 but CJK etc. are 2 units
    count = 0
    for ch in text_no_urls:
        cp = ord(ch)
        if 0x1100 <= cp <= 0x115F or 0x2E80 <= cp <= 0x303F or 0x3040 <= cp <= 0x33FF or \
           0x3400 <= cp <= 0x4DBF or 0x4E00 <= cp <= 0xA4FF or 0xA960 <= cp <= 0xA97F or \
           0xAC00 <= cp <= 0xD7FF or 0xF900 <= cp <= 0xFAFF or 0xFE10 <= cp <= 0xFE1F or \
           0xFE30 <= cp <= 0xFE4F or 0xFF00 <= cp <= 0xFFEF or 0x1B000 <= cp <= 0x1B0FF or \
           0x1F004 <= cp <= 0x1F0CF or 0x1F200 <= cp <= 0x1F2FF or 0x20000 <= cp <= 0x2A6DF or \
           0x2A700 <= cp <= 0x2CEAF or 0x2CEB0 <= cp <= 0x2EBEF or 0x30000 <= cp <= 0x3134F:
            count += 2
        else:
            count += 1
    count += len(urls) * URL_LENGTH
    return count


def handler(event):
    text = event.get("text") if isinstance(event, dict) else None
    if text is None:
        return {"ok": False, "error": "text is required"}
    try:
        t = str(text)
        char_count = _count_twitter_chars(t)
        remaining = TWITTER_LIMIT - char_count
        over_limit = char_count > TWITTER_LIMIT
        return {
            "ok": True,
            "result": char_count,
            "char_count": char_count,
            "limit": TWITTER_LIMIT,
            "remaining": remaining,
            "over_limit": over_limit,
            "fits": not over_limit
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
