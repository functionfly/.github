"""AI Language Translator - Simple rule-based translation dictionary."""
import random


TRANSLATION_DICT = {
    "en": {
        "es": {
            "hello": "hola", "goodbye": "adiós", "yes": "sí", "no": "no", "please": "por favor",
            "thank you": "gracias", "welcome": "bienvenido", "good": "bueno", "bad": "malo",
            "day": "día", "night": "noche", "morning": "mañana", "evening": "tarde",
            "water": "agua", "food": "comida", "house": "casa", "car": "carro", "book": "libro",
            "computer": "computadora", "phone": "teléfono", "work": "trabajo", "home": "hogar",
            "family": "familia", "friend": "amigo", "love": "amor", "happy": "feliz",
            "sad": "triste", "big": "grande", "small": "pequeño", "new": "nuevo",
            "old": "viejo", "hot": "caliente", "cold": "frío", "open": "abierto",
            "closed": "cerrado", "help": "ayuda", "save": "guardar", "delete": "borrar",
            "create": "crear", "update": "actualizar", "search": "buscar", "find": "encontrar",
            "learn": "aprender", "teach": "enseñar", "write": "escribir", "read": "leer",
            "speak": "hablar", "listen": "escuchar", "see": "ver", "hear": "oír",
            "walk": "caminar", "run": "correr", "eat": "comer", "drink": "beber",
            "sleep": "dormir", "wake": "despertar", "start": "comenzar", "stop": "parar",
            "buy": "comprar", "sell": "vender", "give": "dar", "take": "tomar",
            "make": "hacer", "want": "querer", "need": "necesitar", "know": "saber",
            "think": "pensar", "believe": "creer", "understand": "entender", "remember": "recordar",
            "forget": "olvidar", "live": "vivir", "die": "morir", "grow": "crecer",
            "change": "cambiar", "stay": "quedarse", "leave": "salir", "arrive": "llegar",
            "time": "tiempo", "money": "dinero", "people": "personas", "world": "mundo",
            "life": "vida", "way": "camino", "thing": "cosa", "place": "lugar",
            "year": "año", "week": "semana", "month": "mes", "today": "hoy",
            "tomorrow": "mañana", "yesterday": "ayer", "now": "ahora", "later": "después",
            "always": "siempre", "never": "nunca", "sometimes": "a veces", "often": "a menudo",
            "here": "aquí", "there": "allí", "where": "dónde", "what": "qué",
            "who": "quién", "why": "por qué", "how": "cómo", "when": "cuándo"
        },
        "fr": {
            "hello": "bonjour", "goodbye": "au revoir", "yes": "oui", "no": "non", "please": "s'il vous plaît",
            "thank you": "merci", "welcome": "bienvenue", "good": "bon", "bad": "mauvais",
            "day": "jour", "night": "nuit", "morning": "matin", "evening": "soir",
            "water": "eau", "food": "nourriture", "house": "maison", "car": "voiture", "book": "livre",
            "computer": "ordinateur", "phone": "téléphone", "work": "travail", "home": "maison",
            "family": "famille", "friend": "ami", "love": "amour", "happy": "heureux",
            "sad": "triste", "big": "grand", "small": "petit", "new": "nouveau",
            "old": "vieux", "hot": "chaud", "cold": "froid", "open": "ouvert",
            "closed": "fermé", "help": "aide", "save": "sauvegarder", "delete": "supprimer",
            "create": "créer", "update": "mettre à jour", "search": "rechercher", "find": "trouver",
            "learn": "apprendre", "teach": "enseigner", "write": "écrire", "read": "lire",
            "speak": "parler", "listen": "écouter", "see": "voir", "hear": "entendre",
            "walk": "marcher", "run": "courir", "eat": "manger", "drink": "boire",
            "sleep": "dormir", "wake": "se réveiller", "start": "commencer", "stop": "arrêter",
            "buy": "acheter", "sell": "vendre", "give": "donner", "take": "prendre",
            "make": "faire", "want": "vouloir", "need": "avoir besoin", "know": "savoir",
            "think": "penser", "believe": "croire", "understand": "comprendre", "remember": "se souvenir",
            "forget": "oublier", "live": "vivre", "die": "mourir", "grow": "grandir",
            "change": "changer", "stay": "rester", "leave": "partir", "arrive": "arriver",
            "time": "temps", "money": "argent", "people": "personnes", "world": "monde",
            "life": "vie", "way": "chemin", "thing": "chose", "place": "endroit",
            "year": "an", "week": "semaine", "month": "mois", "today": "aujourd'hui",
            "tomorrow": "demain", "yesterday": "hier", "now": "maintenant", "later": "plus tard",
            "always": "toujours", "never": "jamais", "sometimes": "parfois", "often": "souvent",
            "here": "ici", "there": "là", "where": "où", "what": "quoi",
            "who": "qui", "why": "pourquoi", "how": "comment", "when": "quand"
        },
        "de": {
            "hello": "hallo", "goodbye": "auf wiedersehen", "yes": "ja", "no": "nein", "please": "bitte",
            "thank you": "danke", "welcome": "willkommen", "good": "gut", "bad": "schlecht",
            "day": "tag", "night": "nacht", "morning": "morgen", "evening": "abend",
            "water": "wasser", "food": "essen", "house": "haus", "car": "auto", "book": "buch",
            "computer": "computer", "phone": "telefon", "work": "arbeit", "home": "zuhause",
            "family": "familie", "friend": "freund", "love": "liebe", "happy": "glücklich",
            "sad": "traurig", "big": "groß", "small": "klein", "new": "neu",
            "old": "alt", "hot": "heiß", "cold": "kalt", "open": "offen",
            "closed": "geschlossen", "help": "hilfe", "save": "speichern", "delete": "löschen",
            "create": "erstellen", "update": "aktualisieren", "search": "suchen", "find": "finden",
            "learn": "lernen", "teach": "lehren", "write": "schreiben", "read": "lesen",
            "speak": "sprechen", "listen": "hören", "see": "sehen", "hear": "hören",
            "walk": "gehen", "run": "laufen", "eat": "essen", "drink": "trinken",
            "sleep": "schlafen", "wake": "aufwachen", "start": "starten", "stop": "stoppen",
            "buy": "kaufen", "sell": "verkaufen", "give": "geben", "take": "nehmen",
            "make": "machen", "want": "wollen", "need": "brauchen", "know": "wissen",
            "think": "denken", "believe": "glauben", "understand": "verstehen", "remember": "erinnern",
            "forget": "vergessen", "live": "leben", "die": "sterben", "grow": "wachsen",
            "change": "ändern", "stay": "bleiben", "leave": "gehen", "arrive": "ankommen",
            "time": "zeit", "money": "geld", "people": "menschen", "world": "welt",
            "life": "leben", "way": "weg", "thing": "ding", "place": "ort",
            "year": "jahr", "week": "woche", "month": "monat", "today": "heute",
            "tomorrow": "morgen", "yesterday": "gestern", "now": "jetzt", "later": "später",
            "always": "immer", "never": "nie", "sometimes": "manchmal", "often": "oft",
            "here": "hier", "there": "dort", "where": "wo", "what": "was",
            "who": "wer", "why": "warum", "how": "wie", "when": "wann"
        },
        "it": {
            "hello": "ciao", "goodbye": "arrivederci", "yes": "sì", "no": "no", "please": "per favore",
            "thank you": "grazie", "welcome": "benvenuto", "good": "buono", "bad": "cattivo",
            "day": "giorno", "night": "notte", "morning": "mattina", "evening": "sera",
            "water": "acqua", "food": "cibo", "house": "casa", "car": "macchina", "book": "libro",
            "computer": "computer", "phone": "telefono", "work": "lavoro", "home": "casa",
            "family": "famiglia", "friend": "amico", "love": "amore", "happy": "felice",
            "sad": "triste", "big": "grande", "small": "piccolo", "new": "nuovo",
            "old": "vecchio", "hot": "caldo", "cold": "freddo", "open": "aperto",
            "closed": "chiuso", "help": "aiuto", "save": "salvare", "delete": "eliminare",
            "create": "creare", "update": "aggiornare", "search": "cercare", "find": "trovare",
            "learn": "imparare", "teach": "insegnare", "write": "scrivere", "read": "leggere",
            "speak": "parlare", "listen": "ascoltare", "see": "vedere", "hear": "sentire",
            "walk": "camminare", "run": "correre", "eat": "mangiare", "drink": "bere",
            "sleep": "dormire", "wake": "svegliarsi", "start": "iniziare", "stop": "fermare",
            "buy": "comprare", "sell": "vendere", "give": "dare", "take": "prendere",
            "make": "fare", "want": "volere", "need": "avere bisogno", "know": "sapere",
            "think": "pensare", "believe": "credere", "understand": "capire", "remember": "ricordare",
            "forget": "dimenticare", "live": "vivere", "die": "morire", "grow": "crescere",
            "change": "cambiare", "stay": "restare", "leave": "partire", "arrive": "arrivare",
            "time": "tempo", "money": "soldi", "people": "persone", "world": "mondo",
            "life": "vita", "way": "via", "thing": "cosa", "place": "luogo",
            "year": "anno", "week": "settimana", "month": "mese", "today": "oggi",
            "tomorrow": "domani", "yesterday": "ieri", "now": "adesso", "later": "dopo",
            "always": "sempre", "never": "mai", "sometimes": "a volte", "often": "spesso",
            "here": "qui", "there": "lì", "where": "dove", "what": "cosa",
            "who": "chi", "why": "perché", "how": "come", "when": "quando"
        }
    }
}

