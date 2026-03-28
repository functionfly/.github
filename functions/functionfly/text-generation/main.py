import re
import hashlib
from collections import defaultdict

STYLE_CONNECTORS = {
    "formal": ["Furthermore,", "Additionally,", "Moreover,", "In this context,", "It is noteworthy that", "Consequently,", "Therefore,", "As a result,", "In conclusion,", "To summarize,"],
    "casual": ["Also,", "Plus,", "And then,", "So basically,", "You know,", "Like,", "Anyway,", "By the way,", "On top of that,", "The thing is,"],
    "technical": ["Specifically,", "In particular,", "From a technical standpoint,", "The implementation involves", "The algorithm processes", "The system architecture", "The data pipeline", "The computational complexity", "The optimization strategy", "The performance metrics"],
    "creative": ["Suddenly,", "In a twist of fate,", "Beneath the surface,", "Like a whisper in the wind,", "Against all odds,", "In the shadows of", "With a spark of", "Through the mist of", "Beyond the horizon,", "In the realm of"],
}

STYLE_ENDINGS = {
    "formal": ["This represents a significant development in the field.", "The implications are far-reaching and profound.", "Further research is warranted to explore these findings.", "The evidence supports this conclusion.", "This analysis provides valuable insights."],
    "casual": ["Pretty cool, right?", "That's basically the gist of it.", "Makes sense when you think about it.", "It's actually pretty interesting stuff.", "Worth keeping an eye on."],
    "technical": ["The system achieves O(n log n) complexity.", "Performance benchmarks confirm the efficiency gains.", "The implementation follows best practices.", "The architecture ensures scalability and reliability.", "The solution meets all specified requirements."],
    "creative": ["And so the story continues.", "The journey has only just begun.", "What lies ahead remains to be seen.", "The possibilities are endless.", "A new chapter unfolds."],
}


def _build_markov(words, order=2):
    chain = defaultdict(list)
    for i in range(len(words) - order):
        key = tuple(words[i:i+order])
        chain[key].append(words[i+order])
    return chain


def _deterministic_choice(items, seed_str):
    if not items:
        return None
    h = int(hashlib.md5(seed_str.encode()).hexdigest(), 16)
    return items[h % len(items)]


def handler(event):
    prompt = event.get("prompt") if isinstance(event, dict) else None
    if not prompt or not isinstance(prompt, str):
        return {"ok": False, "error": "prompt (string) is required"}
    try:
        max_tokens = min(int(event.get("max_tokens", 50)), 200)
        style = event.get("style", "formal")
        if style not in STYLE_CONNECTORS:
            style = "formal"

        words = re.findall(r'\b\w+\b', prompt)
        if len(words) < 2:
            return {"ok": False, "error": "prompt must contain at least 2 words"}

        # Build Markov chain from prompt
        chain = _build_markov(words, order=min(2, len(words) - 1))

        # Generate continuation
        generated = list(words[-2:]) if len(words) >= 2 else list(words)
        seed = prompt
        for i in range(max_tokens):
            key = tuple(generated[-2:])
            if key in chain:
                next_word = _deterministic_choice(chain[key], seed + str(i))
                generated.append(next_word)
                seed = next_word
            else:
                # Add style connector
                connector = _deterministic_choice(STYLE_CONNECTORS[style], seed + str(i))
                connector_words = connector.split()
                generated.extend(connector_words)
                seed = connector
                # Reset to beginning of prompt words
                if len(words) >= 2:
                    generated.extend(words[:2])
                break

        # Add ending
        ending = _deterministic_choice(STYLE_ENDINGS[style], prompt)
        continuation = " ".join(generated[len(words):])
        full_text = prompt + " " + continuation + " " + ending if continuation else prompt + " " + ending
        full_text = re.sub(r'\s+', ' ', full_text).strip()

        return {
            "ok": True,
            "result": full_text,
            "generated_text": full_text,
            "prompt": prompt,
            "tokens_generated": len(full_text.split()) - len(words),
            "style": style,
            "note": "Template/Markov-based generation — for production use, integrate a language model"
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
