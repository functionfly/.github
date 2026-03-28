import re

# Rule-based POS tagger
DETERMINERS = {"the","a","an","this","that","these","those","my","your","his","her","its","our","their","some","any","each","every","all","both","few","many","much","several","no","another","other","what","which","whose"}
PREPOSITIONS = {"in","on","at","to","for","of","with","by","from","as","into","through","during","before","after","above","below","between","out","off","over","under","again","further","then","once","about","against","along","among","around","behind","beside","besides","beyond","despite","down","except","inside","near","outside","past","since","throughout","toward","towards","under","underneath","unlike","until","up","upon","within","without"}
CONJUNCTIONS = {"and","but","or","nor","for","yet","so","although","because","since","unless","until","when","where","while","after","before","if","though","as","than","that","whether","both","either","neither","not only","but also","however","therefore","moreover","furthermore","nevertheless","nonetheless","meanwhile","otherwise","consequently","accordingly","hence","thus","thereby","whereby","wherefore","whereupon","whereas","whereby","wherein","whereof","whereto","wherewith","wherewithal","whereupon","wherever","whenever","whatever","whoever","whichever","however","whatever","whenever","wherever","whoever","whichever"}
PRONOUNS = {"i","me","my","myself","you","your","yourself","he","him","his","himself","she","her","hers","herself","it","its","itself","we","us","our","ourselves","they","them","their","theirs","themselves","who","whom","whose","which","what","this","that","these","those"}
AUXILIARIES = {"am","is","are","was","were","be","been","being","have","has","had","do","does","did","will","would","could","should","may","might","shall","can","need","dare","ought","used","must","let","make","get","keep","seem","appear","become","remain","stay","turn","grow","go","come","run","fall","feel","look","sound","smell","taste","prove","turn out"}
ADVERBS_SUFFIX = ("ly", "ward", "wards", "wise")
NOUN_SUFFIX = ("tion", "sion", "ness", "ment", "ity", "ty", "ance", "ence", "er", "or", "ist", "ism", "age", "ure", "ture", "ure", "ship", "hood", "dom", "ry", "ery", "ary", "ory")
VERB_SUFFIX = ("ize", "ise", "ify", "ate", "en", "ing", "ed", "es", "s")
ADJ_SUFFIX = ("ful", "less", "ous", "ious", "eous", "ive", "ative", "itive", "ic", "ical", "al", "ial", "ual", "able", "ible", "ible", "ent", "ant", "ish", "like", "ly", "ward", "some", "ary", "ory")

COMMON_NOUNS = {"time","year","people","way","day","man","woman","child","world","life","hand","part","place","case","week","company","system","program","question","government","number","night","point","home","water","room","mother","area","money","story","fact","month","lot","right","study","book","eye","job","word","business","issue","side","kind","head","house","service","friend","father","power","hour","game","line","end","among","while","name","land","different","home","move","try","kind","hand","picture","again","change","off","play","spell","air","away","animal","house","point","page","letter","mother","answer","found","study","still","learn","should","America","world"}


def _tag_word(word, prev_tag=None, next_word=None):
    w = word.lower()
    # Check specific word lists
    if w in DETERMINERS:
        return "DT"
    if w in PREPOSITIONS:
        return "IN"
    if w in CONJUNCTIONS:
        return "CC"
    if w in PRONOUNS:
        return "PRP"
    if w in AUXILIARIES:
        return "MD" if w in {"will","would","could","should","may","might","shall","can","must","need","dare","ought"} else "VBZ"
    # Punctuation
    if re.match(r'^[.!?]$', word):
        return "."
    if re.match(r'^[,;:]$', word):
        return ","
    if re.match(r'^\d+$', word):
        return "CD"
    if re.match(r'^[A-Z]', word) and prev_tag in (None, ".", ","):
        # Capitalized at start of sentence - could be proper noun
        if len(word) > 1 and word[1:].islower():
            return "NNP"
    # Suffix-based rules
    if w.endswith(ADVERBS_SUFFIX) and len(w) > 4:
        return "RB"
    if w.endswith(ADJ_SUFFIX) and len(w) > 4:
        return "JJ"
    if w.endswith(NOUN_SUFFIX) and len(w) > 4:
        return "NN"
    if w.endswith(("ing",)) and len(w) > 4:
        return "VBG"
    if w.endswith(("ed",)) and len(w) > 3:
        return "VBD"
    if w.endswith(("s",)) and len(w) > 3 and not w.endswith("ss"):
        return "NNS" if prev_tag in ("DT", "JJ", None) else "VBZ"
    if w in COMMON_NOUNS:
        return "NN"
    # Default: noun if follows determiner, else verb
    if prev_tag in ("DT", "JJ", "NN"):
        return "NN"
    return "VB"


TAG_DESCRIPTIONS = {
    "NN": "Noun, singular", "NNS": "Noun, plural", "NNP": "Proper noun, singular",
    "NNPS": "Proper noun, plural", "VB": "Verb, base form", "VBD": "Verb, past tense",
    "VBG": "Verb, gerund", "VBN": "Verb, past participle", "VBP": "Verb, present",
    "VBZ": "Verb, 3rd person singular", "JJ": "Adjective", "JJR": "Adjective, comparative",
    "JJS": "Adjective, superlative", "RB": "Adverb", "RBR": "Adverb, comparative",
    "RBS": "Adverb, superlative", "DT": "Determiner", "IN": "Preposition/conjunction",
    "CC": "Coordinating conjunction", "PRP": "Personal pronoun", "PRP$": "Possessive pronoun",
    "MD": "Modal", "CD": "Cardinal number", "EX": "Existential there",
    "FW": "Foreign word", "LS": "List item marker", "PDT": "Predeterminer",
    "POS": "Possessive ending", "RP": "Particle", "SYM": "Symbol",
    "TO": "to", "UH": "Interjection", "WDT": "Wh-determiner",
    "WP": "Wh-pronoun", "WP$": "Possessive wh-pronoun", "WRB": "Wh-adverb",
    ".": "Sentence-final punctuation", ",": "Comma/semicolon/colon"
}


def handler(event):
    text = event.get("text") if isinstance(event, dict) else None
    if not text or not isinstance(text, str):
        return {"ok": False, "error": "text (string) is required"}
    try:
        words = re.findall(r'\b\w+\b|[^\w\s]', text)
        tagged = []
        prev_tag = None
        for i, word in enumerate(words):
            next_word = words[i + 1] if i + 1 < len(words) else None
            tag = _tag_word(word, prev_tag, next_word)
            tagged.append({
                "word": word,
                "tag": tag,
                "description": TAG_DESCRIPTIONS.get(tag, "Unknown")
            })
            prev_tag = tag
        return {
            "ok": True,
            "result": tagged,
            "tagged": tagged,
            "count": len(tagged)
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
