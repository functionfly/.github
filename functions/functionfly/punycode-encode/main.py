def handler(event):
    data = event.get("data") if isinstance(event, dict) else None

    if not data:
        return {"ok": False, "error": "data is required"}

    try:
        val = str(data).strip().lower()
        # Handle full domain (encode each label)
        labels = val.split(".")
        encoded_labels = []
        for label in labels:
            try:
                encoded = label.encode("ascii")
                encoded_labels.append(label)
            except UnicodeEncodeError:
                encoded = label.encode("punycode").decode("ascii")
                encoded_labels.append(f"xn--{encoded}")
        result = ".".join(encoded_labels)
        return {"ok": True, "result": result, "original": data}
    except Exception as e:
        return {"ok": False, "error": str(e)}
