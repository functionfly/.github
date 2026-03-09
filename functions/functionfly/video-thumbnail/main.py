import base64
import io
import subprocess
import tempfile
import os

def handler(event):
    if isinstance(event, dict):
        video_b64 = event.get("video_base64", "")
        offset = event.get("offset_seconds", 0)
    else:
        video_b64, offset = "", 0
    if not video_b64:
        return {"ok": False, "error": "video_base64 is required"}
    try:
        raw = base64.b64decode(str(video_b64).strip(), validate=True)
    except Exception as e:
        return {"ok": False, "error": f"Invalid base64: {e}"}
    try:
        import cv2
        import numpy as np
    except ImportError:
        try:
            with tempfile.NamedTemporaryFile(suffix=".mp4", delete=False) as f:
                f.write(raw)
                path = f.name
            out = subprocess.run(
                ["ffmpeg", "-y", "-ss", str(float(offset)), "-i", path, "-vframes", "1", "-f", "image2pipe", "-vcodec", "png", "pipe:1"],
                capture_output=True, timeout=15
            )
            os.unlink(path)
            if out.returncode != 0 or not out.stdout:
                return {"ok": False, "error": "ffmpeg failed or not installed; install opencv-python or ffmpeg"}
            return {"ok": True, "image_base64": base64.b64encode(out.stdout).decode("ascii")}
        except (FileNotFoundError, subprocess.TimeoutExpired) as e:
            return {"ok": False, "error": "opencv-python or ffmpeg required; pip install opencv-python"}
    try:
        with tempfile.NamedTemporaryFile(suffix=".mp4", delete=False) as f:
            f.write(raw)
            path = f.name
        cap = cv2.VideoCapture(path)
        os.unlink(path)
        if offset > 0:
            cap.set(cv2.CAP_PROP_POS_MSEC, offset * 1000)
        ok, frame = cap.read()
        cap.release()
        if not ok or frame is None:
            return {"ok": False, "error": "Could not read video frame"}
        _, buf = cv2.imencode(".png", frame)
        return {"ok": True, "image_base64": base64.b64encode(buf.tobytes()).decode("ascii")}
    except Exception as e:
        return {"ok": False, "error": str(e)}
