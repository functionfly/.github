MORSE_CODE = {
    'A': '.-', 'B': '-...', 'C': '-.-.', 'D': '-..', 'E': '.', 'F': '..-.',
    'G': '--.', 'H': '....', 'I': '..', 'J': '.---', 'K': '-.-', 'L': '.-..',
    'M': '--', 'N': '-.', 'O': '---', 'P': '.--.', 'Q': '--.-', 'R': '.-.',
    'S': '...', 'T': '-', 'U': '..-', 'V': '...-', 'W': '.--', 'X': '-..-',
    'Y': '-.--', 'Z': '--..', '0': '-----', '1': '.----', '2': '..---',
    '3': '...--', '4': '....-', '5': '.....', '6': '-....', '7': '--...',
    '8': '---..', '9': '----.', '.': '.-.-.-', ',': '--..--', '?': '..--..',
    "'": '.----.', '!': '-.-.--', '/': '-..-.', '(': '-.--.', ')': '-.--.-',
    '&': '.-...', ':': '---...', ';': '-.-.-.', '=': '-...-', '+': '.-.-.',
    '-': '-....-', '_': '..--.-', '"': '.-..-.', '$': '...-..-', '@': '.--.-.',
    ' ': '/'
}

def encode_morse(text: str) -> str:
    """Encode text to Morse code"""
    text = text.upper()
    result = []
    for char in text:
        if char in MORSE_CODE:
            result.append(MORSE_CODE[char])
        else:
            result.append('')
    return ' '.join(result)

def handler(event):
    try:
        text = event.get("text", "") if isinstance(event, dict) else ""
        if not text:
            return {"ok": False, "error": "text is required"}
        morse = encode_morse(text)
        return {"ok": True, "morse": morse}
    except Exception as e:
        return {"ok": False, "error": str(e)}
