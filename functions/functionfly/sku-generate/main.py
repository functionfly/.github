import re


def handler(event):
    product_name = event.get("product_name") if isinstance(event, dict) else None
    category = event.get("category", "")
    variant = event.get("variant", "")
    supplier_code = event.get("supplier_code", "")
    sequence = event.get("sequence")
    if not product_name:
        return {"ok": False, "error": "product_name is required"}
    try:
        def slugify(s, n=4):
            return re.sub(r"[^A-Z0-9]", "", s.upper())[:n]

        parts = []
        if category:
            parts.append(slugify(category, 3))
        parts.append(slugify(product_name, 4))
        if variant:
            parts.append(slugify(variant, 3))
        if supplier_code:
            parts.append(slugify(supplier_code, 3))
        if sequence is not None:
            parts.append(f"{int(sequence):04d}")
        else:
            import random
            parts.append(f"{random.randint(1, 9999):04d}")
        sku = "-".join(p for p in parts if p)
        return {"ok": True, "result": sku, "sku": sku}
    except Exception as e:
        return {"ok": False, "error": str(e)}