for source_dict in TRANSLATION_DICT.values():
    for target_dict in source_dict.values():
        reverse_dict = {}
        for en_word, translated in target_dict.items():
            reverse_dict[translated] = en_word
        target_dict.update(reverse_dict)

for source, targets in list(TRANSLATION_DICT.items()):
    for target, words in list(targets.items()):
        if source != target:
            TRANSLATION_DICT[target][source] = {v: k for k, v in words.items()}


def handler(event):
    try:
        text = event.get("text", "")
        source_lang = event.get("source_lang", "en")
        target_lang = event.get("target_lang", "es")

        if not text:
            return {"ok": False, "error": "text is required"}
        if source_lang not in ["en", "es", "fr", "de", "it"]:
            return {"ok": False, "error": "source_lang must be en, es, fr, de, or it"}
        if target_lang not in ["en", "es", "fr", "de", "it"]:
            return {"ok": False, "error": "target_lang must be en, es, fr, de, or it"}
        if source_lang == target_lang:
            return {"ok": False, "error": "source_lang and target_lang must be different"}

        words = text.lower().split()
        translated_words = []
        translated_count = 0

        if source_lang in TRANSLATION_DICT and target_lang in TRANSLATION_DICT[source_lang]:
            dictionary = TRANSLATION_DICT[source_lang][target_lang]
            for word in words:
                clean_word = word.strip('.,!?;:"\'-')
                punct = word[len(clean_word):] if len(word) > len(clean_word) else ""
                if clean_word in dictionary:
                    translated_words.append(dictionary[clean_word] + punct)
                    translated_count += 1
                else:
                    translated_words.append(word)
        else:
            translated_words = words

        translated_text = " ".join(translated_words)
        translated_text = translated_text.capitalize()

        confidence = round(translated_count / max(len(words), 1), 2)

        return {
            "ok": True,
            "translated_text": translated_text,
            "source_lang": source_lang,
            "target_lang": target_lang,
            "confidence": confidence
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
