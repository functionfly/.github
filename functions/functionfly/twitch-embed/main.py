def handler(event):
    channel = event.get("channel") if isinstance(event, dict) else None
    video_id = event.get("video_id")
    clip_id = event.get("clip_id")
    parent = event.get("parent", "localhost")
    width = int(event.get("width", 640))
    height = int(event.get("height", 360))
    autoplay = event.get("autoplay", False)
    muted = event.get("muted", True)
    if not channel and not video_id and not clip_id:
        return {"ok": False, "error": "channel, video_id, or clip_id is required"}
    try:
        params = [f"parent={parent}", f"width={width}", f"height={height}"]
        if autoplay: params.append("autoplay=true")
        if muted: params.append("muted=true")
        if channel:
            src = f"https://player.twitch.tv/?channel={channel}&{'&'.join(params)}"
        elif video_id:
            src = f"https://player.twitch.tv/?video={video_id}&{'&'.join(params)}"
        else:
            src = f"https://clips.twitch.tv/embed?clip={clip_id}&{'&'.join(params)}"
        html = f'<iframe src="{src}" width="{width}" height="{height}" allowfullscreen frameborder="0" scrolling="no"></iframe>'
        return {"ok": True, "result": html, "embed_html": html, "embed_url": src}
    except Exception as e:
        return {"ok": False, "error": str(e)}
