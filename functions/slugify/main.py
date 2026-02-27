import re
import unicodedata

def handler(event):
    """
    Convert a string to a URL-friendly slug.

    Args:
        event: A string or dict with 'text' or the string directly

    Returns:
        dict with 'slug' key containing the slugified string
    """
    # Handle input - can be string or dict
    if isinstance(event, dict):
        text = event.get('text', '')
    else:
        text = str(event)

    if not text:
        return {'slug': ''}

    # Convert to lowercase
    text = text.lower()

    # Normalize unicode characters (e.g., 'é' -> 'e')
    text = unicodedata.normalize('NFD', text)
    text = ''.join(c for c in text if unicodedata.category(c) != 'Mn')

    # Replace spaces and special chars with hyphens
    text = re.sub(r'[^\w\s-]', '', text)
    text = re.sub(r'[-\s]+', '-', text)

    # Remove leading/trailing hyphens
    slug = text.strip('-')

    return {'slug': slug}
