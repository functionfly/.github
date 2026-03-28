import re

MALE_PRONOUNS = {"he", "him", "his", "himself"}
FEMALE_PRONOUNS = {"she", "her", "hers", "herself"}
NEUTRAL_PRONOUNS = {"they", "them", "their", "theirs", "themselves", "it", "its", "itself"}
ALL_PRONOUNS = MALE_PRONOUNS | FEMALE_PRONOUNS | NEUTRAL_PRONOUNS


def _split_sentences(text):
    return re.split(r'(?<=[.!?])\s+', text.strip())


def _find_noun_phrases(sentence):
    """Find capitalized noun phrases (potential antecedents)."""
    return re.findall(r'\b[A-Z][a-z]+(?:\s+[A-Z][a-z]+)*\b', sentence)


def handler(event):
    text = event.get("text") if isinstance(event, dict) else None
    if not text or not isinstance(text, str):
        return {"ok": False, "error": "text (string) is required"}
    try:
        sentences = _split_sentences(text)
        chains = []
        antecedent_stack = []  # (entity, gender_hint)
        resolved_sentences = []

        for sent_idx, sentence in enumerate(sentences):
            words = sentence.split()
            noun_phrases = _find_noun_phrases(sentence)
            # Add new noun phrases to antecedent stack
            for np in noun_phrases:
                antecedent_stack.append(np)

            # Find pronouns and try to resolve
            resolved = sentence
            for word in words:
                w_lower = word.lower().strip('.,!?;:')
                if w_lower in ALL_PRONOUNS and antecedent_stack:
                    # Use most recent antecedent
                    antecedent = antecedent_stack[-1]
                    chains.append({
                        "pronoun": w_lower,
                        "antecedent": antecedent,
                        "sentence_index": sent_idx
                    })
            resolved_sentences.append(resolved)

        # Build coreference chains grouped by antecedent
        chain_map = {}
        for item in chains:
            ant = item["antecedent"]
            if ant not in chain_map:
                chain_map[ant] = []
            chain_map[ant].append(item["pronoun"])

        coref_chains = [
            {"antecedent": ant, "pronouns": list(set(pronouns)), "mentions": len(pronouns)}
            for ant, pronouns in chain_map.items()
        ]

        return {
            "ok": True,
            "result": coref_chains,
            "chains": coref_chains,
            "resolved_text": " ".join(resolved_sentences),
            "total_pronouns": len(chains),
            "note": "Rule-based heuristic coreference — for production use, integrate neuralcoref or similar"
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
