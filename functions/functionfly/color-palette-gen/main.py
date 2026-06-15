"""Color Palette Generator - Generate harmonious color palettes."""
import re
import colorsys
from datetime import datetime
from typing import Any


COLOR_NAMES = {
    "red": (255, 0, 0), "maroon": (128, 0, 0), "crimson": (220, 20, 60),
    "orange": (255, 165, 0), "gold": (255, 215, 0), "amber": (255, 191, 0),
    "yellow": (255, 255, 0), "lime": (0, 255, 0), "green": (0, 128, 0),
    "emerald": (80, 200, 120), "teal": (0, 128, 128), "cyan": (0, 255, 255),
    "sky": (135, 206, 235), "blue": (0, 0, 255), "navy": (0, 0, 128),
    "purple": (128, 0, 128), "violet": (238, 130, 238), "pink": (255, 192, 203),
    "magenta": (255, 0, 255), "white": (255, 255, 255), "gray": (128, 128, 128),
    "silver": (192, 192, 192), "black": (0, 0, 0), "brown": (165, 42, 42),
    "coral": (255, 127, 80), "salmon": (250, 128, 114), "indigo": (75, 0, 130),
    "turquoise": (64, 224, 208), "beige": (245, 245, 220), "ivory": (255, 255, 240),
    "lavender": (230, 230, 250), "burgundy": (128, 0, 32), "ochre": (204, 119, 34),
    "mint": (189, 252, 201), "peach": (255, 218, 185), "plum": (221, 160, 221),
}


def hex_to_rgb(hex_color: str) -> tuple:
    """Convert hex color to RGB tuple."""
    hex_color = hex_color.lstrip('#')
    if len(hex_color) == 3:
        hex_color = ''.join([c*2 for c in hex_color])
    if len(hex_color) != 6:
        raise ValueError(f"Invalid hex color: {hex_color}")
    return tuple(int(hex_color[i:i+2], 16) for i in (0, 2, 4))


def rgb_to_hex(rgb: tuple) -> str:
    """Convert RGB tuple to hex string."""
    return "#{:02x}{:02x}{:02x}".format(int(rgb[0]), int(rgb[1]), int(rgb[2]))


def find_closest_color_name(rgb: tuple) -> str:
    """Find the closest named color to the given RGB value."""
    min_dist = float('inf')
    closest_name = "unknown"

    for name, color_rgb in COLOR_NAMES.items():
        dist = sum((a - b) ** 2 for a, b in zip(rgb, color_rgb))
        if dist < min_dist:
            min_dist = dist
            closest_name = name

    return closest_name.replace("_", " ").title()


def adjust_brightness(rgb: tuple, factor: float) -> tuple:
    """Adjust RGB brightness by factor."""
    return tuple(max(0, min(255, int(c * factor))) for c in rgb)


def generate_complementary(base_rgb: tuple, num_colors: int) -> list:
    """Generate complementary palette."""
    h, s, v = colorsys.rgb_to_hsv(base_rgb[0]/255, base_rgb[1]/255, base_rgb[2]/255)
    comp_h = (h + 0.5) % 1.0

    colors = []
    for i in range(num_colors):
        if i % 2 == 0:
            new_h, new_s, new_v = h, s, max(0.3, v - (i * 0.15))
        else:
            new_h, new_s, new_v = comp_h, s, max(0.3, v - ((i-1) * 0.15))
        rgb = colorsys.hsv_to_rgb(new_h, new_s, new_v)
        colors.append((int(rgb[0]*255), int(rgb[1]*255), int(rgb[2]*255)))
    return colors


