FB_POST_LIMIT = 63206
FB_COMMENT_LIMIT = 8000
FB_STORY_LIMIT = 500


def handler(event):
    text = event.get("text") if isinstance(event, dict) else None
    content_type = event.get("content_type", "post")
    if text is None:
        return {"ok": False, "error": "text is required"}
    try:
        t = str(text)
        char_count = len(t)
        limits = {"post": FB_POST_LIMIT, "comment": FB_COMMENT_LIMIT, "story": FB_STORY_LIMIT}
        limit = limits.get(content_type, FB_POST_LIMIT)
        remaining = limit - char_count
        over_limit = char_count > limit
        return {
            "ok": True,
            "result": char_count,
            "char_count": char_count,
            "limit": limit,
            "content_type": content_type,
            "remaining": remaining,
            "over_limit": over_limit,
            "fits": not over_limit
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
