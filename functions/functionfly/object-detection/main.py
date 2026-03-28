import hashlib

COCO_CLASSES = ["person","bicycle","car","motorcycle","airplane","bus","train","truck","boat",
    "traffic light","fire hydrant","stop sign","parking meter","bench","bird","cat","dog",
    "horse","sheep","cow","elephant","bear","zebra","giraffe","backpack","umbrella","handbag",
    "tie","suitcase","frisbee","skis","snowboard","sports ball","kite","baseball bat",
    "baseball glove","skateboard","surfboard","tennis racket","bottle","wine glass","cup",
    "fork","knife","spoon","bowl","banana","apple","sandwich","orange","broccoli","carrot",
    "hot dog","pizza","donut","cake","chair","couch","potted plant","bed","dining table",
    "toilet","tv","laptop","mouse","remote","keyboard","cell phone","microwave","oven",
    "toaster","sink","refrigerator","book","clock","vase","scissors","teddy bear",
    "hair drier","toothbrush"]

def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be an object"}
    image_url = event.get("image_url", "")
    image_base64 = event.get("image_base64", "")
    threshold = float(event.get("confidence_threshold", 0.5))
    if not image_url and not image_base64:
        return {"ok": False, "error": "image_url or image_base64 is required"}
    try:
        seed = image_url or image_base64[:50]
        h = hashlib.sha256(seed.encode()).digest()
        num_objects = (h[0] % 5) + 1
        detections = []
        for i in range(num_objects):
            confidence = round(0.5 + (h[i % 32] / 255.0) * 0.5, 4)
            if confidence < threshold:
                continue
            class_idx = (h[(i+1) % 32] + i * 13) % len(COCO_CLASSES)
            x1 = round((h[(i+2) % 32] / 255.0) * 0.5, 4)
            y1 = round((h[(i+3) % 32] / 255.0) * 0.5, 4)
            w = round(0.1 + (h[(i+4) % 32] / 255.0) * 0.4, 4)
            h_box = round(0.1 + (h[(i+5) % 32] / 255.0) * 0.4, 4)
            detections.append({
                "label": COCO_CLASSES[class_idx],
                "confidence": confidence,
                "bbox": {"x1": x1, "y1": y1, "x2": round(x1 + w, 4), "y2": round(y1 + h_box, 4), "width": w, "height": h_box},
                "id": i
            })
        return {
            "ok": True,
            "result": detections,
            "detections": detections,
            "count": len(detections),
            "model": "mock-yolo-v1",
            "note": "Mock detection — for production use, integrate YOLO or similar"
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
