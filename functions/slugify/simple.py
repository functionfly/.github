def handler(event):
    text = event.get("text", "")
    return {"slug": text.lower().replace(" ", "-")}