def generate_analogous(base_rgb: tuple, num_colors: int) -> list:
    """Generate analogous palette."""
    h, s, v = colorsys.rgb_to_hsv(base_rgb[0]/255, base_rgb[1]/255, base_rgb[2]/255)
    angle_step = 0.1

    colors = []
    start_idx = -(num_colors // 2)
    for i in range(num_colors):
        new_h = (h + (start_idx + i) * angle_step) % 1.0
        new_s = max(0.3, s - (abs(start_idx + i) * 0.08))
        new_v = max(0.4, v - (abs(start_idx + i) * 0.05))
        rgb = colorsys.hsv_to_rgb(new_h, new_s, new_v)
        colors.append((int(rgb[0]*255), int(rgb[1]*255), int(rgb[2]*255)))
    return colors


def generate_triadic(base_rgb: tuple, num_colors: int) -> list:
    """Generate triadic palette."""
    h, s, v = colorsys.rgb_to_hsv(base_rgb[0]/255, base_rgb[1]/255, base_rgb[2]/255)

    colors = []
    for i in range(num_colors):
        if i < 3:
            new_h = (h + (i * 0.33)) % 1.0
            new_s, new_v = s, v
        else:
            new_h = (h + ((i - 3) * 0.33)) % 1.0
            new_s = max(0.4, s - 0.2)
            new_v = max(0.5, v - 0.15)
        rgb = colorsys.hsv_to_rgb(new_h, new_s, new_v)
        colors.append((int(rgb[0]*255), int(rgb[1]*255), int(rgb[2]*255)))
    return colors


def generate_monochromatic(base_rgb: tuple, num_colors: int) -> list:
    """Generate monochromatic palette."""
    h, s, v = colorsys.rgb_to_hsv(base_rgb[0]/255, base_rgb[1]/255, base_rgb[2]/255)

    colors = []
    step = 0.15
    start_v = 0.25

    for i in range(num_colors):
        new_v = min(1.0, start_v + (i * step))
        new_s = max(0.2, s - (i * 0.05))
        rgb = colorsys.hsv_to_rgb(h, new_s, new_v)
        colors.append((int(rgb[0]*255), int(rgb[1]*255), int(rgb[2]*255)))
    return colors


def handler(event: dict) -> dict:
    """Generate a color palette based on input parameters."""
    try:
        base_color = event.get("base_color")
        palette_type = event.get("palette_type", "complementary")
        num_colors = event.get("num_colors", 5)

        if not base_color:
            return {"ok": False, "error": "base_color is required"}

        base_color = base_color.strip()
        if not re.match(r'^#?[0-9A-Fa-f]{3}$|^#?[0-9A-Fa-f]{6}$', base_color):
            return {"ok": False, "error": "base_color must be a valid hex color (e.g., #FF5733 or #F53)"}

        if palette_type not in ["complementary", "analogous", "triadic", "monochromatic"]:
            return {"ok": False, "error": "palette_type must be one of: complementary, analogous, triadic, monochromatic"}

        if not isinstance(num_colors, int) or num_colors < 2 or num_colors > 10:
            return {"ok": False, "error": "num_colors must be an integer between 2 and 10"}

        if not base_color.startswith('#'):
            base_color = '#' + base_color

        base_rgb = hex_to_rgb(base_color)

        generators = {
            "complementary": generate_complementary,
            "analogous": generate_analogous,
            "triadic": generate_triadic,
            "monochromatic": generate_monochromatic
        }

        raw_colors = generators[palette_type](base_rgb, num_colors)

        colors = []
        for rgb in raw_colors:
            colors.append({
                "hex": rgb_to_hex(rgb),
                "rgb": {"r": rgb[0], "g": rgb[1], "b": rgb[2]},
                "name": find_closest_color_name(rgb)
            })

        palette_names = {
            "complementary": f"Complementary with {find_closest_color_name(base_rgb)}",
            "analogous": f"Analogous {find_closest_color_name(base_rgb)} Harmony",
            "triadic": f"Triadic {find_closest_color_name(base_rgb)} Blend",
            "monochromatic": f"{find_closest_color_name(base_rgb)} Shades"
        }

        return {
            "ok": True,
            "base_color": base_color.upper(),
            "palette_type": palette_type,
            "palette_name": palette_names[palette_type],
            "num_colors": len(colors),
            "colors": colors,
            "generated_at": datetime.now().isoformat()
        }

    except Exception as e:
        return {"ok": False, "error": f"Failed to generate palette: {str(e)}"}


if __name__ == "__main__":
    result = handler({
        "base_color": "#3B82F6",
        "palette_type": "complementary",
        "num_colors": 5
    })
    print(result)
