import re


def check_isbn10(isbn):
    isbn = re.sub(r'[\-\s]', '', isbn)
    if len(isbn) != 10:
        return False
    if not re.match(r'^\d{9}[\dXx]$', isbn):
        return False
    total = sum((10 - i) * (10 if c in 'Xx' else int(c)) for i, c in enumerate(isbn))
    return total % 11 == 0


def check_isbn13(isbn):
    isbn = re.sub(r'[\-\s]', '', isbn)
    if len(isbn) != 13 or not isbn.isdigit():
        return False
    weights = [1, 3] * 6
    total = sum(int(d) * w for d, w in zip(isbn[:12], weights))
    check = (10 - total % 10) % 10
    return check == int(isbn[12])


def handler(event):
    value = event.get("value") if isinstance(event, dict) else None
    if value is None:
        return {"ok": False, "error": "value is required"}

    val = str(value).strip()
    clean = re.sub(r'[\-\s]', '', val)

    if len(clean) == 10:
        valid = check_isbn10(val)
        isbn_type = "ISBN-10" if valid else None
    elif len(clean) == 13:
        valid = check_isbn13(val)
        isbn_type = "ISBN-13" if valid else None
    else:
        valid = False
        isbn_type = None

    return {"ok": True, "value": value, "result": valid, "isbn_type": isbn_type}
