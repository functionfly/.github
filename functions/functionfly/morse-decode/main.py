MORSE_CODE_REVERSE = {
    '.-': 'A', '-...': 'B', '-.-.': 'C', '-..': 'D', '.': 'E', '..-.': 'F',
    '--.': 'G', '....': 'H', '..': 'I', '.---': 'J', '-.-': 'K', '.-..': 'L',
    '--': 'M', '-.': 'N', '---': 'O', '.--.': 'P', '--.-': 'Q', '.-.': 'R',
    '...': 'S', '-': 'T', '..-': 'U', '...-': 'V', '.--': 'W', '-..-': 'X',
    '-.--': 'Y', '--..': 'Z', '-----': '0', '.----': '1', '..---': '2',
    '...--': '3', '....-': '4', '.....': '5', '-....': '6', '--...': '7',
    '---..': '8', '----.': '9', '.-.-.-': '.', '--..--': ',', '..--..': '?',
    '.----.': "'", '-.-.--': '!', '-..-.': '/', '-.--.': '(', '-.--.-': ')',
    '.-...': '&', '---...': ':', '-.-.-.': ';', '-...-': '=', '.-.-.': '+',
    '-....-': '-', '..--.-': '_', '.-..-.': '"', '...-..-': '$', '.--.-.': '@'
}

def decode_morse(morse: str) -> str:
    """Decode Morse code to text"""
    words = morse.split(' / ')
    result = []
    for word in words:
        letters = word.split(' ')
        decoded_word = ''
        for letter in letters:
            if letter in MORSE_CODE_REVERSE:
                decoded_word += MORSE_CODE_REVERSE[letter]
            elif letter == '':
                continue
            else:
                decoded_word += '?'
        result.append(decoded_word)
    return ' '.join(result)

def handler(event):
    try:
        morse = event.get("morse", "") if isinstance(event, dict) else ""
        if not morse:
            return {"ok": False, "error": "morse is required"}
        text = decode_morse(morse)
        return {"ok": True, "text": text}
    except Exception as e:
        return {"ok": False, "error": str(e)}
