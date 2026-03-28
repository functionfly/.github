import hashlib

IMAGENET_CLASSES = [
    "cat","dog","bird","fish","horse","cow","sheep","elephant","bear","zebra",
    "giraffe","backpack","umbrella","handbag","tie","suitcase","frisbee","skis",
    "snowboard","sports ball","kite","baseball bat","baseball glove","skateboard",
    "surfboard","tennis racket","bottle","wine glass","cup","fork","knife","spoon",
    "bowl","banana","apple","sandwich","orange","broccoli","carrot","hot dog",
    "pizza","donut","cake","chair","couch","potted plant","bed","dining table",
    "toilet","tv","laptop","mouse","remote","keyboard","cell phone","microwave",
    "oven","toaster","sink","refrigerator","book","clock","vase","scissors",
    "teddy bear","hair drier","toothbrush","person","bicycle","car","motorcycle",
    "airplane","bus","train","truck","boat","traffic light","fire hydrant",
    "stop sign","parking meter","bench","bird","cat","dog","horse","sheep","cow",
    "elephant","bear","zebra","giraffe","backpack","umbrella","handbag","tie",
]

def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be an object"}
    image_url = event.get("image_url", "")
    image_base64 = event.get("image_base64", "")
    top_k = int(event.get("top_k", 5))
    if not image_url and not image_base64:
        return {"ok": False, "error": "image_url or image_base64 is required"}
    try:
        seed = image_url or image_base64[:50]
        h = hashlib.sha256(seed.encode()).digest()
        predictions = []
        used = set()
        for i in range(min(top_k, len(IMAGENET_CLASSES))):
            idx = (h[i % 32] + i * 7) % len(IMAGENET_CLASSES)
            while idx in used:
                idx = (idx + 1) % len(IMAGENET_CLASSES)
            used.add(idx)
            score = round(max(0.01, (h[i % 32] / 255.0) * (1.0 - i * 0.15)), 4)
            predictions.append({"label": IMAGENET_CLASSES[idx], "score": score, "rank": i + 1})
        predictions.sort(key=lambda x: x["score"], reverse=True)
        return {
            "ok": True,
            "result": predictions,
            "predictions": predictions,
            "top_label": predictions[0]["label"] if predictions else None,
            "model": "mock-imagenet-v1",
            "note": "Mock classification — for production use, integrate a vision model"
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
