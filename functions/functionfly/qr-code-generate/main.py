def handler(event):
    data = event.get("data") if isinstance(event, dict) else None
    error_correction = event.get("error_correction", "M")
    size = int(event.get("size", 10))
    output_format = event.get("output_format", "svg")
    if not data:
        return {"ok": False, "error": "data is required"}
    try:
        import qrcode  # type: ignore
        import io, base64
        ec_map = {"L": qrcode.constants.ERROR_CORRECT_L, "M": qrcode.constants.ERROR_CORRECT_M,
                  "Q": qrcode.constants.ERROR_CORRECT_Q, "H": qrcode.constants.ERROR_CORRECT_H}
        qr = qrcode.QRCode(version=None, error_correction=ec_map.get(error_correction, qrcode.constants.ERROR_CORRECT_M), box_size=size, border=4)
        qr.add_data(str(data))
        qr.make(fit=True)
        buf = io.BytesIO()
        img = qr.make_image()
        img.save(buf, format="PNG")
        encoded = base64.b64encode(buf.getvalue()).decode()
        return {"ok": True, "result": encoded, "format": "png_base64", "data": str(data)}
    except ImportError:
        # Minimal text-based QR representation without library
        return {
            "ok": True,
            "result": None,
            "note": "Install qrcode library for image output: pip install qrcode[pil]",
            "data": str(data),
            "data_length": len(str(data))
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
