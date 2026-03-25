LI_POST_LIMIT = 3000
LI_ARTICLE_LIMIT = 125000
LI_COMMENT_LIMIT = 1250
LI_HEADLINE_LIMIT = 220
LI_ABOUT_LIMIT = 2600


def handler(event):
    text = event.get("text") if isinstance(event, dict) else None
    content_type = event.get("content_type", "post")
    if text is None:
        return {"ok": False, "error": "text is required"}
    try:
        t = str(text)
        char_count = len(t)
        limits = {
            "post": LI_POST_LIMIT,
            "article": LI_ARTICLE_LIMIT,
            "comment": LI_COMMENT_LIMIT,
            "headline": LI_HEADLINE_LIMIT,
            "about": LI_ABOUT_LIMIT
        }
        limit = limits.get(content_type, LI_POST_LIMIT)
        remaining = limit - char_count
        over_limit = char_count > limit
        # LinkedIn shows "see more" at ~210 chars for posts
        truncated_at = 210 if content_type == "post" else None
        return {
            "ok": True,
            "result": char_count,
            "char_count": char_count,
            "limit": limit,
            "content_type": content_type,
            "remaining": remaining,
            "over_limit": over_limit,
            "fits": not over_limit,
            "truncated_preview_at": truncated_at
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
