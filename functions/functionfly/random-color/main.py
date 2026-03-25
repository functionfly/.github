import random


def handler(event):
    format_ = event.get("format", "hex") if isinstance(event, dict) else "hex"
    seed = event.get("seed")

    rng = random.Random(seed) if seed is not None else random.Random()
    r, g, b = rng.randint(0, 255), rng.randint(0, 255), rng.randint(0, 255)
    hex_val = f"#{r:02X}{g:02X}{b:02X}"
    css = f"rgb({r},{g},{b})"
    if format_ == "rgb":
        result = css
    elif format_ == "object":
        result = {"r": r, "g": g, "b": b}
    else:
        result = hex_val
    return {"ok": True, "result": result, "hex": hex_val, "css": css, "r": r, "g": g, "b": b}
