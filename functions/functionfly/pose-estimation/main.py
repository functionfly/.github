import hashlib

BODY_KEYPOINTS = ["nose","left_eye","right_eye","left_ear","right_ear","left_shoulder",
    "right_shoulder","left_elbow","right_elbow","left_wrist","right_wrist","left_hip",
    "right_hip","left_knee","right_knee","left_ankle","right_ankle"]

HAND_KEYPOINTS = ["wrist","thumb_cmc","thumb_mcp","thumb_ip","thumb_tip",
    "index_mcp","index_pip","index_dip","index_tip","middle_mcp","middle_pip",
    "middle_dip","middle_tip","ring_mcp","ring_pip","ring_dip","ring_tip",
    "pinky_mcp","pinky_pip","pinky_dip","pinky_tip"]

def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be an object"}
    image_url = event.get("image_url", "")
    image_base64 = event.get("image_base64", "")
    model = event.get("model", "body")
    if not image_url and not image_base64:
        return {"ok": False, "error": "image_url or image_base64 is required"}
    try:
        seed = image_url or image_base64[:50]
        h = hashlib.sha256(seed.encode()).digest()
        keypoints_list = HAND_KEYPOINTS if model == "hand" else BODY_KEYPOINTS
        keypoints = {}
        for i, kp in enumerate(keypoints_list):
            x = round((h[i % 32] / 255.0) * 0.8 + 0.1, 4)
            y = round((h[(i+1) % 32] / 255.0) * 0.8 + 0.1, 4)
            confidence = round(0.7 + (h[(i+2) % 32] / 255.0) * 0.3, 4)
            keypoints[kp] = {"x": x, "y": y, "confidence": confidence}
        return {
            "ok": True,
            "result": keypoints,
            "keypoints": keypoints,
            "model_type": model,
            "num_keypoints": len(keypoints),
            "model": "mock-openpose-v1",
            "note": "Mock pose estimation — for production use, integrate OpenPose or MediaPipe"
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
