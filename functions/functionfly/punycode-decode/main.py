def handler(event):
    data = event.get("data") if isinstance(event, dict) else None

    if not data:
        return {"ok": False, "error": "data is required"}

    try:
        val = str(data).strip().lower()
        labels = val.split(".")
        decoded_labels = []
        for label in labels:
            if label.startswith("xn--"):
                puny = label[4:]
                try:
                    decoded_labels.append(puny.encode("ascii").decode("punycode"))
                except Exception:
                    decoded_labels.append(label)
            else:
                decoded_labels.append(label)
        result = ".".join(decoded_labels)
        return {"ok": True, "result": result, "original": data}
    except Exception as e:
        return {"ok": False, "error": str(e)}
