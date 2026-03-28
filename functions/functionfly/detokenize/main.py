import re

NO_SPACE_BEFORE = {'.', ',', '!', '?', ';', ':', ')', ']', '}', "'s", "'t", "'re", "'ve", "'ll", "'d", "'m", "n't"}
NO_SPACE_AFTER = {'(', '[', '{', '$', '#', '@'}


def handler(event):
    tokens = event.get("tokens") if isinstance(event, dict) else None
    if not tokens or not isinstance(tokens, list):
        return {"ok": False, "error": "tokens (array of strings) is required"}
    try:
        str_tokens = [str(t) for t in tokens]
        if not str_tokens:
            return {"ok": True, "result": "", "text": ""}
        result = str_tokens[0]
        for i in range(1, len(str_tokens)):
            token = str_tokens[i]
            prev = str_tokens[i - 1]
            # Determine spacing
            if token in NO_SPACE_BEFORE:
                result += token
            elif prev in NO_SPACE_AFTER:
                result += token
            elif token.startswith("'"):
                result += token
            elif re.match(r'^[^\w]', token) and not re.match(r'^\w', prev[-1] if prev else ''):
                result += token
            else:
                result += " " + token
        return {
            "ok": True,
            "result": result,
            "text": result,
            "token_count": len(str_tokens)
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
