import re

# Common phrase translations (en -> target)
PHRASE_DICT = {
    "es": {
        "hello": "hola", "hi": "hola", "goodbye": "adiós", "bye": "adiós",
        "thank you": "gracias", "thanks": "gracias", "please": "por favor",
        "yes": "sí", "no": "no", "good": "bueno", "bad": "malo",
        "morning": "mañana", "afternoon": "tarde", "evening": "noche",
        "how are you": "cómo estás", "i am fine": "estoy bien",
        "what is your name": "cómo te llamas", "my name is": "me llamo",
        "i love you": "te amo", "i like": "me gusta", "i want": "quiero",
        "i need": "necesito", "i have": "tengo", "i am": "soy",
        "the": "el", "a": "un", "and": "y", "or": "o", "but": "pero",
        "in": "en", "on": "en", "at": "en", "to": "a", "from": "de",
        "with": "con", "without": "sin", "for": "para", "of": "de",
        "is": "es", "are": "son", "was": "era", "were": "eran",
        "have": "tener", "has": "tiene", "had": "tenía",
        "will": "va a", "would": "podría", "could": "podría",
        "can": "puede", "cannot": "no puede", "not": "no",
        "this": "este", "that": "ese", "these": "estos", "those": "esos",
        "here": "aquí", "there": "allí", "where": "dónde",
        "when": "cuándo", "why": "por qué", "how": "cómo", "what": "qué",
        "who": "quién", "which": "cuál", "all": "todo", "some": "algunos",
        "many": "muchos", "few": "pocos", "more": "más", "less": "menos",
        "big": "grande", "small": "pequeño", "new": "nuevo", "old": "viejo",
        "good morning": "buenos días", "good afternoon": "buenas tardes",
        "good evening": "buenas noches", "good night": "buenas noches",
        "how are you": "¿cómo estás?", "hello how are you": "hola, ¿cómo estás?",
        "water": "agua", "food": "comida", "house": "casa", "car": "coche",
        "dog": "perro", "cat": "gato", "book": "libro", "time": "tiempo",
        "day": "día", "night": "noche", "year": "año", "month": "mes",
        "week": "semana", "hour": "hora", "minute": "minuto",
        "man": "hombre", "woman": "mujer", "child": "niño", "people": "gente",
        "city": "ciudad", "country": "país", "world": "mundo",
        "work": "trabajo", "school": "escuela", "money": "dinero",
        "love": "amor", "friend": "amigo", "family": "familia",
        "happy": "feliz", "sad": "triste", "angry": "enojado",
        "beautiful": "hermoso", "ugly": "feo", "fast": "rápido", "slow": "lento",
        "hot": "caliente", "cold": "frío", "hard": "duro", "soft": "suave",
    },
    "fr": {
        "hello": "bonjour", "hi": "salut", "goodbye": "au revoir", "bye": "au revoir",
        "thank you": "merci", "thanks": "merci", "please": "s'il vous plaît",
        "yes": "oui", "no": "non", "good": "bon", "bad": "mauvais",
        "morning": "matin", "afternoon": "après-midi", "evening": "soir",
        "how are you": "comment allez-vous", "i am fine": "je vais bien",
        "i love you": "je t'aime", "i like": "j'aime", "i want": "je veux",
        "i need": "j'ai besoin", "i have": "j'ai", "i am": "je suis",
        "the": "le", "a": "un", "and": "et", "or": "ou", "but": "mais",
        "in": "dans", "on": "sur", "at": "à", "to": "à", "from": "de",
        "with": "avec", "without": "sans", "for": "pour", "of": "de",
        "is": "est", "are": "sont", "was": "était", "were": "étaient",
        "good morning": "bonjour", "good afternoon": "bon après-midi",
        "good evening": "bonsoir", "good night": "bonne nuit",
        "water": "eau", "food": "nourriture", "house": "maison", "car": "voiture",
        "dog": "chien", "cat": "chat", "book": "livre", "time": "temps",
        "day": "jour", "night": "nuit", "year": "an", "month": "mois",
        "man": "homme", "woman": "femme", "child": "enfant", "people": "gens",
        "city": "ville", "country": "pays", "world": "monde",
        "love": "amour", "friend": "ami", "family": "famille",
        "happy": "heureux", "sad": "triste", "beautiful": "beau",
    },
    "de": {
        "hello": "hallo", "hi": "hallo", "goodbye": "auf wiedersehen", "bye": "tschüss",
        "thank you": "danke", "thanks": "danke", "please": "bitte",
        "yes": "ja", "no": "nein", "good": "gut", "bad": "schlecht",
        "morning": "morgen", "afternoon": "nachmittag", "evening": "abend",
        "how are you": "wie geht es ihnen", "i am fine": "es geht mir gut",
        "i love you": "ich liebe dich", "i like": "ich mag", "i want": "ich will",
        "i need": "ich brauche", "i have": "ich habe", "i am": "ich bin",
        "the": "der", "a": "ein", "and": "und", "or": "oder", "but": "aber",
        "in": "in", "on": "auf", "at": "bei", "to": "zu", "from": "von",
        "with": "mit", "without": "ohne", "for": "für", "of": "von",
        "is": "ist", "are": "sind", "was": "war", "were": "waren",
        "good morning": "guten morgen", "good afternoon": "guten nachmittag",
        "good evening": "guten abend", "good night": "gute nacht",
        "water": "wasser", "food": "essen", "house": "haus", "car": "auto",
        "dog": "hund", "cat": "katze", "book": "buch", "time": "zeit",
        "day": "tag", "night": "nacht", "year": "jahr", "month": "monat",
        "man": "mann", "woman": "frau", "child": "kind", "people": "leute",
        "city": "stadt", "country": "land", "world": "welt",
        "love": "liebe", "friend": "freund", "family": "familie",
        "happy": "glücklich", "sad": "traurig", "beautiful": "schön",
    },
    "it": {
        "hello": "ciao", "hi": "ciao", "goodbye": "arrivederci", "bye": "ciao",
        "thank you": "grazie", "thanks": "grazie", "please": "per favore",
        "yes": "sì", "no": "no", "good": "buono", "bad": "cattivo",
        "morning": "mattina", "afternoon": "pomeriggio", "evening": "sera",
        "how are you": "come stai", "i am fine": "sto bene",
        "i love you": "ti amo", "i like": "mi piace", "i want": "voglio",
        "i need": "ho bisogno", "i have": "ho", "i am": "sono",
        "the": "il", "a": "un", "and": "e", "or": "o", "but": "ma",
        "good morning": "buongiorno", "good afternoon": "buon pomeriggio",
        "good evening": "buonasera", "good night": "buonanotte",
        "water": "acqua", "food": "cibo", "house": "casa", "car": "macchina",
        "dog": "cane", "cat": "gatto", "book": "libro", "time": "tempo",
        "day": "giorno", "night": "notte", "year": "anno", "month": "mese",
        "man": "uomo", "woman": "donna", "child": "bambino", "people": "gente",
        "city": "città", "country": "paese", "world": "mondo",
        "love": "amore", "friend": "amico", "family": "famiglia",
        "happy": "felice", "sad": "triste", "beautiful": "bello",
    },
    "pt": {
        "hello": "olá", "hi": "oi", "goodbye": "adeus", "bye": "tchau",
        "thank you": "obrigado", "thanks": "obrigado", "please": "por favor",
        "yes": "sim", "no": "não", "good": "bom", "bad": "mau",
        "morning": "manhã", "afternoon": "tarde", "evening": "noite",
        "how are you": "como vai você", "i am fine": "estou bem",
        "i love you": "eu te amo", "i like": "eu gosto", "i want": "eu quero",
        "i need": "eu preciso", "i have": "eu tenho", "i am": "eu sou",
        "the": "o", "a": "um", "and": "e", "or": "ou", "but": "mas",
        "good morning": "bom dia", "good afternoon": "boa tarde",
        "good evening": "boa noite", "good night": "boa noite",
        "water": "água", "food": "comida", "house": "casa", "car": "carro",
        "dog": "cachorro", "cat": "gato", "book": "livro", "time": "tempo",
        "day": "dia", "night": "noite", "year": "ano", "month": "mês",
        "man": "homem", "woman": "mulher", "child": "criança", "people": "pessoas",
        "city": "cidade", "country": "país", "world": "mundo",
        "love": "amor", "friend": "amigo", "family": "família",
        "happy": "feliz", "sad": "triste", "beautiful": "bonito",
    },
}

