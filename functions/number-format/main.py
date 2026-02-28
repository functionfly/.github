def handler(event):
    try:
        if isinstance(event, dict):
            number = event.get("number")
            decimal_places = event.get("decimal_places", 2)
            thousands_separator = event.get("thousands_separator", ",")
            decimal_separator = event.get("decimal_separator", ".")
        else:
            return {"ok": False, "error": "Missing required field: number"}

        if number is None:
            return {"ok": False, "error": "Missing required field: number"}

        try:
            number = float(number)
        except (TypeError, ValueError):
            return {"ok": False, "error": "number must be a numeric value"}

        try:
            decimal_places = int(decimal_places)
            if decimal_places < 0:
                decimal_places = 0
        except (TypeError, ValueError):
            return {"ok": False, "error": "decimal_places must be an integer"}

        if not isinstance(thousands_separator, str):
            thousands_separator = str(thousands_separator)
        if not isinstance(decimal_separator, str):
            decimal_separator = str(decimal_separator)

        # Format with fixed decimal places
        formatted = f"{number:,.{decimal_places}f}"

        # Replace default separators with custom ones
        # Default format uses ',' for thousands and '.' for decimal
        # We need to swap them if custom separators differ
        if thousands_separator != "," or decimal_separator != ".":
            # Use a placeholder to avoid collision
            formatted = formatted.replace(",", "\x00")
            formatted = formatted.replace(".", decimal_separator)
            formatted = formatted.replace("\x00", thousands_separator)

        return {"ok": True, "result": formatted}
    except Exception as e:
        return {"ok": False, "error": str(e)}