SUPPORTED_LANGS = list(PHRASE_DICT.keys()) + ["en"]


def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be an object"}
    text = event.get("text")
    target_lang = event.get("target_lang")
    if not text or not isinstance(text, str):
        return {"ok": False, "error": "text (string) is required"}
    if not target_lang or not isinstance(target_lang, str):
        return {"ok": False, "error": "target_lang (string) is required"}
    try:
        source_lang = event.get("source_lang", "en")
        target_lang = target_lang.lower()[:2]
        if target_lang == "en" or target_lang == source_lang:
            return {
                "ok": True,
                "result": text,
                "translated_text": text,
                "source_lang": source_lang,
                "target_lang": target_lang,
                "note": "Source and target language are the same"
            }
        if target_lang not in PHRASE_DICT:
            return {
                "ok": True,
                "result": text,
                "translated_text": f"[{target_lang.upper()} translation not available] {text}",
                "source_lang": source_lang,
                "target_lang": target_lang,
                "supported_languages": SUPPORTED_LANGS,
                "note": f"Translation to '{target_lang}' not supported in this mock implementation"
            }
        dictionary = PHRASE_DICT[target_lang]
        t_lower = text.lower()
        # Try phrase-level translation first
        result = t_lower
        for phrase, translation in sorted(dictionary.items(), key=lambda x: -len(x[0])):
            result = result.replace(phrase, translation)
        # Capitalize first letter
        if result:
            result = result[0].upper() + result[1:]
        return {
            "ok": True,
            "result": result,
            "translated_text": result,
            "source_lang": source_lang,
            "target_lang": target_lang,
            "note": "Dictionary-based translation — for production use, integrate a translation API"
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
